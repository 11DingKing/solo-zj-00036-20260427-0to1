<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <nav
      class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700"
    >
      <div class="max-w-full mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16 items-center">
          <div class="flex items-center space-x-4">
            <NuxtLink
              to="/"
              class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            >
              ← 返回
            </NuxtLink>
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">
              数据分析：{{ survey?.title }}
            </h1>
          </div>
          <div class="flex items-center space-x-3">
            <UButton variant="outline" @click="exportCSV"> 导出 CSV </UButton>
            <UButton variant="outline" @click="exportExcel">
              导出 Excel
            </UButton>
          </div>
        </div>
      </div>
    </nav>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
        ></div>
      </div>

      <template v-else>
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div
            class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center">
              <div class="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <span class="text-2xl">📊</span>
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-500 dark:text-gray-400">总回收量</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ surveyStats?.total_responses || 0 }}
                </p>
              </div>
            </div>
          </div>

          <div
            class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center">
              <div class="p-3 bg-green-100 dark:bg-green-900/30 rounded-lg">
                <span class="text-2xl">✅</span>
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-500 dark:text-gray-400">完成数</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ surveyStats?.completed_count || 0 }}
                </p>
              </div>
            </div>
          </div>

          <div
            class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center">
              <div class="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                <span class="text-2xl">📈</span>
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-500 dark:text-gray-400">完成率</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ surveyStats?.completion_rate?.toFixed(1) || 0 }}%
                </p>
              </div>
            </div>
          </div>

          <div
            class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center">
              <div class="p-3 bg-orange-100 dark:bg-orange-900/30 rounded-lg">
                <span class="text-2xl">⏱️</span>
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-500 dark:text-gray-400">平均时长</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ surveyStats?.average_duration?.toFixed(0) || 0 }}秒
                </p>
              </div>
            </div>
          </div>
        </div>

        <div
          class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700 mb-8"
        >
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            交叉分析
          </h3>
          <div class="flex flex-wrap gap-4 items-end">
            <UFormGroup label="行题目">
              <USelect
                v-model="crosstabRow"
                :options="questionOptions"
                placeholder="选择题目"
              />
            </UFormGroup>
            <UFormGroup label="列题目">
              <USelect
                v-model="crosstabCol"
                :options="questionOptions"
                placeholder="选择题目"
              />
            </UFormGroup>
            <UButton @click="loadCrosstab"> 分析 </UButton>
          </div>

          <div v-if="crosstabData" class="mt-6 overflow-x-auto">
            <h4 class="font-medium text-gray-700 dark:text-gray-300 mb-3">
              {{ crosstabData.row_question_title }} ×
              {{ crosstabData.col_question_title }}
            </h4>
            <table
              class="min-w-full border border-gray-200 dark:border-gray-600"
            >
              <thead>
                <tr class="bg-gray-50 dark:bg-gray-700">
                  <th
                    class="p-3 text-left border-b border-gray-200 dark:border-gray-600"
                  ></th>
                  <th
                    v-for="(col, index) in crosstabData.col_labels"
                    :key="index"
                    class="p-3 text-center border-b border-r border-gray-200 dark:border-gray-600 text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ col }}
                  </th>
                  <th
                    class="p-3 text-center border-b border-gray-200 dark:border-gray-600 text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    合计
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(row, rowIndex) in crosstabData.row_labels"
                  :key="rowIndex"
                  class="hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <td
                    class="p-3 border-b border-r border-gray-200 dark:border-gray-600 font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ row }}
                  </td>
                  <td
                    v-for="(col, colIndex) in crosstabData.col_labels"
                    :key="colIndex"
                    class="p-3 border-b border-r border-gray-200 dark:border-gray-600 text-center"
                  >
                    <div>{{ crosstabData.counts[rowIndex][colIndex] }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{
                        crosstabData.percentages[rowIndex][colIndex]?.toFixed(
                          1,
                        )
                      }}%
                    </div>
                  </td>
                  <td
                    class="p-3 border-b border-gray-200 dark:border-gray-600 text-center font-medium"
                  >
                    {{
                      crosstabData.counts[rowIndex].reduce(
                        (a: number, b: number) => a + b,
                        0,
                      )
                    }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="space-y-8">
          <div
            v-for="(stats, index) in questionStatsList"
            :key="stats.question_id"
            class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700"
            :id="`question-${index}`"
          >
            <div class="flex justify-between items-start mb-4">
              <div>
                <span
                  class="inline-block px-2 py-1 text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 rounded mb-2"
                >
                  第 {{ index + 1 }} 题 - {{ getQuestionTypeLabel(stats.type) }}
                </span>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ getQuestionTitle(stats.question_id) }}
                </h3>
              </div>
              <UButton size="sm" variant="outline" @click="exportChart(index)">
                导出图表
              </UButton>
            </div>

            <div
              v-if="
                ['single_choice', 'multiple_choice', 'dropdown'].includes(
                  stats.type,
                )
              "
            >
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div ref="el => chartRefs[index] = el" class="h-80">
                  <Bar
                    v-if="stats.option_counts"
                    :data="getBarChartData(stats)"
                    :options="chartOptions"
                  />
                </div>
                <div class="space-y-3">
                  <h4 class="font-medium text-gray-700 dark:text-gray-300">
                    选项分布
                  </h4>
                  <div
                    v-for="option in getQuestionOptions(stats.question_id)"
                    :key="option.value"
                    class="flex items-center justify-between"
                  >
                    <span class="text-gray-700 dark:text-gray-300">{{
                      option.label
                    }}</span>
                    <div class="flex items-center space-x-3">
                      <div
                        class="w-32 bg-gray-200 dark:bg-gray-700 rounded-full h-2"
                      >
                        <div
                          class="bg-blue-500 h-2 rounded-full"
                          :style="{
                            width: `${stats.percentages?.[option.value] || 0}%`,
                          }"
                        ></div>
                      </div>
                      <span
                        class="text-sm text-gray-500 dark:text-gray-400 w-20 text-right"
                      >
                        {{ stats.option_counts?.[option.value] || 0 }} ({{
                          stats.percentages?.[option.value]?.toFixed(1) || 0
                        }}%)
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="stats.type === 'rating'">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="text-center py-8">
                  <div
                    class="text-5xl font-bold text-blue-600 dark:text-blue-400 mb-2"
                  >
                    {{ stats.average_rating?.toFixed(1) || 0 }}
                  </div>
                  <div class="text-gray-500 dark:text-gray-400">平均分</div>
                </div>
                <div ref="el => chartRefs[index] = el" class="h-80">
                  <Bar
                    v-if="stats.rating_counts"
                    :data="getRatingChartData(stats)"
                    :options="chartOptions"
                  />
                </div>
              </div>
            </div>

            <div v-else-if="stats.type === 'text' || stats.type === 'textarea'">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div
                  ref="el => chartRefs[index] = el"
                  class="h-80 flex items-center justify-center"
                >
                  <div
                    ref="el => wordCloudRefs[index] = el"
                    class="w-full h-full"
                  ></div>
                </div>
                <div class="space-y-2 max-h-80 overflow-y-auto">
                  <h4 class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                    回答示例（前50条）
                  </h4>
                  <div
                    v-for="(text, i) in (stats.text_answers || []).slice(0, 50)"
                    :key="i"
                    class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg text-sm text-gray-700 dark:text-gray-300"
                  >
                    {{ text }}
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="stats.type === 'matrix'">
              <div class="overflow-x-auto" ref="el => chartRefs[index] = el">
                <table
                  class="min-w-full border border-gray-200 dark:border-gray-600"
                >
                  <thead>
                    <tr class="bg-gray-50 dark:bg-gray-700">
                      <th
                        class="p-3 text-left border-b border-gray-200 dark:border-gray-600"
                      ></th>
                      <th
                        v-for="col in getMatrixCols(stats.question_id)"
                        :key="col.id"
                        class="p-3 text-center border-b border-r border-gray-200 dark:border-gray-600 text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ col.label }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="row in getMatrixRows(stats.question_id)"
                      :key="row.id"
                    >
                      <td
                        class="p-3 border-b border-r border-gray-200 dark:border-gray-600 font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ row.label }}
                      </td>
                      <td
                        v-for="col in getMatrixCols(stats.question_id)"
                        :key="col.id"
                        :class="[
                          'p-3 border-b border-r border-gray-200 dark:border-gray-600 text-center',
                          getHeatmapClass(stats, row.id, col.id),
                        ]"
                      >
                        {{ stats.matrix_stats?.[row.id]?.[col.id] || 0 }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div v-else-if="stats.type === 'sorting'">
              <div class="space-y-3">
                <h4 class="font-medium text-gray-700 dark:text-gray-300">
                  排序模式分布
                </h4>
                <div
                  v-for="(count, pattern) in stats.sort_patterns"
                  :key="pattern"
                  class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg"
                >
                  <span class="text-gray-700 dark:text-gray-300">
                    {{ getSortPatternLabel(pattern) }}
                  </span>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ count }} 次
                  </span>
                </div>
              </div>
            </div>

            <div v-else-if="stats.type === 'date'">
              <div ref="el => chartRefs[index] = el" class="h-80">
                <Bar
                  v-if="stats.date_counts"
                  :data="getDateChartData(stats)"
                  :options="chartOptions"
                />
              </div>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from "vue";
