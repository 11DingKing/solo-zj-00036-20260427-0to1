<template>
  <div class="min-h-screen bg-gray-100 dark:bg-gray-900 flex flex-col">
    <nav
      class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700 z-10"
    >
      <div class="max-w-full mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center space-x-4">
            <NuxtLink
              to="/"
              class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            >
              ← 返回
            </NuxtLink>
            <div class="h-6 w-px bg-gray-300 dark:bg-gray-600"></div>
            <input
              v-model="surveyTitle"
              class="text-xl font-semibold text-gray-900 dark:text-white bg-transparent border-none focus:ring-0 p-0"
              @blur="updateTitle"
              @keydown.enter="updateTitle"
            />
          </div>
          <div class="flex items-center space-x-4">
            <UButton variant="outline" @click="saveSurvey"> 保存 </UButton>
            <UButton variant="outline" @click="showSettingsModal = true">
              设置
            </UButton>
            <UButton @click="previewSurvey"> 预览 </UButton>
            <UButton
              v-if="survey.status === 'draft'"
              color="green"
              @click="publishSurvey"
            >
              发布
            </UButton>
          </div>
        </div>
      </div>
    </nav>

    <div class="flex-1 flex overflow-hidden">
      <div
        class="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 overflow-y-auto"
      >
        <div class="p-4">
          <h3
            class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider mb-4"
          >
            题型组件
          </h3>
          <div class="space-y-2">
            <div
              v-for="type in questionTypes"
              :key="type.type"
              draggable="true"
              @dragstart="startDrag($event, type.type)"
              class="flex items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg cursor-move hover:bg-blue-50 dark:hover:bg-blue-900/20 hover:border-blue-200 dark:hover:border-blue-800 border border-gray-200 dark:border-gray-600 transition-colors"
            >
              <span class="text-blue-600 mr-3 text-xl">{{ type.icon }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                type.label
              }}</span>
            </div>
          </div>
        </div>

        <div class="p-4 border-t border-gray-200 dark:border-gray-700">
          <h3
            class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider mb-4"
          >
            逻辑跳转
          </h3>
          <UButton
            variant="outline"
            class="w-full"
            @click="showLogicModal = true"
          >
            配置跳转规则
          </UButton>
        </div>
      </div>

      <div
        class="flex-1 overflow-y-auto p-8"
        @dragover.prevent
        @drop="handleDrop"
      >
        <div class="max-w-3xl mx-auto">
          <div
            v-if="questions.length === 0"
            class="text-center py-24 bg-white dark:bg-gray-800 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600"
          >
            <div class="text-gray-400 dark:text-gray-500 text-6xl mb-4">📝</div>
            <h3
              class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2"
            >
              从左侧拖拽题型到这里
            </h3>
            <p class="text-gray-500 dark:text-gray-400">开始创建您的问卷</p>
          </div>

          <div v-else class="space-y-4">
            <draggable
              v-model="questions"
              :move="checkMove"
              @end="onQuestionOrderChange"
              handle=".drag-handle"
              item-key="id"
            >
              <template #item="{ element: question, index }">
                <div
                  class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden"
                >
                  <div
                    class="flex items-center px-4 py-3 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-200 dark:border-gray-700"
                  >
                    <div
                      class="drag-handle cursor-move text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 mr-3"
                    >
                      <span class="text-lg">⋮⋮</span>
                    </div>
                    <span
                      class="text-sm font-medium text-gray-600 dark:text-gray-400"
                    >
                      第 {{ index + 1 }} 题
                    </span>
                    <UBadge
                      :color="getQuestionTypeColor(question.question_type)"
                      class="ml-2"
                    >
                      {{ getQuestionTypeLabel(question.question_type) }}
                    </UBadge>
                    <div class="flex-1"></div>
                    <div class="flex items-center space-x-2">
                      <UButton
                        size="2xs"
                        variant="outline"
                        @click="toggleRequired(question)"
                      >
                        {{ question.is_required ? "必填" : "可选" }}
                      </UButton>
                      <UButton
                        size="2xs"
                        variant="outline"
                        color="red"
                        @click="deleteQuestion(question)"
                      >
                        删除
                      </UButton>
                    </div>
                  </div>

                  <div class="p-4 space-y-4">
                    <div>
                      <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                      >
                        题目
                      </label>
                      <input
                        v-model="question.title"
                        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="请输入题目"
                      />
                    </div>

                    <div>
                      <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                      >
                        题目描述（可选）
                      </label>
                      <input
                        v-model="question.description"
                        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="题目描述或说明"
                      />
                    </div>

                    <div
                      v-if="
                        [
                          'single_choice',
                          'multiple_choice',
                          'dropdown',
                        ].includes(question.question_type)
                      "
                    >
                      <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                      >
                        选项
                      </label>
                      <div class="space-y-2">
                        <div
                          v-for="(option, optIndex) in question.options"
                          :key="option.id"
                          class="flex items-center space-x-2"
                        >
                          <span class="text-gray-500 text-sm w-6"
                            >{{ optIndex + 1 }}.</span
                          >
                          <input
                            v-model="option.label"
                            class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            placeholder="选项内容"
                          />
                          <UButton
                            size="2xs"
                            variant="outline"
                            color="red"
                            @click="removeOption(question, optIndex)"
                          >
                            ×
                          </UButton>
                        </div>
                        <UButton
                          size="sm"
                          variant="outline"
                          @click="addOption(question)"
                        >
                          + 添加选项
                        </UButton>
                      </div>

                      <div class="mt-3">
                        <label class="inline-flex items-center">
                          <input
                            type="checkbox"
                            v-model="question.shuffle_options"
                            class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                          />
                          <span
                            class="ml-2 text-sm text-gray-700 dark:text-gray-300"
                          >
                            选项随机排列
                          </span>
                        </label>
                      </div>
                    </div>

                    <div v-else-if="question.question_type === 'rating'">
                      <div class="grid grid-cols-2 gap-4">
                        <div>
                          <label
                            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                          >
                            最低分
                          </label>
                          <input
                            v-model.number="question.min_rating"
                            type="number"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          />
                        </div>
                        <div>
                          <label
                            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                          >
                            最高分
                          </label>
                          <input
                            v-model.number="question.max_rating"
                            type="number"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          />
                        </div>
                      </div>
                    </div>

                    <div v-else-if="question.question_type === 'matrix'">
                      <div class="space-y-4">
                        <div>
                          <label
                            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                          >
                            行选项
                          </label>
                          <div class="space-y-2">
                            <div
                              v-for="(row, rowIndex) in question.matrix_rows"
                              :key="row.id"
                              class="flex items-center space-x-2"
                            >
                              <input
                                v-model="row.label"
                                class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                placeholder="行选项"
                              />
                              <UButton
                                size="2xs"
                                variant="outline"
                                color="red"
                                @click="removeMatrixRow(question, rowIndex)"
                              >
                                ×
                              </UButton>
                            </div>
                            <UButton
                              size="sm"
                              variant="outline"
                              @click="addMatrixRow(question)"
                            >
                              + 添加行
                            </UButton>
                          </div>
                        </div>

                        <div>
                          <label
                            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                          >
                            列选项
                          </label>
                          <div class="space-y-2">
                            <div
                              v-for="(col, colIndex) in question.matrix_cols"
                              :key="col.id"
                              class="flex items-center space-x-2"
                            >
                              <input
                                v-model="col.label"
                                class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                placeholder="列选项"
                              />
                              <UButton
                                size="2xs"
                                variant="outline"
                                color="red"
                                @click="removeMatrixCol(question, colIndex)"
                              >
                                ×
                              </UButton>
                            </div>
                            <UButton
                              size="sm"
                              variant="outline"
                              @click="addMatrixCol(question)"
                            >
                              + 添加列
                            </UButton>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div v-else-if="question.question_type === 'sorting'">
                      <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                      >
                        排序选项
                      </label>
                      <div class="space-y-2">
                        <div
                          v-for="(option, optIndex) in question.options"
                          :key="option.id"
                          class="flex items-center space-x-2"
                        >
                          <input
                            v-model="option.label"
                            class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            placeholder="排序选项"
                          />
                          <UButton
                            size="2xs"
                            variant="outline"
                            color="red"
                            @click="removeOption(question, optIndex)"
                          >
                            ×
                          </UButton>
                        </div>
                        <UButton
                          size="sm"
                          variant="outline"
                          @click="addOption(question)"
                        >
                          + 添加排序项
                        </UButton>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </draggable>
          </div>
        </div>
      </div>
    </div>

    <UModal v-model="showSettingsModal" title="问卷设置">
      <div class="space-y-4">
        <UFormGroup label="问卷标题">
          <UInput v-model="surveySettings.title" />
        </UFormGroup>

        <UFormGroup label="问卷描述">
          <UTextarea v-model="surveySettings.description" rows="3" />
        </UFormGroup>

        <div class="grid grid-cols-2 gap-4">
          <UFormGroup label="开始时间">
            <UInput type="datetime-local" v-model="surveySettings.start_time" />
          </UFormGroup>
          <UFormGroup label="结束时间">
            <UInput type="datetime-local" v-model="surveySettings.end_time" />
          </UFormGroup>
        </div>

        <UFormGroup label="最大回收量（0 表示不限）">
          <UInput type="number" v-model.number="surveySettings.max_responses" />
        </UFormGroup>

        <div class="space-y-3">
          <label class="inline-flex items-center">
            <input
              type="checkbox"
              v-model="surveySettings.require_login"
              class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
              需要登录才能填写
            </span>
          </label>

          <label class="inline-flex items-center">
            <input
              type="checkbox"
              v-model="surveySettings.allow_duplicate"
              class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
              允许重复填写
            </span>
          </label>
        </div>

        <div class="flex justify-end gap-3 pt-4">
          <UButton variant="outline" @click="showSettingsModal = false">
            取消
          </UButton>
          <UButton @click="saveSettings"> 保存 </UButton>
        </div>
      </div>
    </UModal>

    <UModal v-model="showLogicModal" title="逻辑跳转规则配置">
      <div class="space-y-4">
        <div v-if="logicRules.length > 0" class="space-y-3">
          <div
            v-for="(rule, index) in logicRules"
            :key="rule.id"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg"
          >
            <div class="text-sm">
              <span class="font-medium text-gray-700 dark:text-gray-300">
                {{ getQuestionTitle(rule.trigger_question_id) }}
              </span>
              <span class="text-gray-500 mx-2">选</span>
              <span class="font-medium text-blue-600 dark:text-blue-400">
                {{
                  getOptionLabel(
                    rule.trigger_question_id,
                    rule.trigger_option_value,
                  )
                }}
              </span>
              <span class="text-gray-500 mx-2">→ 跳转到</span>
              <span class="font-medium text-green-600 dark:text-green-400">
                {{ getQuestionTitle(rule.target_question_id) }}
              </span>
            </div>
            <UButton
              size="2xs"
              variant="outline"
              color="red"
              @click="deleteLogicRule(rule.id, index)"
            >
              删除
            </UButton>
          </div>
        </div>

        <div class="border-t pt-4">
          <h4 class="font-medium text-gray-700 dark:text-gray-300 mb-3">
            添加新规则
          </h4>
          <div class="grid grid-cols-1 gap-3">
            <UFormGroup label="触发题目">
              <USelect
                v-model="newRule.trigger_question_id"
                :options="selectableQuestions"
                placeholder="选择题目"
              />
            </UFormGroup>

            <UFormGroup v-if="newRule.trigger_question_id" label="触发选项">
              <USelect
                v-model="newRule.trigger_option_value"
                :options="getQuestionOptions(newRule.trigger_question_id)"
                placeholder="选择选项"
              />
            </UFormGroup>

            <UFormGroup label="跳转目标题目">
              <USelect
                v-model="newRule.target_question_id"
                :options="selectableQuestions"
                placeholder="选择目标题目"
              />
            </UFormGroup>

            <UButton
              :disabled="
                !newRule.trigger_question_id ||
                !newRule.trigger_option_value ||
                !newRule.target_question_id
              "
              @click="addLogicRule"
            >
              添加规则
            </UButton>
          </div>
        </div>

        <div class="flex justify-end pt-4">
          <UButton @click="saveLogicRules"> 保存并关闭 </UButton>
        </div>
      </div>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type {
  Survey,
  Question,
  QuestionType,
  LogicRule,
  Option,
} from "~/types";
import { QUESTION_TYPE_LABELS } from "~/types";
import draggable from "vuedraggable";

