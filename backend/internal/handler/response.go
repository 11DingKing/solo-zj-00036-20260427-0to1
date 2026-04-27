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

type ResponseHandler struct{}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

type StartResponseRequest struct {
	SurveyID uuid.UUID `json:"survey_id" binding:"required"`
}

type AnswerRequest struct {
	QuestionID uuid.UUID       `json:"question_id" binding:"required"`
	AnswerValue string          `json:"answer_value"`
	AnswerJSON  json.RawMessage `json:"answer_json"`
}

type SubmitResponseRequest struct {
	ResponseID uuid.UUID        `json:"response_id" binding:"required"`
	Answers    []*AnswerRequest `json:"answers" binding:"required"`
}

func (h *ResponseHandler) Start(c *gin.Context) {
	var req StartResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var survey models.Survey
	err := database.DB.QueryRow(context.Background(),
		`SELECT id, status, start_time, end_time, max_responses, require_login, allow_duplicate
		 FROM surveys WHERE id = $1`,
		req.SurveyID,
	).Scan(
		&survey.ID, &survey.Status, &survey.StartTime, &survey.EndTime,
		&survey.MaxResponses, &survey.RequireLogin, &survey.AllowDuplicate,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
		return
	}

	if survey.Status != models.SurveyStatusActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "survey is not active"})
		return
	}

	now := time.Now()
	if survey.StartTime != nil && now.Before(*survey.StartTime) {
		c.JSON(http.StatusForbidden, gin.H{"error": "survey has not started yet"})
		return
	}
	if survey.EndTime != nil && now.After(*survey.EndTime) {
		c.JSON(http.StatusForbidden, gin.H{"error": "survey has ended"})
		return
	}

	if survey.MaxResponses > 0 {
		var count int
		database.DB.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM responses WHERE survey_id = $1 AND is_completed = true",
			survey.ID,
		).Scan(&count)
		if count >= survey.MaxResponses {
			c.JSON(http.StatusForbidden, gin.H{"error": "survey has reached max responses"})
			return
		}
	}

	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		uidVal := uid.(uuid.UUID)
		userID = &uidVal
	}

	if survey.RequireLogin && userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login required to take this survey"})
		return
	}

	if !survey.AllowDuplicate && userID != nil {
		var exists bool
		database.DB.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM responses WHERE survey_id = $1 AND user_id = $2 AND is_completed = true)",
			survey.ID, userID,
		).Scan(&exists)
		if exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "you have already completed this survey"})
			return
		}
	}

	clientIP := c.ClientIP()

	var response models.Response
	err = database.DB.QueryRow(context.Background(),
		`INSERT INTO responses (survey_id, user_id, submitter_ip, start_time, is_completed)
		 VALUES ($1, $2, $3, CURRENT_TIMESTAMP, false)
		 RETURNING id, survey_id, user_id, submitter_ip, start_time, is_completed, created_at`,
		survey.ID, userID, clientIP,
	).Scan(
		&response.ID, &response.SurveyID, &response.UserID, &response.SubmitterIP,
		&response.StartTime, &response.IsCompleted, &response.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"response_id": response.ID,
		"start_time":  response.StartTime,
	})
}

func (h *ResponseHandler) Submit(c *gin.Context) {
	var req SubmitResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response models.Response
	err := database.DB.QueryRow(context.Background(),
		`SELECT id, survey_id, start_time FROM responses WHERE id = $1 AND is_completed = false`,
		req.ResponseID,
	).Scan(&response.ID, &response.SurveyID, &response.StartTime)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "response not found or already submitted"})
		return
	}

	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(context.Background())

	submitTime := time.Now()
	durationSeconds := int(submitTime.Sub(response.StartTime).Seconds())

	_, err = tx.Exec(context.Background(),
		`UPDATE responses 
		 SET submit_time = $1, is_completed = true, duration_seconds = $2
		 WHERE id = $3`,
		submitTime, durationSeconds, response.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, answer := range req.Answers {
		var answerJSON json.RawMessage
		if len(answer.AnswerJSON) > 0 {
			answerJSON = answer.AnswerJSON
		}

		_, err = tx.Exec(context.Background(),
			`INSERT INTO response_answers (response_id, question_id, answer_value, answer_json)
			 VALUES ($1, $2, $3, $4)`,
			response.ID, answer.QuestionID, answer.AnswerValue, answerJSON,
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

	cache.Delete(context.Background(), fmt.Sprintf("survey:stats:%s", response.SurveyID))

	c.JSON(http.StatusOK, gin.H{
		"message":         "response submitted successfully",
		"duration_seconds": durationSeconds,
	})
}
