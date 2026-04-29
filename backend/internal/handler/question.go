package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"survey-platform/internal/cache"
	"survey-platform/internal/database"
	"survey-platform/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuestionHandler struct{}

func NewQuestionHandler() *QuestionHandler {
	return &QuestionHandler{}
}

type CreateQuestionRequest struct {
	QuestionType   string             `json:"question_type" binding:"required"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	IsRequired     bool               `json:"is_required"`
	ShuffleOptions bool               `json:"shuffle_options"`
	QuestionOrder  int                `json:"question_order"`
	Options        []*models.Option   `json:"options"`
	MatrixRows     []*models.MatrixRow `json:"matrix_rows"`
	MatrixCols     []*models.MatrixCol `json:"matrix_cols"`
	MinRating      int                `json:"min_rating"`
	MaxRating      int                `json:"max_rating"`
}

type UpdateQuestionRequest struct {
	QuestionOrder  int                `json:"question_order"`
	QuestionType   string             `json:"question_type"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	IsRequired     *bool              `json:"is_required"`
	ShuffleOptions *bool              `json:"shuffle_options"`
	Options        []*models.Option   `json:"options"`
	MatrixRows     []*models.MatrixRow `json:"matrix_rows"`
	MatrixCols     []*models.MatrixCol `json:"matrix_cols"`
	MinRating      int                `json:"min_rating"`
	MaxRating      int                `json:"max_rating"`
}

func (h *QuestionHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var req CreateQuestionRequest
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

	var newOrder int
	if req.QuestionOrder > 0 {
		newOrder = req.QuestionOrder
	} else {
		var maxOrder int
		database.DB.QueryRow(context.Background(),
			"SELECT COALESCE(MAX(question_order), 0) FROM questions WHERE survey_id = $1",
			surveyID,
		).Scan(&maxOrder)
		newOrder = maxOrder + 1
	}

	optionsJSON, _ := json.Marshal(req.Options)
	matrixRowsJSON, _ := json.Marshal(req.MatrixRows)
	matrixColsJSON, _ := json.Marshal(req.MatrixCols)

	var question models.Question
	err = database.DB.QueryRow(context.Background(),
		`INSERT INTO questions (
			survey_id, question_order, question_type, title, description, 
			is_required, shuffle_options, options, matrix_rows, matrix_cols, 
			min_rating, max_rating
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, survey_id, question_order, question_type, title, description, 
		          is_required, shuffle_options, options, matrix_rows, matrix_cols, 
		          min_rating, max_rating, created_at, updated_at`,
		surveyID, newOrder, req.QuestionType, req.Title, req.Description,
		req.IsRequired, req.ShuffleOptions, optionsJSON, matrixRowsJSON, matrixColsJSON,
		req.MinRating, req.MaxRating,
	).Scan(
		&question.ID, &question.SurveyID, &question.QuestionOrder, &question.QuestionType,
		&question.Title, &question.Description, &question.IsRequired, &question.ShuffleOptions,
		&optionsJSON, &matrixRowsJSON, &matrixColsJSON, &question.MinRating, &question.MaxRating,
		&question.CreatedAt, &question.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	question.Options = req.Options
	question.MatrixRows = req.MatrixRows
	question.MatrixCols = req.MatrixCols

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))
	cache.Delete(context.Background(), fmt.Sprintf("survey:stats:%s", surveyID))

	c.JSON(http.StatusCreated, question)
}

func (h *QuestionHandler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	questionID, err := uuid.Parse(c.Param("question_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	var req UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var surveyID uuid.UUID
	err = database.DB.QueryRow(context.Background(),
		`SELECT q.survey_id FROM questions q
		 JOIN surveys s ON q.survey_id = s.id
		 WHERE q.id = $1 AND s.user_id = $2`,
		questionID, userID,
	).Scan(&surveyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	optionsJSON, _ := json.Marshal(req.Options)
	matrixRowsJSON, _ := json.Marshal(req.MatrixRows)
	matrixColsJSON, _ := json.Marshal(req.MatrixCols)

	var question models.Question
	err = database.DB.QueryRow(context.Background(),
		`UPDATE questions SET 
			question_order = COALESCE(NULLIF($1, 0), question_order),
			question_type = COALESCE(NULLIF($2, ''), question_type),
			title = COALESCE(NULLIF($3, ''), title),
			description = COALESCE($4, description),
			is_required = COALESCE($5, is_required),
			shuffle_options = COALESCE($6, shuffle_options),
			options = $7,
			matrix_rows = $8,
			matrix_cols = $9,
			min_rating = COALESCE(NULLIF($10, 0), min_rating),
			max_rating = COALESCE(NULLIF($11, 0), max_rating),
			updated_at = CURRENT_TIMESTAMP
		 WHERE id = $12
		 RETURNING id, survey_id, question_order, question_type, title, description, 
		           is_required, shuffle_options, options, matrix_rows, matrix_cols, 
		           min_rating, max_rating, created_at, updated_at`,
		req.QuestionOrder, req.QuestionType, req.Title, req.Description,
		req.IsRequired, req.ShuffleOptions, optionsJSON, matrixRowsJSON, matrixColsJSON,
		req.MinRating, req.MaxRating, questionID,
	).Scan(
		&question.ID, &question.SurveyID, &question.QuestionOrder, &question.QuestionType,
		&question.Title, &question.Description, &question.IsRequired, &question.ShuffleOptions,
		&optionsJSON, &matrixRowsJSON, &matrixColsJSON, &question.MinRating, &question.MaxRating,
		&question.CreatedAt, &question.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	question.Options = req.Options
	question.MatrixRows = req.MatrixRows
	question.MatrixCols = req.MatrixCols

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))
	cache.Delete(context.Background(), fmt.Sprintf("survey:stats:%s", surveyID))

	c.JSON(http.StatusOK, question)
}

func (h *QuestionHandler) UpdateOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	var orderMap []struct {
		QuestionID   uuid.UUID `json:"question_id" binding:"required"`
		QuestionOrder int      `json:"question_order" binding:"required"`
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
			"UPDATE questions SET question_order = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND survey_id = $3",
			item.QuestionOrder, item.QuestionID, surveyID,
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

	c.JSON(http.StatusOK, gin.H{"message": "order updated"})
}

func (h *QuestionHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	questionID, err := uuid.Parse(c.Param("question_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	var surveyID uuid.UUID
	err = database.DB.QueryRow(context.Background(),
		`SELECT q.survey_id FROM questions q
		 JOIN surveys s ON q.survey_id = s.id
		 WHERE q.id = $1 AND s.user_id = $2`,
		questionID, userID,
	).Scan(&surveyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	result, err := database.DB.Exec(context.Background(),
		"DELETE FROM questions WHERE id = $1", questionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	cache.Delete(context.Background(), fmt.Sprintf("survey:structure:%s", surveyID))
	cache.Delete(context.Background(), fmt.Sprintf("survey:stats:%s", surveyID))

	c.JSON(http.StatusOK, gin.H{"message": "question deleted"})
}
