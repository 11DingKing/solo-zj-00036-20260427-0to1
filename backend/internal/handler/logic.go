package handler

import (
	"context"
	"fmt"
	"net/http"
	"survey-platform/internal/cache"
	"survey-platform/internal/database"
	"survey-platform/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LogicHandler struct{}

func NewLogicHandler() *LogicHandler {
	return &LogicHandler{}
}

type CreateLogicRuleRequest struct {
	TriggerQuestionID  uuid.UUID `json:"trigger_question_id" binding:"required"`
	TriggerOptionValue string    `json:"trigger_option_value" binding:"required"`
	ActionType         string    `json:"action_type" binding:"required"`
	TargetQuestionID   uuid.UUID `json:"target_question_id" binding:"required"`
}

func (h *LogicHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var exists bool
	err = database.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM surveys WHERE id = $1 AND user_id = $2)",
		surveyID, userID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	rules, err := loadLogicRules(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

func (h *LogicHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var req CreateLogicRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	err = database.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM surveys WHERE id = $1 AND user_id = $2)",
		surveyID, userID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	var maxOrder int
	database.DB.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(rule_order), 0) FROM logic_rules WHERE survey_id = $1",
		surveyID,
	).Scan(&maxOrder)
	newOrder := maxOrder + 1

	var rule models.LogicRule
	err = database.DB.QueryRow(context.Background(),
		`INSERT INTO logic_rules (
			survey_id, trigger_question_id, trigger_option_value, 
			action_type, target_question_id, rule_order
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, survey_id, trigger_question_id, trigger_option_value, 
		          action_type, target_question_id, rule_order, created_at`,
		surveyID, req.TriggerQuestionID, req.TriggerOptionValue,
		req.ActionType, req.TargetQuestionID, newOrder,
	).Scan(
		&rule.ID, &rule.SurveyID, &rule.TriggerQuestionID, &rule.TriggerOptionValue,
		&rule.ActionType, &rule.TargetQuestionID, &rule.RuleOrder, &rule.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))

	c.JSON(http.StatusCreated, rule)
}

func (h *LogicHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	ruleID, err := uuid.Parse(c.Param("logic_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	var surveyID uuid.UUID
	err = database.DB.QueryRow(context.Background(),
		`SELECT lr.survey_id FROM logic_rules lr
		 JOIN surveys s ON lr.survey_id = s.id
		 WHERE lr.id = $1 AND s.user_id = $2`,
		ruleID, userID,
	).Scan(&surveyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "logic rule not found"})
		return
	}

	result, err := database.DB.Exec(context.Background(),
		"DELETE FROM logic_rules WHERE id = $1", ruleID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "logic rule not found"})
		return
	}

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))

	c.JSON(http.StatusOK, gin.H{"message": "logic rule deleted"})
}

func (h *LogicHandler) UpdateOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var orderMap []struct {
		RuleID    uuid.UUID `json:"rule_id" binding:"required"`
		RuleOrder int       `json:"rule_order" binding:"required"`
	}

	if err := c.ShouldBindJSON(&orderMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	err = database.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM surveys WHERE id = $1 AND user_id = $2)",
		surveyID, userID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(context.Background())

	for _, item := range orderMap {
		_, err := tx.Exec(context.Background(),
			"UPDATE logic_rules SET rule_order = $1 WHERE id = $2 AND survey_id = $3",
			item.RuleOrder, item.RuleID, surveyID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))

	c.JSON(http.StatusOK, gin.H{"message": "rule order updated"})
}