import { Bar } from "vue-chartjs";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from "chart.js";
import type { ChartOptions } from "chart.js";
import type { Survey, Question, QuestionStats, CrosstabData } from "~/types";
import { QUESTION_TYPE_LABELS } from "~/types";
import WordCloud from "wordcloud";

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
);

const route = useRoute();
const toast = useToast();

const surveyId = computed(() => route.params.id as string);

const loading = ref(true);
const survey = ref<Survey | null>(null);
const questions = ref<Question[]>([]);
const surveyStats = ref<any>(null);
const questionStatsList = ref<QuestionStats[]>([]);

const crosstabRow = ref("");
const crosstabCol = ref("");
const crosstabData = ref<CrosstabData | null>(null);

const chartRefs = ref<(HTMLElement | null)[]>([]);
const wordCloudRefs = ref<(HTMLElement | null)[]>([]);

const questionOptions = computed(() => {
  return questions.value
    .filter((q) => ["single_choice", "dropdown"].includes(q.question_type))
    .map((q) => ({
      value: q.id,
      label: q.title,
    }));
});

const chartOptions: ChartOptions<"bar"> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
  },
  scales: {
    y: {
      beginAtZero: true,
    },
  },
};

const getQuestionTitle = (id: string) => {
  const q = questions.value.find((x) => x.id === id);
  return q?.title || "未知题目";
};

