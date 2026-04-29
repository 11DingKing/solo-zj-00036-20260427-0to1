<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
    <div class="max-w-2xl mx-auto px-4">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
        ></div>
      </div>

      <div v-else-if="error" class="text-center py-12">
        <div class="text-red-500 text-6xl mb-4">⚠️</div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
          无法访问问卷
        </h2>
        <p class="text-gray-600 dark:text-gray-400">{{ error }}</p>
      </div>

      <div v-else-if="submitted" class="text-center py-12">
        <div class="text-green-500 text-6xl mb-4">✓</div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
          提交成功！
        </h2>
        <p class="text-gray-600 dark:text-gray-400">感谢您的参与</p>
      </div>

      <div
        v-else
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden"
      >
        <div class="bg-blue-600 px-6 py-8 text-white">
          <h1 class="text-2xl font-bold">{{ survey?.title }}</h1>
          <p v-if="survey?.description" class="mt-2 text-blue-100">
            {{ survey.description }}
          </p>
        </div>

        <div
          v-if="currentQuestionIndex < visibleQuestions.length"
          class="p-6 space-y-6"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-500 dark:text-gray-400">
              问题 {{ currentQuestionIndex + 1 }} /
              {{ visibleQuestions.length }}
            </span>
            <div class="flex space-x-1">
              <span
                v-for="(_, index) in visibleQuestions"
                :key="index"
                :class="[
                  'h-2 w-8 rounded-full',
                  index < currentQuestionIndex
                    ? 'bg-green-500'
                    : index === currentQuestionIndex
                      ? 'bg-blue-500'
                      : 'bg-gray-200 dark:bg-gray-700',
                ]"
              ></span>
            </div>
          </div>

          <div class="space-y-4">
            <div class="border-l-4 border-blue-500 pl-4">
              <h3 class="text-lg font-medium text-gray-900 dark:text-white">
                {{ currentQuestion?.title }}
                <span
                  v-if="currentQuestion?.is_required"
                  class="text-red-500 ml-1"
                  >*</span
                >
              </h3>
              <p
                v-if="currentQuestion?.description"
                class="text-sm text-gray-500 dark:text-gray-400 mt-1"
              >
                {{ currentQuestion.description }}
              </p>
            </div>

            <div class="mt-4">
              <div
                v-if="currentQuestion?.question_type === 'single_choice'"
                class="space-y-2"
              >
                <label
                  v-for="option in shuffledOptions(currentQuestion!)"
                  :key="option.value"
                  :class="[
                    'flex items-center p-3 rounded-lg border cursor-pointer transition-colors',
                    answers[currentQuestion!.id] === option.value
                      ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                      : 'border-gray-200 dark:border-gray-600 hover:border-blue-300 dark:hover:border-blue-700',
                  ]"
                  @click="selectOption(option.value)"
                >
                  <input
                    type="radio"
                    :name="`q_${currentQuestion!.id}`"
                    :value="option.value"
                    v-model="answers[currentQuestion!.id]"
                    class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                  />
                  <span class="ml-3 text-gray-700 dark:text-gray-300">{{
                    option.label
                  }}</span>
                </label>
              </div>

              <div
                v-else-if="currentQuestion?.question_type === 'multiple_choice'"
                class="space-y-2"
              >
                <label
                  v-for="option in shuffledOptions(currentQuestion!)"
                  :key="option.value"
                  :class="[
                    'flex items-center p-3 rounded-lg border cursor-pointer transition-colors',
                    getMultipleAnswers(currentQuestion!.id).includes(option.value)
                      ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                      : 'border-gray-200 dark:border-gray-600 hover:border-blue-300 dark:hover:border-blue-700',
                  ]"
                  @click="toggleMultipleOption(option.value)"
                >
                  <input
                    type="checkbox"
                    :checked="
                      getMultipleAnswers(currentQuestion!.id).includes(
                        option.value,
                      )
                    "
                    class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span class="ml-3 text-gray-700 dark:text-gray-300">{{
                    option.label
                  }}</span>
                </label>
              </div>

              <div v-else-if="currentQuestion?.question_type === 'dropdown'">
                <select
                  v-model="answers[currentQuestion.id]"
                  class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">请选择</option>
                  <option
                    v-for="option in shuffledOptions(currentQuestion!)"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
              </div>

              <div v-else-if="currentQuestion?.question_type === 'text'">
                <input
                  v-model="answers[currentQuestion.id]"
                  type="text"
                  placeholder="请输入答案"
                  class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div v-else-if="currentQuestion?.question_type === 'textarea'">
                <textarea
                  v-model="answers[currentQuestion.id]"
                  placeholder="请输入答案"
                  rows="4"
                  class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                ></textarea>
              </div>

              <div
                v-else-if="currentQuestion?.question_type === 'rating'"
                class="space-y-4"
              >
                <div class="flex justify-between items-center px-2">
                  <button
                    v-for="n in (currentQuestion.max_rating - currentQuestion.min_rating + 1)"
                    :key="n"
                    :value="currentQuestion.min_rating + n - 1"
                    @click="selectRating(currentQuestion!.min_rating + n - 1)"
                    :class="[
                      'w-12 h-12 rounded-lg border-2 font-medium transition-all',
                      Number(answers[currentQuestion!.id]) ===
                      currentQuestion!.min_rating + n - 1
                        ? 'border-blue-500 bg-blue-500 text-white'
                        : 'border-gray-200 dark:border-gray-600 hover:border-blue-300 dark:hover:border-blue-700',
                    ]"
                  >
                    {{ currentQuestion.min_rating + n - 1 }}
                  </button>
                </div>
                <div
                  class="flex justify-between text-sm text-gray-500 dark:text-gray-400 px-2"
                >
                  <span>非常不满意</span>
                  <span>非常满意</span>
                </div>
              </div>

              <div
                v-else-if="currentQuestion?.question_type === 'matrix'"
                class="overflow-x-auto"
              >
                <table class="w-full">
                  <thead>
                    <tr class="border-b border-gray-200 dark:border-gray-600">
                      <th
                        class="p-3 text-left text-sm font-medium text-gray-700 dark:text-gray-300"
                      ></th>
                      <th
                        v-for="col in currentQuestion.matrix_cols"
                        :key="col.id"
                        class="p-3 text-center text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ col.label }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="row in currentQuestion.matrix_rows"
                      :key="row.id"
                      class="border-b border-gray-100 dark:border-gray-700"
                    >
                      <td
                        class="p-3 text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ row.label }}
                      </td>
                      <td
                        v-for="col in currentQuestion.matrix_cols"
                        :key="col.id"
                        class="p-3 text-center"
                      >
                        <input
                          type="radio"
                          :name="`q_${currentQuestion!.id}_${row.id}`"
                          :checked="getMatrixValue(row.id) === col.id"
                          @click="setMatrixValue(row.id, col.id)"
                          class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div
                v-else-if="currentQuestion?.question_type === 'sorting'"
                class="space-y-2"
              >
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">
                  拖拽选项以排序
                </p>
                <draggable
                  v-model="sortingOrder[currentQuestion.id]"
                  item-key="value"
                  handle=".sort-handle"
                >
                  <template #item="{ element }">
                    <div
                      class="flex items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600"
                    >
                      <span class="sort-handle cursor-move text-gray-400 mr-3"
                        >⋮⋮</span
                      >
                      <span class="text-gray-700 dark:text-gray-300">{{
                        element.label
                      }}</span>
                    </div>
                  </template>
                </draggable>
              </div>

              <div v-else-if="currentQuestion?.question_type === 'date'">
                <input
                  v-model="answers[currentQuestion.id]"
                  type="date"
                  class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
          </div>

          <div
            class="flex justify-between pt-4 border-t border-gray-200 dark:border-gray-700"
          >
            <UButton
              variant="outline"
              @click="goBack"
              :disabled="currentQuestionIndex === 0"
            >
              上一题
            </UButton>
            <UButton
              v-if="currentQuestionIndex < visibleQuestions.length - 1"
              @click="goNext"
            >
              下一题
            </UButton>
            <UButton
              v-else
              color="green"
              @click="handleSubmit"
              :loading="submitting"
            >
              提交
            </UButton>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Survey, Question, LogicRule } from "~/types";
