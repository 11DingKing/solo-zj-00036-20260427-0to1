package database

import (
	"context"
	"fmt"
	"survey-platform/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB(cfg *config.DBConfig) error {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	var err error
	DB, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func RunMigrations() error {
	_, err := DB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS surveys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(20) DEFAULT 'draft',
			start_time TIMESTAMP WITH TIME ZONE,
			end_time TIMESTAMP WITH TIME ZONE,
			max_responses INTEGER DEFAULT 0,
			require_login BOOLEAN DEFAULT false,
			allow_duplicate BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_surveys_user_id ON surveys(user_id);
		CREATE INDEX IF NOT EXISTS idx_surveys_status ON surveys(status);

		CREATE TABLE IF NOT EXISTS questions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
			question_order INTEGER NOT NULL,
			question_type VARCHAR(30) NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			is_required BOOLEAN DEFAULT true,
			shuffle_options BOOLEAN DEFAULT false,
			options JSONB DEFAULT '[]'::jsonb,
			matrix_rows JSONB DEFAULT '[]'::jsonb,
			matrix_cols JSONB DEFAULT '[]'::jsonb,
			min_rating INTEGER DEFAULT 1,
			max_rating INTEGER DEFAULT 10,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_questions_survey_id ON questions(survey_id);

		CREATE TABLE IF NOT EXISTS logic_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
			trigger_question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
			trigger_option_value TEXT,
			action_type VARCHAR(20) DEFAULT 'jump',
			target_question_id UUID REFERENCES questions(id) ON DELETE CASCADE,
			rule_order INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_logic_rules_survey_id ON logic_rules(survey_id);
		CREATE INDEX IF NOT EXISTS idx_logic_rules_trigger ON logic_rules(trigger_question_id);

		CREATE TABLE IF NOT EXISTS responses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			submitter_ip VARCHAR(45),
			start_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			submit_time TIMESTAMP WITH TIME ZONE,
			is_completed BOOLEAN DEFAULT false,
			duration_seconds INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_responses_survey_id ON responses(survey_id);
		CREATE INDEX IF NOT EXISTS idx_responses_user_id ON responses(user_id);
		CREATE INDEX IF NOT EXISTS idx_responses_completed ON responses(is_completed);

		CREATE TABLE IF NOT EXISTS response_answers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			response_id UUID NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
			question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
			answer_value TEXT,
			answer_json JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_answers_response_id ON response_answers(response_id);
		CREATE INDEX IF NOT EXISTS idx_answers_question_id ON response_answers(question_id);
	`)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
