export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export type SurveyStatus = "draft" | "active" | "paused" | "closed";

export interface Survey {
  id: string;
  user_id: string;
  title: string;
  description: string;
  status: SurveyStatus;
  start_time: string | null;
  end_time: string | null;
  max_responses: number;
  require_login: boolean;
  allow_duplicate: boolean;
  questions?: Question[];
  logic_rules?: LogicRule[];
  created_at: string;
  updated_at: string;
}

export type QuestionType =
  | "single_choice"
  | "multiple_choice"
  | "dropdown"
  | "text"
  | "textarea"
  | "rating"
  | "matrix"
  | "sorting"
  | "date";

export interface Option {
  id: string;
  label: string;
  value: string;
}

export interface MatrixRow {
  id: string;
  label: string;
}

export interface MatrixCol {
  id: string;
  label: string;
}

export interface Question {
  id: string;
  survey_id: string;
  question_order: number;
  question_type: QuestionType;
  title: string;
  description: string;
  is_required: boolean;
  shuffle_options: boolean;
  options: Option[];
  matrix_rows: MatrixRow[];
  matrix_cols: MatrixCol[];
  min_rating: number;
  max_rating: number;
  created_at: string;
  updated_at: string;
}

export type LogicAction = "jump";

export interface LogicRule {
  id: string;
  survey_id: string;
  trigger_question_id: string;
  trigger_option_value: string;
  action_type: LogicAction;
  target_question_id: string;
  rule_order: number;
  created_at: string;
}

export interface Response {
  id: string;
  survey_id: string;
  user_id: string | null;
  submitter_ip: string;
  start_time: string;
  submit_time: string | null;
  is_completed: boolean;
  duration_seconds: number;
  created_at: string;
}

export interface SurveyStats {
  total_responses: number;
  completed_count: number;
  completion_rate: number;
  average_duration: number;
}

export interface QuestionStats {
  question_id: string;
  type: string;
  option_counts?: Record<string, number>;
  percentages?: Record<string, number>;
  average_rating?: number;
  rating_counts?: Record<string, number>;
  text_answers?: string[];
  word_cloud?: Record<string, number>;
  matrix_stats?: Record<string, Record<string, number>>;
  sort_patterns?: Record<string, number>;
  date_counts?: Record<string, number>;
}

export interface CrosstabData {
  row_question_id: string;
  row_question_title: string;
  row_labels: string[];
  col_question_id: string;
  col_question_title: string;
  col_labels: string[];
  counts: number[][];
  percentages: number[][];
}

export const QUESTION_TYPE_LABELS: Record<QuestionType, string> = {
  single_choice: "单选题",
  multiple_choice: "多选题",
  dropdown: "下拉选择",
  text: "填空题",
  textarea: "多行文本",
  rating: "评分题",
  matrix: "矩阵题",
  sorting: "排序题",
  date: "日期选择",
};

export const QUESTION_TYPE_ICONS: Record<QuestionType, string> = {
  single_choice: "i-heroicons-list-bullet",
  multiple_choice: "i-heroicons-check-square",
  dropdown: "i-heroicons-chevron-down",
  text: "i-heroicons-document-text",
  textarea: "i-heroicons-document",
  rating: "i-heroicons-star",
  matrix: "i-heroicons-table-cells",
  sorting: "i-heroicons-arrows-up-down",
  date: "i-heroicons-calendar",
};