import draggable from "vuedraggable";

definePageMeta({
  middleware: false,
});

const route = useRoute();
const toast = useToast();

const surveyId = computed(() => route.params.id as string);

const loading = ref(true);
const error = ref("");
const submitted = ref(false);
const submitting = ref(false);

const survey = ref<Survey | null>(null);
const questions = ref<Question[]>([]);
const logicRules = ref<LogicRule[]>([]);
const responseId = ref<string>("");
const startTime = ref<Date>(new Date());

const visibleQuestions = ref<Question[]>([]);
const currentQuestionIndex = ref(0);

const answers = ref<Record<string, any>>({});
const sortingOrder = ref<Record<string, any[]>>({});

const currentQuestion = computed(
  () => visibleQuestions.value[currentQuestionIndex.value],
);

const shuffledOptions = (question: Question) => {
  if (!question.shuffle_options) {
    return question.options;
  }
  const shuffled = [...question.options];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
};

const getMultipleAnswers = (questionId: string) => {
  return answers.value[questionId] || [];
};

const toggleMultipleOption = (value: string) => {
  const questionId = currentQuestion.value!.id;
  if (!answers.value[questionId]) {
    answers.value[questionId] = [];
  }
  const index = answers.value[questionId].indexOf(value);
  if (index > -1) {
    answers.value[questionId].splice(index, 1);
  } else {
    answers.value[questionId].push(value);
  }
};

const selectOption = (value: string) => {
  answers.value[currentQuestion.value!.id] = value;
};

const selectRating = (value: number) => {
  answers.value[currentQuestion.value!.id] = value.toString();
};

const getMatrixValue = (rowId: string) => {
  const questionId = currentQuestion.value!.id;
  return answers.value[`${questionId}_${rowId}`];
};

