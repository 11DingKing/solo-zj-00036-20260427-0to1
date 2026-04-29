package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SurveyStatus string

const (
	SurveyStatusDraft   SurveyStatus = "draft"
	SurveyStatusActive  SurveyStatus = "active"
	SurveyStatusPaused  SurveyStatus = "paused"
	SurveyStatusClosed  SurveyStatus = "closed"
)

type Survey struct {
	ID             uuid.UUID     `json:"id"`
	UserID         uuid.UUID     `json:"user_id"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Status         SurveyStatus  `json:"status"`
	StartTime      *time.Time    `json:"start_time"`
	EndTime        *time.Time    `json:"end_time"`
	MaxResponses   int           `json:"max_responses"`
	RequireLogin   bool          `json:"require_login"`
	AllowDuplicate bool          `json:"allow_duplicate"`
	Questions      []*Question   `json:"questions,omitempty"`
	LogicRules     []*LogicRule  `json:"logic_rules,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type QuestionType string

const (
	QuestionTypeSingleChoice QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeDropdown      QuestionType = "dropdown"
	QuestionTypeText          QuestionType = "text"
	QuestionTypeTextarea      QuestionType = "textarea"
	QuestionTypeRating        QuestionType = "rating"
	QuestionTypeMatrix        QuestionType = "matrix"
	QuestionTypeSorting       QuestionType = "sorting"
	QuestionTypeDate          QuestionType = "date"
)

type Option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type MatrixRow struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type MatrixCol struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type Question struct {
	ID             uuid.UUID    `json:"id"`
	SurveyID       uuid.UUID    `json:"survey_id"`
	QuestionOrder  int          `json:"question_order"`
	QuestionType   QuestionType `json:"question_type"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	IsRequired     bool         `json:"is_required"`
	ShuffleOptions bool         `json:"shuffle_options"`
	Options        []*Option    `json:"options"`
	MatrixRows     []*MatrixRow `json:"matrix_rows"`
	MatrixCols     []*MatrixCol `json:"matrix_cols"`
	MinRating      int          `json:"min_rating"`
	MaxRating      int          `json:"max_rating"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

func (q *Question) OptionsFromJSON(data json.RawMessage) error {
	if len(data) == 0 {
		q.Options = []*Option{}
		return nil
	}
	return json.Unmarshal(data, &q.Options)
}

func (q *Question) MatrixRowsFromJSON(data json.RawMessage) error {
	if len(data) == 0 {
		q.MatrixRows = []*MatrixRow{}
		return nil
	}
	return json.Unmarshal(data, &q.MatrixRows)
}

func (q *Question) MatrixColsFromJSON(data json.RawMessage) error {
	if len(data) == 0 {
		q.MatrixCols = []*MatrixCol{}
		return nil
	}
	return json.Unmarshal(data, &q.MatrixCols)
}

type LogicAction string

const (
	LogicActionJump LogicAction = "jump"
)

type LogicRule struct {
	ID                 uuid.UUID     `json:"id"`
	SurveyID           uuid.UUID     `json:"survey_id"`
	TriggerQuestionID  uuid.UUID     `json:"trigger_question_id"`
	TriggerOptionValue string        `json:"trigger_option_value"`
	ActionType         LogicAction   `json:"action_type"`
	TargetQuestionID   uuid.UUID     `json:"target_question_id"`
	RuleOrder          int           `json:"rule_order"`
	CreatedAt          time.Time     `json:"created_at"`
}

type Response struct {
	ID              uuid.UUID  `json:"id"`
	SurveyID        uuid.UUID  `json:"survey_id"`
	UserID          *uuid.UUID `json:"user_id"`
	SubmitterIP     string     `json:"submitter_ip"`
	StartTime       time.Time  `json:"start_time"`
	SubmitTime      *time.Time `json:"submit_time"`
	IsCompleted     bool       `json:"is_completed"`
	DurationSeconds int        `json:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ResponseAnswer struct {
	ID          uuid.UUID       `json:"id"`
	ResponseID  uuid.UUID       `json:"response_id"`
	QuestionID  uuid.UUID       `json:"question_id"`
	AnswerValue string          `json:"answer_value"`
	AnswerJSON  json.RawMessage `json:"answer_json"`
	CreatedAt   time.Time       `json:"created_at"`
}

type SurveyStats struct {
	TotalResponses   int     `json:"total_responses"`
	CompletedCount   int     `json:"completed_count"`
	CompletionRate   float64 `json:"completion_rate"`
	AverageDuration  float64 `json:"average_duration"`
}

type QuestionStats struct {
	QuestionID uuid.UUID `json:"question_id"`
	Type       string    `json:"type"`
	
	// 单选题/多选题/下拉题
	OptionCounts map[string]int `json:"option_counts,omitempty"`
	Percentages  map[string]float64 `json:"percentages,omitempty"`
	
	// 评分题
	AverageRating float64 `json:"average_rating,omitempty"`
	RatingCounts  map[int]int `json:"rating_counts,omitempty"`
	
	// 填空题/多行文本
	TextAnswers []string `json:"text_answers,omitempty"`
	WordCloud   map[string]int `json:"word_cloud,omitempty"`
	
	// 矩阵题
	MatrixStats map[string]map[string]int `json:"matrix_stats,omitempty"`
	
	// 排序题
	SortPatterns map[string]int `json:"sort_patterns,omitempty"`
	
	// 日期题
	DateCounts map[string]int `json:"date_counts,omitempty"`
}

type CrosstabData struct {
	RowQuestionID    uuid.UUID              `json:"row_question_id"`
	RowQuestionTitle string                 `json:"row_question_title"`
	RowLabels        []string               `json:"row_labels"`
	
	ColQuestionID    uuid.UUID              `json:"col_question_id"`
	ColQuestionTitle string                 `json:"col_question_title"`
	ColLabels        []string               `json:"col_labels"`
	
	Counts           [][]int                `json:"counts"`
	Percentages      [][]float64            `json:"percentages"`
}
