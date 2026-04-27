package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"survey-platform/internal/cache"
	"survey-platform/internal/database"
	"survey-platform/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SurveyHandler struct{}

func NewSurveyHandler() *SurveyHandler {
	return &SurveyHandler{}
}

type CreateSurveyRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateSurveyRequest struct {
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Status         string        `json:"status"`
	StartTime      *time.Time    `json:"start_time"`
	EndTime        *time.Time    `json:"end_time"`
	MaxResponses   int           `json:"max_responses"`
	RequireLogin   bool          `json:"require_login"`
	AllowDuplicate bool          `json:"allow_duplicate"`
}

func (h *SurveyHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	rows, err := database.DB.Query(context.Background(),
		`SELECT id, user_id, title, description, status, start_time, end_time, 
		        max_responses, require_login, allow_duplicate, created_at, updated_at
		 FROM surveys 
		 WHERE user_id = $1 
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var surveys []*models.Survey
	for rows.Next() {
		var s models.Survey
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Title, &s.Description, &s.Status,
			&s.StartTime, &s.EndTime, &s.MaxResponses, &s.RequireLogin,
			&s.AllowDuplicate, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		surveys = append(surveys, &s)
	}

	c.JSON(http.StatusOK, surveys)
}

func (h *SurveyHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req CreateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var survey models.Survey
	err := database.DB.QueryRow(context.Background(),
		`INSERT INTO surveys (user_id, title, status) 
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, title, description, status, start_time, end_time, 
		           max_responses, require_login, allow_duplicate, created_at, updated_at`,
		userID, req.Title, models.SurveyStatusDraft,
	).Scan(
		&survey.ID, &survey.UserID, &survey.Title, &survey.Description, &survey.Status,
		&survey.StartTime, &survey.EndTime, &survey.MaxResponses, &survey.RequireLogin,
		&survey.AllowDuplicate, &survey.CreatedAt, &survey.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, survey)
}

func (h *SurveyHandler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var survey models.Survey
	err = database.DB.QueryRow(context.Background(),
		`SELECT id, user_id, title, description, status, start_time, end_time, 
		        max_responses, require_login, allow_duplicate, created_at, updated_at
		 FROM surveys 
		 WHERE id = $1 AND user_id = $2`,
		surveyID, userID,
	).Scan(
		&survey.ID, &survey.UserID, &survey.Title, &survey.Description, &survey.Status,
		&survey.StartTime, &survey.EndTime, &survey.MaxResponses, &survey.RequireLogin,
		&survey.AllowDuplicate, &survey.CreatedAt, &survey.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	questions, err := loadQuestions(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	survey.Questions = questions

	logicRules, err := loadLogicRules(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	survey.LogicRules = logicRules

	c.JSON(http.StatusOK, survey)
}

func (h *SurveyHandler) GetForFill(c *gin.Context) {
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	cacheKey := fmt.Sprintf("survey:structure:%s", surveyID)
	var survey models.Survey
	if err := cache.Get(context.Background(), cacheKey, &survey); err == nil {
		c.JSON(http.StatusOK, survey)
		return
	}

	var surveyData models.Survey
	err = database.DB.QueryRow(context.Background(),
		`SELECT id, user_id, title, description, status, start_time, end_time, 
		        max_responses, require_login, allow_duplicate, created_at, updated_at
		 FROM surveys 
		 WHERE id = $1`,
		surveyID,
	).Scan(
		&surveyData.ID, &surveyData.UserID, &surveyData.Title, &surveyData.Description, &surveyData.Status,
		&surveyData.StartTime, &surveyData.EndTime, &surveyData.MaxResponses, &surveyData.RequireLogin,
		&surveyData.AllowDuplicate, &surveyData.CreatedAt, &surveyData.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	if surveyData.Status != models.SurveyStatusActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "survey is not active"})
		return
	}

	now := time.Now()
	if surveyData.StartTime != nil && now.Before(*surveyData.StartTime) {
		c.JSON(http.StatusForbidden, gin.H{"error": "survey has not started yet"})
		return
	}
	if surveyData.EndTime != nil && now.After(*surveyData.EndTime) {
		c.JSON(http.StatusForbidden, gin.H{"error": "survey has ended"})
		return
	}

	questions, err := loadQuestions(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	surveyData.Questions = questions

	logicRules, err := loadLogicRules(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	surveyData.LogicRules = logicRules

	cache.Set(context.Background(), cacheKey, surveyData, 10*time.Minute)

	c.JSON(http.StatusOK, surveyData)
}

func (h *SurveyHandler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var req UpdateSurveyRequest
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

	var survey models.Survey
	err = database.DB.QueryRow(context.Background(),
		`UPDATE surveys SET 
			title = COALESCE(NULLIF($1, ''), title),
			description = COALESCE($2, description),
			status = COALESCE(NULLIF($3, ''), status),
			start_time = $4,
			end_time = $5,
			max_responses = COALESCE(NULLIF($6, 0), max_responses),
			require_login = COALESCE($7, require_login),
			allow_duplicate = COALESCE($8, allow_duplicate),
			updated_at = CURRENT_TIMESTAMP
		 WHERE id = $9 AND user_id = $10
		 RETURNING id, user_id, title, description, status, start_time, end_time, 
		           max_responses, require_login, allow_duplicate, created_at, updated_at`,
		req.Title, req.Description, req.Status, req.StartTime, req.EndTime,
		req.MaxResponses, req.RequireLogin, req.AllowDuplicate, surveyID, userID,
	).Scan(
		&survey.ID, &survey.UserID, &survey.Title, &survey.Description, &survey.Status,
		&survey.StartTime, &survey.EndTime, &survey.MaxResponses, &survey.RequireLogin,
		&survey.AllowDuplicate, &survey.CreatedAt, &survey.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))
	cache.Delete(context.Background(), fmt.Sprintf("survey:stats:%s", surveyID))

	c.JSON(http.StatusOK, survey)
}

func (h *SurveyHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	result, err := database.DB.Exec(context.Background(),
		"DELETE FROM surveys WHERE id = $1 AND user_id = $2",
		surveyID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))
	cache.Delete(context.Background(), fmt.Sprintf("survey:stats:%s", surveyID))

	c.JSON(http.StatusOK, gin.H{"message": "survey deleted"})
}

func loadQuestions(ctx context.Context, surveyID uuid.UUID) ([]*models.Question, error) {
	rows, err := database.DB.Query(ctx,
		`SELECT id, survey_id, question_order, question_type, title, description, 
		        is_required, shuffle_options, options, matrix_rows, matrix_cols, 
		        min_rating, max_rating, created_at, updated_at
		 FROM questions 
		 WHERE survey_id = $1 
		 ORDER BY question_order`,
		surveyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		var q models.Question
		var optionsJSON, matrixRowsJSON, matrixColsJSON json.RawMessage
		err := rows.Scan(
			&q.ID, &q.SurveyID, &q.QuestionOrder, &q.QuestionType, &q.Title, &q.Description,
			&q.IsRequired, &q.ShuffleOptions, &optionsJSON, &matrixRowsJSON, &matrixColsJSON,
			&q.MinRating, &q.MaxRating, &q.CreatedAt, &q.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := q.OptionsFromJSON(optionsJSON); err != nil {
			return nil, err
		}
		if err := q.MatrixRowsFromJSON(matrixRowsJSON); err != nil {
			return nil, err
		}
		if err := q.MatrixColsFromJSON(matrixColsJSON); err != nil {
			return nil, err
		}

		questions = append(questions, &q)
	}

	return questions, nil
}

func loadLogicRules(ctx context.Context, surveyID uuid.UUID) ([]*models.LogicRule, error) {
	rows, err := database.DB.Query(ctx,
		`SELECT id, survey_id, trigger_question_id, trigger_option_value, 
		        action_type, target_question_id, rule_order, created_at
		 FROM logic_rules 
		 WHERE survey_id = $1 
		 ORDER BY rule_order`,
		surveyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.LogicRule
	for rows.Next() {
		var r models.LogicRule
		err := rows.Scan(
			&r.ID, &r.SurveyID, &r.TriggerQuestionID, &r.TriggerOptionValue,
			&r.ActionType, &r.TargetQuestionID, &r.RuleOrder, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &r)
	}

	return rules, nil
}