const getQuestionTypeLabel = (type: string) => {
  return QUESTION_TYPE_LABELS[type as any] || type;
};

const getQuestionOptions = (id: string) => {
  const q = questions.value.find((x) => x.id === id);
  return q?.options || [];
};

const getMatrixRows = (id: string) => {
  const q = questions.value.find((x) => x.id === id);
  return q?.matrix_rows || [];
};

const getMatrixCols = (id: string) => {
  const q = questions.value.find((x) => x.id === id);
  return q?.matrix_cols || [];
};

const getBarChartData = (stats: QuestionStats) => {
  const options = getQuestionOptions(stats.question_id);
  return {
    labels: options.map((o) => o.label),
    datasets: [
      {
        data: options.map((o) => stats.option_counts?.[o.value] || 0),
        backgroundColor: "rgba(59, 130, 246, 0.7)",
        borderColor: "rgb(59, 130, 246)",
        borderWidth: 1,
      },
    ],
  };
};

const getRatingChartData = (stats: QuestionStats) => {
  const q = questions.value.find((x) => x.id === stats.question_id);
  const minRating = q?.min_rating || 1;
  const maxRating = q?.max_rating || 10;

  const labels: string[] = [];
  const data: number[] = [];
  for (let i = minRating; i <= maxRating; i++) {
    labels.push(i.toString());
    data.push(stats.rating_counts?.[i] || 0);
  }

  return {
    labels,
    datasets: [
      {
        data,
        backgroundColor: "rgba(234, 179, 8, 0.7)",
        borderColor: "rgb(234, 179, 8)",
        borderWidth: 1,
      },
    ],
  };
};

