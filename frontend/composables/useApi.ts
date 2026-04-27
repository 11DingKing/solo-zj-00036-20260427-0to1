import type {
  User,
  Survey,
  Question,
  LogicRule,
  SurveyStats,
  QuestionStats,
  CrosstabData,
} from "~/types";

const getAuthToken = () => {
  if (import.meta.client) {
    return localStorage.getItem("auth_token");
  }
  return null;
};

const apiFetch = async <T>(
  url: string,
  options: {
    method?: "GET" | "POST" | "PUT" | "DELETE";
    body?: any;
    requiresAuth?: boolean;
  } = {},
): Promise<T> => {
  const config = useRuntimeConfig();
  const baseUrl = config.public.apiBase;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  const token = getAuthToken();
  if (token && options.requiresAuth !== false) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const fetchOptions: RequestInit = {
    method: options.method || "GET",
    headers,
  };

  if (options.body) {
    fetchOptions.body = JSON.stringify(options.body);
  }

  const response = await fetch(`${baseUrl}${url}`, fetchOptions);

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Unknown error" }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  if (response.headers.get("Content-Type")?.includes("application/json")) {
    return await response.json();
  }
  return null as T;
};

export const useApi = () => {
  const auth = {
    register: (email: string, password: string, name: string) =>
      apiFetch<{ token: string; user: User }>("/auth/register", {
        method: "POST",
        body: { email, password, name },
        requiresAuth: false,
      }),

    login: (email: string, password: string) =>
      apiFetch<{ token: string; user: User }>("/auth/login", {
        method: "POST",
        body: { email, password },
        requiresAuth: false,
      }),

    getCurrentUser: () => apiFetch<User>("/auth/me"),
  };

  const surveys = {
    list: () => apiFetch<Survey[]>("/surveys"),

    create: (title: string) =>
      apiFetch<Survey>("/surveys", {
        method: "POST",
        body: { title },
      }),

    get: (id: string) => apiFetch<Survey>(`/surveys/${id}`),

    getForFill: (id: string) =>
      apiFetch<Survey>(`/surveys/${id}/fill`, { requiresAuth: false }),

    update: (id: string, data: Partial<Survey>) =>
      apiFetch<Survey>(`/surveys/${id}`, {
        method: "PUT",
        body: data,
      }),

    delete: (id: string) =>
      apiFetch<void>(`/surveys/${id}`, { method: "DELETE" }),
  };

  const questions = {
    create: (surveyId: string, data: Partial<Question>) =>
      apiFetch<Question>(`/surveys/${surveyId}/questions`, {
        method: "POST",
        body: data,
      }),

    update: (id: string, data: Partial<Question>) =>
      apiFetch<Question>(`/surveys/questions/${id}`, {
        method: "PUT",
        body: data,
      }),

    delete: (id: string) =>
      apiFetch<void>(`/surveys/questions/${id}`, { method: "DELETE" }),

    updateOrder: (
      surveyId: string,
      orders: { question_id: string; question_order: number }[],
    ) =>
      apiFetch<void>(`/surveys/${surveyId}/questions/order`, {
        method: "PUT",
        body: orders,
      }),
  };

  const logic = {
    list: (surveyId: string) =>
      apiFetch<LogicRule[]>(`/surveys/${surveyId}/logic`),

    create: (surveyId: string, data: Partial<LogicRule>) =>
      apiFetch<LogicRule>(`/surveys/${surveyId}/logic`, {
        method: "POST",
        body: data,
      }),

    delete: (id: string) =>
      apiFetch<void>(`/surveys/logic/${id}`, { method: "DELETE" }),

    updateOrder: (
      surveyId: string,
      orders: { rule_id: string; rule_order: number }[],
    ) =>
      apiFetch<void>(`/surveys/${surveyId}/logic/order`, {
        method: "PUT",
        body: orders,
      }),
  };

  const responses = {
    start: (surveyId: string) =>
      apiFetch<{ response_id: string; start_time: string }>(
        "/responses/start",
        {
          method: "POST",
          body: { survey_id: surveyId },
          requiresAuth: false,
        },
      ),

    submit: (
      responseId: string,
      answers: {
        question_id: string;
        answer_value: string;
        answer_json?: any;
      }[],
    ) =>
      apiFetch<{ message: string; duration_seconds: number }>(
        "/responses/submit",
        {
          method: "POST",
          body: { response_id: responseId, answers },
          requiresAuth: false,
        },
      ),
  };

  const stats = {
    getSurveyStats: (surveyId: string) =>
      apiFetch<SurveyStats>(`/surveys/${surveyId}/stats`),

    getQuestionStats: (surveyId: string) =>
      apiFetch<QuestionStats[]>(`/surveys/${surveyId}/questions-stats`),

    getCrosstab: (
      surveyId: string,
      rowQuestionId: string,
      colQuestionId: string,
    ) =>
      apiFetch<CrosstabData>(
        `/surveys/${surveyId}/crosstab?row_question_id=${rowQuestionId}&col_question_id=${colQuestionId}`,
      ),

    exportCSV: (surveyId: string) => {
      const config = useRuntimeConfig();
      const token = getAuthToken();
      window.open(
        `${config.public.apiBase}/surveys/${surveyId}/export/csv?token=${token}`,
        "_blank",
      );
    },

    exportExcel: (surveyId: string) => {
      const config = useRuntimeConfig();
      const token = getAuthToken();
      window.open(
        `${config.public.apiBase}/surveys/${surveyId}/export/excel?token=${token}`,
        "_blank",
      );
    },
  };

  return {
    auth,
    surveys,
    questions,
    logic,
    responses,
    stats,
  };
};