const route = useRoute();
const toast = useToast();

const surveyId = computed(() => route.params.id as string);

const survey = ref<Survey | null>(null);
const surveyTitle = ref("");
const questions = ref<Question[]>([]);
const logicRules = ref<LogicRule[]>([]);

const showSettingsModal = ref(false);
const showLogicModal = ref(false);
const saving = ref(false);

const surveySettings = ref({
  title: "",
  description: "",
  start_time: "",
  end_time: "",
  max_responses: 0,
  require_login: false,
  allow_duplicate: false,
});

const newRule = ref({
  trigger_question_id: "",
  trigger_option_value: "",
  target_question_id: "",
});

const questionTypes = [
  { type: "single_choice", label: "单选题", icon: "○" },
  { type: "multiple_choice", label: "多选题", icon: "☐" },
  { type: "dropdown", label: "下拉选择", icon: "▼" },
  { type: "text", label: "填空题", icon: "_" },
  { type: "textarea", label: "多行文本", icon: "≡" },
  { type: "rating", label: "评分题", icon: "★" },
  { type: "matrix", label: "矩阵题", icon: "▦" },
  { type: "sorting", label: "排序题", icon: "↕" },
  { type: "date", label: "日期选择", icon: "📅" },
];

const generateId = () =>
  `temp_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

const startDrag = (event: DragEvent, type: string) => {
  event.dataTransfer?.setData("questionType", type);
};

const handleDrop = async (event: DragEvent) => {
  const type = event.dataTransfer?.getData("questionType") as QuestionType;
  if (!type) return;

  const newQuestion: Question = {
    id: generateId(),
    survey_id: surveyId.value,
    question_order: questions.value.length + 1,
    question_type: type,
    title: "",
    description: "",
    is_required: true,
    shuffle_options: false,
    options: [],
    matrix_rows: [],
    matrix_cols: [],
    min_rating: 1,
    max_rating: 10,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  if (
    ["single_choice", "multiple_choice", "dropdown", "sorting"].includes(type)
  ) {
    newQuestion.options = [
      { id: generateId(), label: "选项 1", value: "opt_1" },
      { id: generateId(), label: "选项 2", value: "opt_2" },
      { id: generateId(), label: "选项 3", value: "opt_3" },
    ];
  }

  if (type === "matrix") {
    newQuestion.matrix_rows = [
      { id: generateId(), label: "行 1" },
      { id: generateId(), label: "行 2" },
    ];
    newQuestion.matrix_cols = [
      { id: generateId(), label: "列 1" },
      { id: generateId(), label: "列 2" },
      { id: generateId(), label: "列 3" },
    ];
  }

  questions.value.push(newQuestion);
};

const checkMove = (evt: any) => {
  return evt.draggedContext.element.id !== evt.relatedContext.element.id;
};

const onQuestionOrderChange = (evt: any) => {
  questions.value.forEach((q, index) => {
    q.question_order = index + 1;
  });
};

const getQuestionTypeLabel = (type: QuestionType) =>
  QUESTION_TYPE_LABELS[type] || type;

const getQuestionTypeColor = (type: QuestionType) => {
  const colors: Record<QuestionType, string> = {
    single_choice: "blue",
    multiple_choice: "green",
    dropdown: "gray",
    text: "purple",
    textarea: "purple",
    rating: "yellow",
    matrix: "orange",
    sorting: "teal",
    date: "pink",
  };
  return colors[type] || "gray";
};

const toggleRequired = (question: Question) => {
  question.is_required = !question.is_required;
};

const deleteQuestion = (question: Question) => {
  const index = questions.value.findIndex((q) => q.id === question.id);
  if (index !== -1) {
    questions.value.splice(index, 1);
    questions.value.forEach((q, i) => {
      q.question_order = i + 1;
    });
  }
};

const addOption = (question: Question) => {
  const newOpt: Option = {
    id: generateId(),
    label: `选项 ${question.options.length + 1}`,
    value: `opt_${question.options.length + 1}`,
  };
  question.options.push(newOpt);
};

const removeOption = (question: Question, index: number) => {
  question.options.splice(index, 1);
};

const addMatrixRow = (question: Question) => {
  question.matrix_rows.push({
    id: generateId(),
    label: `行 ${question.matrix_rows.length + 1}`,
  });
};

const removeMatrixRow = (question: Question, index: number) => {
  question.matrix_rows.splice(index, 1);
};

const addMatrixCol = (question: Question) => {
  question.matrix_cols.push({
    id: generateId(),
    label: `列 ${question.matrix_cols.length + 1}`,
  });
};

const removeMatrixCol = (question: Question, index: number) => {
  question.matrix_cols.splice(index, 1);
};

const updateTitle = () => {
  if (survey.value) {
    survey.value.title = surveyTitle.value;
  }
};

const saveSurvey = async () => {
  saving.value = true;
  try {
    const api = useApi();

    const idMap: Record<string, string> = {};

    for (const question of questions.value) {
      if (question.id.startsWith("temp_")) {
        const created = await api.questions.create(surveyId.value, {
          question_type: question.question_type,
          title: question.title,
          description: question.description,
          is_required: question.is_required,
          shuffle_options: question.shuffle_options,
          question_order: question.question_order,
          options: question.options,
          matrix_rows: question.matrix_rows,
          matrix_cols: question.matrix_cols,
          min_rating: question.min_rating,
          max_rating: question.max_rating,
        });
        idMap[question.id] = created.id;
        question.id = created.id;
      } else {
        await api.questions.update(question.id, {
          question_type: question.question_type,
          title: question.title,
          description: question.description,
          is_required: question.is_required,
          shuffle_options: question.shuffle_options,
          question_order: question.question_order,
          options: question.options,
          matrix_rows: question.matrix_rows,
          matrix_cols: question.matrix_cols,
          min_rating: question.min_rating,
          max_rating: question.max_rating,
        });
      }
    }

    const orderUpdates = questions.value.map((q) => ({
      question_id: q.id,
      question_order: q.question_order,
    }));

    if (orderUpdates.length > 0) {
      await api.questions.updateOrder(surveyId.value, orderUpdates);
    }

    if (survey.value) {
      survey.value.title = surveyTitle.value;
      await api.surveys.update(surveyId.value, {
        title: surveyTitle.value,
      });
    }

    toast.add({
      title: "已保存",
      color: "green",
    });
  } catch (error: any) {
    toast.add({
      title: "保存失败",
      description: error.message,
      color: "red",
    });
  } finally {
    saving.value = false;
  }
};

const previewSurvey = () => {
  toast.add({
    title: "预览功能开发中",
    color: "orange",
  });
};

const publishSurvey = async () => {
  try {
    const api = useApi();
    await api.surveys.update(surveyId.value, { status: "active" });
    if (survey.value) {
      survey.value.status = "active";
    }
    toast.add({
      title: "发布成功",
      color: "green",
    });
  } catch (error: any) {
    toast.add({
      title: "发布失败",
      description: error.message,
      color: "red",
    });
  }
};

const saveSettings = async () => {
  try {
    const api = useApi();
    const updateData: any = {};
    if (surveySettings.value.title)
      updateData.title = surveySettings.value.title;
    if (surveySettings.value.description)
      updateData.description = surveySettings.value.description;
    if (surveySettings.value.max_responses !== undefined)
      updateData.max_responses = surveySettings.value.max_responses;
    updateData.require_login = surveySettings.value.require_login;
    updateData.allow_duplicate = surveySettings.value.allow_duplicate;

    await api.surveys.update(surveyId.value, updateData);

    if (survey.value) {
      Object.assign(survey.value, updateData);
    }

    showSettingsModal.value = false;
    toast.add({
      title: "设置已保存",
      color: "green",
    });
  } catch (error: any) {
    toast.add({
      title: "保存失败",
      description: error.message,
      color: "red",
    });
  }
};

const selectableQuestions = computed(() => {
  return questions.value.map((q) => ({
    value: q.id,
    label: q.title || `第 ${q.question_order} 题`,
  }));
});

const getQuestionTitle = (id: string) => {
  const q = questions.value.find((x) => x.id === id);
  return q?.title || `题目 ${q?.question_order}`;
};

const getQuestionOptions = (questionId: string) => {
  const q = questions.value.find((x) => x.id === questionId);
  if (!q) return [];
  return q.options.map((opt) => ({
    value: opt.value,
    label: opt.label,
  }));
};

const getOptionLabel = (questionId: string, value: string) => {
  const q = questions.value.find((x) => x.id === questionId);
  if (!q) return value;
  const opt = q.options.find((o) => o.value === value);
  return opt?.label || value;
};

const addLogicRule = () => {
  logicRules.value.push({
    id: generateId(),
    survey_id: surveyId.value,
    trigger_question_id: newRule.value.trigger_question_id,
    trigger_option_value: newRule.value.trigger_option_value,
    action_type: "jump",
    target_question_id: newRule.value.target_question_id,
    rule_order: logicRules.value.length + 1,
    created_at: new Date().toISOString(),
  });

  newRule.value = {
    trigger_question_id: "",
    trigger_option_value: "",
    target_question_id: "",
  };
};

const deleteLogicRule = (id: string, index: number) => {
  logicRules.value.splice(index, 1);
};

const saveLogicRules = () => {
  showLogicModal.value = false;
  toast.add({
    title: "逻辑规则已保存",
    color: "green",
  });
};

const loadSurvey = async () => {
  try {
    const api = useApi();
    survey.value = await api.surveys.get(surveyId.value);
    surveyTitle.value = survey.value.title;
    questions.value = survey.value.questions || [];
    logicRules.value = survey.value.logic_rules || [];

    surveySettings.value = {
      title: survey.value.title,
      description: survey.value.description,
      start_time: survey.value.start_time || "",
      end_time: survey.value.end_time || "",
      max_responses: survey.value.max_responses,
      require_login: survey.value.require_login,
      allow_duplicate: survey.value.allow_duplicate,
    };
  } catch (error: any) {
    toast.add({
      title: "加载失败",
      description: error.message,
      color: "red",
    });
  }
};

onMounted(() => {
  loadSurvey();
});
</script>
