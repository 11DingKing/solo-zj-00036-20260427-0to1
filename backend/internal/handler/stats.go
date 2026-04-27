package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"survey-platform/internal/cache"
	"survey-platform/internal/database"
	"survey-platform/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type StatsHandler struct{}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

func (h *StatsHandler) GetSurveyStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("survey_id"))
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

	cacheKey := fmt.Sprintf("survey:stats:%s", surveyID)
	var stats models.SurveyStats
	if err := cache.Get(context.Background(), cacheKey, &stats); err == nil {
		c.JSON(http.StatusOK, stats)
		return
	}

	var totalResponses, completedCount int
	var totalDuration float64

	err = database.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM responses WHERE survey_id = $1`,
		surveyID,
	).Scan(&totalResponses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = database.DB.QueryRow(context.Background(),
		`SELECT COUNT(*), COALESCE(AVG(duration_seconds), 0) 
		 FROM responses WHERE survey_id = $1 AND is_completed = true`,
		surveyID,
	).Scan(&completedCount, &totalDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats.TotalResponses = totalResponses
	stats.CompletedCount = completedCount
	if totalResponses > 0 {
		stats.CompletionRate = float64(completedCount) / float64(totalResponses) * 100
	}
	stats.AverageDuration = totalDuration

	cache.Set(context.Background(), cacheKey, stats, 5*time.Minute)

	c.JSON(http.StatusOK, stats)
}

func (h *StatsHandler) GetQuestionStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("survey_id"))
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

	questions, err := loadQuestions(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []*models.QuestionStats
	for _, q := range questions {
		stats, err := calculateQuestionStats(q, surveyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result = append(result, stats)
	}

	c.JSON(http.StatusOK, result)
}

func calculateQuestionStats(q *models.Question, surveyID uuid.UUID) (*models.QuestionStats, error) {
	stats := &models.QuestionStats{
		QuestionID: q.ID,
		Type:       string(q.QuestionType),
	}

	switch q.QuestionType {
	case models.QuestionTypeSingleChoice, models.QuestionTypeDropdown:
		optionCounts := make(map[string]int)
		for _, opt := range q.Options {
			optionCounts[opt.Value] = 0
		}

		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_value, COUNT(*) 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true
			 GROUP BY answer_value`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		total := 0
		for rows.Next() {
			var val string
			var count int
			if err := rows.Scan(&val, &count); err != nil {
				return nil, err
			}
			optionCounts[val] = count
			total += count
		}

		stats.OptionCounts = optionCounts
		stats.Percentages = make(map[string]float64)
		if total > 0 {
			for k, v := range optionCounts {
				stats.Percentages[k] = float64(v) / float64(total) * 100
			}
		}

	case models.QuestionTypeMultipleChoice:
		optionCounts := make(map[string]int)
		for _, opt := range q.Options {
			optionCounts[opt.Value] = 0
		}

		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_json 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		total := 0
		for rows.Next() {
			var answerJSON json.RawMessage
			if err := rows.Scan(&answerJSON); err != nil {
				return nil, err
			}
			var values []string
			if err := json.Unmarshal(answerJSON, &values); err != nil {
				continue
			}
			total++
			for _, val := range values {
				optionCounts[val]++
			}
		}

		stats.OptionCounts = optionCounts
		stats.Percentages = make(map[string]float64)
		if total > 0 {
			for k, v := range optionCounts {
				stats.Percentages[k] = float64(v) / float64(total) * 100
			}
		}

	case models.QuestionTypeRating:
		ratingCounts := make(map[int]int)
		for i := q.MinRating; i <= q.MaxRating; i++ {
			ratingCounts[i] = 0
		}

		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_value 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		total := 0
		sum := 0
		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err != nil {
				return nil, err
			}
			rating, err := strconv.Atoi(val)
			if err != nil {
				continue
			}
			ratingCounts[rating]++
			total++
			sum += rating
		}

		stats.RatingCounts = ratingCounts
		if total > 0 {
			stats.AverageRating = float64(sum) / float64(total)
		}

	case models.QuestionTypeText, models.QuestionTypeTextarea:
		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_value 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true AND answer_value IS NOT NULL AND answer_value != ''
			 LIMIT 1000`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var textAnswers []string
		wordCloud := make(map[string]int)
		stopWords := map[string]bool{
			"the": true, "a": true, "an": true, "is": true, "are": true,
			"was": true, "were": true, "be": true, "been": true, "being": true,
			"to": true, "of": true, "in": true, "for": true, "on": true,
			"with": true, "at": true, "by": true, "和": true, "的": true, "了": true,
		}

		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err != nil {
				return nil, err
			}
			textAnswers = append(textAnswers, val)
			
			words := strings.FieldsFunc(val, func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r >= 0x4e00 && r <= 0x9fff))
			})
			for _, word := range words {
				word = strings.ToLower(word)
				if len(word) > 1 && !stopWords[word] {
					wordCloud[word]++
				}
			}
			
			for _, r := range val {
				if r >= 0x4e00 && r <= 0x9fff {
					char := string(r)
					if !stopWords[char] {
						wordCloud[char]++
					}
				}
			}
		}

		stats.TextAnswers = textAnswers
		stats.WordCloud = wordCloud

	case models.QuestionTypeMatrix:
		matrixStats := make(map[string]map[string]int)
		for _, row := range q.MatrixRows {
			matrixStats[row.Value] = make(map[string]int)
			for _, col := range q.MatrixCols {
				matrixStats[row.Value][col.Value] = 0
			}
		}

		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_json 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var answerJSON json.RawMessage
			if err := rows.Scan(&answerJSON); err != nil {
				return nil, err
			}
			var values map[string]string
			if err := json.Unmarshal(answerJSON, &values); err != nil {
				continue
			}
			for rowVal, colVal := range values {
				if matrixStats[rowVal] != nil {
					matrixStats[rowVal][colVal]++
				}
			}
		}

		stats.MatrixStats = matrixStats

	case models.QuestionTypeSorting:
		sortPatterns := make(map[string]int)

		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_json 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var answerJSON json.RawMessage
			if err := rows.Scan(&answerJSON); err != nil {
				return nil, err
			}
			var values []string
			if err := json.Unmarshal(answerJSON, &values); err != nil {
				continue
			}
			key := strings.Join(values, "|")
			sortPatterns[key]++
		}

		stats.SortPatterns = sortPatterns

	case models.QuestionTypeDate:
		dateCounts := make(map[string]int)

		rows, err := database.DB.Query(context.Background(),
			`SELECT answer_value 
			 FROM response_answers ra
			 JOIN responses r ON ra.response_id = r.id
			 WHERE ra.question_id = $1 AND r.is_completed = true AND answer_value IS NOT NULL`,
			q.ID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err != nil {
				return nil, err
			}
			dateCounts[val]++
		}

		stats.DateCounts = dateCounts
	}

	return stats, nil
}

func (h *StatsHandler) CrosstabAnalysis(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("survey_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid survey id"})
		return
	}

	rowQuestionID, err := uuid.Parse(c.Query("row_question_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid row question id"})
		return
	}

	colQuestionID, err := uuid.Parse(c.Query("col_question_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column question id"})
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

	var rowQ, colQ models.Question
	var rowOptionsJSON, colOptionsJSON json.RawMessage

	err = database.DB.QueryRow(context.Background(),
		`SELECT id, title, question_type, options FROM questions WHERE id = $1 AND survey_id = $2`,
		rowQuestionID, surveyID,
	).Scan(&rowQ.ID, &rowQ.Title, &rowQ.QuestionType, &rowOptionsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "row question not found"})
		return
	}

	err = database.DB.QueryRow(context.Background(),
		`SELECT id, title, question_type, options FROM questions WHERE id = $1 AND survey_id = $2`,
		colQuestionID, surveyID,
	).Scan(&colQ.ID, &colQ.Title, &colQ.QuestionType, &colOptionsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "column question not found"})
		return
	}

	if err := rowQ.OptionsFromJSON(rowOptionsJSON); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := colQ.OptionsFromJSON(colOptionsJSON); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowLabels := make([]string, len(rowQ.Options))
	rowValueToIndex := make(map[string]int)
	for i, opt := range rowQ.Options {
		rowLabels[i] = opt.Label
		rowValueToIndex[opt.Value] = i
	}

	colLabels := make([]string, len(colQ.Options))
	colValueToIndex := make(map[string]int)
	for i, opt := range colQ.Options {
		colLabels[i] = opt.Label
		colValueToIndex[opt.Value] = i
	}

	counts := make([][]int, len(rowLabels))
	for i := range counts {
		counts[i] = make([]int, len(colLabels))
	}

	rows, err := database.DB.Query(context.Background(),
		`SELECT ra1.answer_value, ra2.answer_value
		 FROM response_answers ra1
		 JOIN response_answers ra2 ON ra1.response_id = ra2.response_id
		 JOIN responses r ON ra1.response_id = r.id
		 WHERE ra1.question_id = $1 AND ra2.question_id = $2 AND r.is_completed = true`,
		rowQuestionID, colQuestionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	rowTotals := make([]int, len(rowLabels))
	for rows.Next() {
		var rowVal, colVal string
		if err := rows.Scan(&rowVal, &colVal); err != nil {
			return
		}
		rowIdx, rowOK := rowValueToIndex[rowVal]
		colIdx, colOK := colValueToIndex[colVal]
		if rowOK && colOK {
			counts[rowIdx][colIdx]++
			rowTotals[rowIdx]++
		}
	}

	percentages := make([][]float64, len(rowLabels))
	for i := range percentages {
		percentages[i] = make([]float64, len(colLabels))
		if rowTotals[i] > 0 {
			for j := range percentages[i] {
				percentages[i][j] = float64(counts[i][j]) / float64(rowTotals[i]) * 100
			}
		}
	}

	result := models.CrosstabData{
		RowQuestionID:    rowQ.ID,
		RowQuestionTitle: rowQ.Title,
		RowLabels:        rowLabels,
		ColQuestionID:    colQ.ID,
		ColQuestionTitle: colQ.Title,
		ColLabels:        colLabels,
		Counts:           counts,
		Percentages:      percentages,
	}

	c.JSON(http.StatusOK, result)
}

func (h *StatsHandler) ExportCSV(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("survey_id"))
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

	questions, err := loadQuestions(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	questionIDToTitle := make(map[uuid.UUID]string)
	questionOrder := make([]uuid.UUID, 0, len(questions))
	for _, q := range questions {
		questionIDToTitle[q.ID] = q.Title
		questionOrder = append(questionOrder, q.ID)
	}

	respRows, err := database.DB.Query(context.Background(),
		`SELECT id, submit_time, duration_seconds, is_completed
		 FROM responses WHERE survey_id = $1 ORDER BY created_at`,
		surveyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer respRows.Close()

	type ResponseData struct {
		ID             uuid.UUID
		SubmitTime     *time.Time
		DurationSeconds int
		IsCompleted    bool
		Answers        map[uuid.UUID]string
	}

	var responses []*ResponseData
	for respRows.Next() {
		var r ResponseData
		r.Answers = make(map[uuid.UUID]string)
		if err := respRows.Scan(&r.ID, &r.SubmitTime, &r.DurationSeconds, &r.IsCompleted); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		responses = append(responses, &r)
	}

	for _, r := range responses {
		answerRows, err := database.DB.Query(context.Background(),
			`SELECT question_id, answer_value, answer_json
			 FROM response_answers WHERE response_id = $1`,
			r.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for answerRows.Next() {
			var qid uuid.UUID
			var val *string
			var jsonVal json.RawMessage
			if err := answerRows.Scan(&qid, &val, &jsonVal); err != nil {
				answerRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if val != nil {
				r.Answers[qid] = *val
			} else if len(jsonVal) > 0 {
				r.Answers[qid] = string(jsonVal)
			}
		}
		answerRows.Close()
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=survey_responses_%s.csv", surveyID.String()))
	c.Header("X-Content-Type-Options", "nosniff")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"Response ID", "Submit Time", "Duration (seconds)", "Completed"}
	for _, qid := range questionOrder {
		headers = append(headers, questionIDToTitle[qid])
	}
	writer.Write(headers)

	for _, r := range responses {
		row := []string{
			r.ID.String(),
		}
		if r.SubmitTime != nil {
			row = append(row, r.SubmitTime.Format(time.RFC3339))
		} else {
			row = append(row, "")
		}
		row = append(row, strconv.Itoa(r.DurationSeconds))
		row = append(row, strconv.FormatBool(r.IsCompleted))

		for _, qid := range questionOrder {
			row = append(row, r.Answers[qid])
		}
		writer.Write(row)
	}
}

func (h *StatsHandler) ExportExcel(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	surveyID, err := uuid.Parse(c.Param("survey_id"))
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

	questions, err := loadQuestions(context.Background(), surveyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	questionIDToTitle := make(map[uuid.UUID]string)
	questionOrder := make([]uuid.UUID, 0, len(questions))
	for _, q := range questions {
		questionIDToTitle[q.ID] = q.Title
		questionOrder = append(questionOrder, q.ID)
	}

	respRows, err := database.DB.Query(context.Background(),
		`SELECT id, submit_time, duration_seconds, is_completed
		 FROM responses WHERE survey_id = $1 ORDER BY created_at`,
		surveyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer respRows.Close()

	type ResponseData struct {
		ID             uuid.UUID
		SubmitTime     *time.Time
		DurationSeconds int
		IsCompleted    bool
		Answers        map[uuid.UUID]string
	}

	var responses []*ResponseData
	for respRows.Next() {
		var r ResponseData
		r.Answers = make(map[uuid.UUID]string)
		if err := respRows.Scan(&r.ID, &r.SubmitTime, &r.DurationSeconds, &r.IsCompleted); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		responses = append(responses, &r)
	}

	for _, r := range responses {
		answerRows, err := database.DB.Query(context.Background(),
			`SELECT question_id, answer_value, answer_json
			 FROM response_answers WHERE response_id = $1`,
			r.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for answerRows.Next() {
			var qid uuid.UUID
			var val *string
			var jsonVal json.RawMessage
			if err := answerRows.Scan(&qid, &val, &jsonVal); err != nil {
				answerRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if val != nil {
				r.Answers[qid] = *val
			} else if len(jsonVal) > 0 {
				r.Answers[qid] = string(jsonVal)
			}
		}
		answerRows.Close()
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Responses"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"Response ID", "Submit Time", "Duration (seconds)", "Completed"}
	for _, qid := range questionOrder {
		headers = append(headers, questionIDToTitle[qid])
	}

	for col, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetName, cellName, header)
	}

	for rowIdx, r := range responses {
		rowNum := rowIdx + 2
		
		cellName, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetCellValue(sheetName, cellName, r.ID.String())

		cellName, _ = excelize.CoordinatesToCellName(2, rowNum)
		if r.SubmitTime != nil {
			f.SetCellValue(sheetName, cellName, r.SubmitTime.Format(time.RFC3339))
		}

		cellName, _ = excelize.CoordinatesToCellName(3, rowNum)
		f.SetCellValue(sheetName, cellName, r.DurationSeconds)

		cellName, _ = excelize.CoordinatesToCellName(4, rowNum)
		f.SetCellValue(sheetName, cellName, r.IsCompleted)

		for colIdx, qid := range questionOrder {
			cellName, _ = excelize.CoordinatesToCellName(colIdx+5, rowNum)
			f.SetCellValue(sheetName, cellName, r.Answers[qid])
		}
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=survey_responses_%s.xlsx", surveyID.String()))

	f.Write(c.Writer)
}