const getDateChartData = (stats: QuestionStats) => {
  const entries = Object.entries(stats.date_counts || {}).sort();
  return {
    labels: entries.map(([date]) => date),
    datasets: [
      {
        data: entries.map(([, count]) => count),
        backgroundColor: "rgba(236, 72, 153, 0.7)",
        borderColor: "rgb(236, 72, 153)",
        borderWidth: 1,
      },
    ],
  };
};

const getHeatmapClass = (
  stats: QuestionStats,
  rowId: string,
  colId: string,
) => {
  const count = stats.matrix_stats?.[rowId]?.[colId] || 0;
  const maxCount = Math.max(
    ...Object.values(stats.matrix_stats || {}).flatMap((row) =>
      Object.values(row || {}),
    ),
  );

  if (maxCount === 0) return "";

  const intensity = count / maxCount;

  if (intensity > 0.8) return "bg-red-200 dark:bg-red-900/50";
  if (intensity > 0.6) return "bg-orange-200 dark:bg-orange-900/50";
  if (intensity > 0.4) return "bg-yellow-200 dark:bg-yellow-900/50";
  if (intensity > 0.2) return "bg-green-200 dark:bg-green-900/50";
  return "bg-blue-100 dark:bg-blue-900/30";
};

const getSortPatternLabel = (pattern: string) => {
  const parts = pattern.split("|");
  return parts
    .map((val, idx) => {
      const q = questions.value
        .flatMap((q) => q.options)
        .find((o) => o.value === val);
      return `${idx + 1}. ${q?.label || val}`;
    })
    .join(" → ");
};

const renderWordCloud = (index: number) => {
  const stats = questionStatsList.value[index];
  const el = wordCloudRefs.value[index];
  if (!el || !stats.word_cloud || Object.keys(stats.word_cloud).length === 0)
    return;

  const words = Object.entries(stats.word_cloud)
    .filter(([word]) => word.length > 1)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 100)
    .map(([word, count]) => [word, count]);

  if (words.length === 0) return;

  WordCloud(el, {
    list: words,
    gridSize: 8,
    weightFactor: 5,
    fontFamily: "sans-serif",
    color: () => {
      const colors = [
        "#3B82F6",
        "#10B981",
        "#F59E0B",
        "#EF4444",
        "#8B5CF6",
        "#EC4899",
      ];
      return colors[Math.floor(Math.random() * colors.length)];
    },
    rotateRatio: 0.5,
    rotationSteps: 2,
    backgroundColor: "transparent",
  });
};

const exportChart = (index: number) => {
  const el = chartRefs.value[index];
  if (!el) return;

  const canvas = el.querySelector("canvas");
  if (canvas) {
    const link = document.createElement("a");
    link.download = `chart-${index + 1}.png`;
    link.href = canvas.toDataURL();
    link.click();
  }
};

const exportCSV = () => {
  const api = useApi();
  api.stats.exportCSV(surveyId.value);
};

const exportExcel = () => {
  const api = useApi();
  api.stats.exportExcel(surveyId.value);
};

const loadCrosstab = async () => {
  if (!crosstabRow.value || !crosstabCol.value) return;

  try {
    const api = useApi();
    crosstabData.value = await api.stats.getCrosstab(
      surveyId.value,
      crosstabRow.value,
      crosstabCol.value,
    );
  } catch (error: any) {
    toast.add({
      title: "分析失败",
      description: error.message,
      color: "red",
    });
  }
};

const loadData = async () => {
  try {
    const api = useApi();

    survey.value = await api.surveys.get(surveyId.value);
    questions.value = survey.value.questions || [];

    surveyStats.value = await api.stats.getSurveyStats(surveyId.value);
    questionStatsList.value = await api.stats.getQuestionStats(surveyId.value);
  } catch (error: any) {
    toast.add({
      title: "加载失败",
      description: error.message,
      color: "red",
    });
  } finally {
    loading.value = false;
  }
};

watch(loading, async (isLoading) => {
  if (!isLoading) {
    await nextTick();
    questionStatsList.value.forEach((stats, index) => {
      if (
        (stats.type === "text" || stats.type === "textarea") &&
        stats.word_cloud
      ) {
        renderWordCloud(index);
      }
    });
  }
});

onMounted(() => {
  loadData();
});
</script>