const setMatrixValue = (rowId: string, colId: string) => {
  const questionId = currentQuestion.value!.id;
  answers.value[`${questionId}_${rowId}`] = colId;
};

const calculateVisibleQuestions = () => {
  const allQuestions = [...questions.value].sort(
    (a, b) => a.question_order - b.question_order,
  );

  const visible: Question[] = [];
  const skipped = new Set<string>();

  for (const question of allQuestions) {
    if (skipped.has(question.id)) continue;
    visible.push(question);

    const relevantRules = logicRules.value
      .filter((rule) => rule.trigger_question_id === question.id)
      .sort((a, b) => a.rule_order - b.rule_order);

    for (const rule of relevantRules) {
      const answer = answers.value[question.id];
      let shouldJump = false;

      if (question.question_type === "multiple_choice") {
        shouldJump =
          Array.isArray(answer) && answer.includes(rule.trigger_option_value);
      } else {
        shouldJump = answer === rule.trigger_option_value;
      }

      if (shouldJump && rule.action_type === "jump") {
        const targetIndex = allQuestions.findIndex(
          (q) => q.id === rule.target_question_id,
        );
        if (targetIndex > -1) {
          const currentIndex = allQuestions.indexOf(question);
          for (let i = currentIndex + 1; i < targetIndex; i++) {
            skipped.add(allQuestions[i].id);
          }
        }
        break;
      }
    }
  }

  visibleQuestions.value = visible;
};

const goBack = () => {
  if (currentQuestionIndex.value > 0) {
    currentQuestionIndex.value--;
  }
};

const goNext = () => {
  const question = currentQuestion.value;
  if (!question) return;

  if (question.is_required) {
    const answer = answers.value[question.id];
    if (question.question_type === "multiple_choice") {
      if (!Array.isArray(answer) || answer.length === 0) {
        toast.add({
          title: "请回答此问题",
          color: "orange",
        });
        return;
      }
    } else if (question.question_type === "matrix") {
      const matrixAnswers = Object.keys(answers.value).filter((key) =>
        key.startsWith(`${question.id}_`),
      );
      if (matrixAnswers.length < question.matrix_rows.length) {
        toast.add({
          title: "请完成矩阵题所有行",
          color: "orange",
        });
        return;
      }
    } else if (
      !answer ||
      (typeof answer === "string" && answer.trim() === "")
    ) {
      toast.add({
        title: "请回答此问题",
        color: "orange",
      });
      return;
    }
  }

  calculateVisibleQuestions();

  if (currentQuestionIndex.value < visibleQuestions.value.length - 1) {
    currentQuestionIndex.value++;
  }
};

const handleSubmit = async () => {
  const question = currentQuestion.value;
  if (question && question.is_required) {
    const answer = answers.value[question.id];
    if (!answer || (typeof answer === "string" && answer.trim() === "")) {
      toast.add({
        title: "请回答此问题",
        color: "orange",
      });
      return;
    }
  }

  submitting.value = true;

  try {
    const api = useApi();

    const allAnswers: {
      question_id: string;
      answer_value: string;
      answer_json?: any;
    }[] = [];

    for (const question of questions.value) {
      if (question.question_type === "multiple_choice") {
        allAnswers.push({
          question_id: question.id,
          answer_value: "",
          answer_json: answers.value[question.id] || [],
        });
      } else if (question.question_type === "matrix") {
        const matrixAnswers: Record<string, string> = {};
        for (const row of question.matrix_rows) {
          const key = `${question.id}_${row.id}`;
          if (answers.value[key]) {
            matrixAnswers[row.id] = answers.value[key];
          }
        }
        allAnswers.push({
          question_id: question.id,
          answer_value: "",
          answer_json: matrixAnswers,
        });
      } else if (question.question_type === "sorting") {
        const order = sortingOrder.value[question.id] || question.options;
        allAnswers.push({
          question_id: question.id,
          answer_value: "",
          answer_json: order.map((o) => o.value),
        });
      } else {
        allAnswers.push({
          question_id: question.id,
          answer_value: answers.value[question.id] || "",
        });
      }
    }

    await api.responses.submit(responseId.value, allAnswers);
    submitted.value = true;
  } catch (error: any) {
    toast.add({
      title: "提交失败",
      description: error.message,
      color: "red",
    });
  } finally {
    submitting.value = false;
  }
};

const loadSurvey = async () => {
  try {
    const api = useApi();
    survey.value = await api.surveys.getForFill(surveyId.value);
    questions.value = survey.value.questions || [];
    logicRules.value = survey.value.logic_rules || [];

    questions.value.sort((a, b) => a.question_order - b.question_order);

    for (const question of questions.value) {
      if (question.question_type === "sorting") {
        sortingOrder.value[question.id] = [...question.options];
      }
    }

    visibleQuestions.value = [...questions.value];

    const result = await api.responses.start(surveyId.value);
    responseId.value = result.response_id;
    startTime.value = new Date(result.start_time);
  } catch (error: any) {
    error.value = error.message || "问卷不存在或已关闭";
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadSurvey();
});
</script>
