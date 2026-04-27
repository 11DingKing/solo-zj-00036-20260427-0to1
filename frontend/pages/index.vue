<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <nav class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">
              <span class="text-blue-600">📝</span> 问卷调查平台
            </h1>
          </div>
          <div class="flex items-center space-x-4">
            <span class="text-gray-700 dark:text-gray-300">
              {{ authStore.user?.name }}
            </span>
            <UButton variant="outline" @click="authStore.logout">
              退出登录
            </UButton>
          </div>
        </div>
      </div>
    </nav>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="flex justify-between items-center mb-8">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          我的问卷
        </h2>
        <UButton @click="showCreateModal = true">
          <span class="mr-2">+</span> 创建问卷
        </UButton>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>

      <div v-else-if="surveys.length === 0" class="text-center py-12">
        <div class="text-gray-500 dark:text-gray-400 text-6xl mb-4">📋</div>
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
          暂无问卷
        </h3>
        <p class="text-gray-500 dark:text-gray-400 mb-4">
          点击上方按钮创建您的第一个问卷
        </p>
        <UButton @click="showCreateModal = true">
          创建问卷
        </UButton>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="survey in surveys"
          :key="survey.id"
          class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden hover:shadow-md transition-shadow"
        >
          <div class="p-6">
            <div class="flex justify-between items-start mb-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white truncate flex-1">
                {{ survey.title }}
              </h3>
              <UBadge :color="getStatusColor(survey.status)" class="ml-2">
                {{ getStatusLabel(survey.status) }}
              </UBadge>
            </div>

            <p v-if="survey.description" class="text-gray-600 dark:text-gray-400 text-sm mb-4 line-clamp-2">
              {{ survey.description }}
            </p>

            <div class="text-sm text-gray-500 dark:text-gray-400 space-y-1">
              <p>创建时间: {{ formatDate(survey.created_at) }}</p>
            </div>
          </div>

          <div class="bg-gray-50 dark:bg-gray-700 px-6 py-4 border-t border-gray-200 dark:border-gray-600">
            <div class="flex flex-wrap gap-2">
              <UButton size="sm" variant="outline" @click="navigateTo(`/survey/${survey.id}/edit`)">
                编辑
              </UButton>
              <UButton
                v-if="survey.status === 'draft'"
                size="sm"
                @click="publishSurvey(survey.id)"
              >
                发布
              </UButton>
              <UButton
                v-if="survey.status === 'active'"
                size="sm"
                variant="outline"
                color="orange"
                @click="pauseSurvey(survey.id)"
              >
                暂停
              </UButton>
              <UButton
                v-if="survey.status === 'active'"
                size="sm"
                @click="copyLink(survey.id)"
              >
                复制链接
              </UButton>
              <UButton
                v-if="survey.status === 'active'"
                size="sm"
                variant="outline"
                @click="navigateTo(`/survey/${survey.id}/stats`)"
              >
                数据分析
              </UButton>
              <UButton
                size="sm"
                variant="outline"
                color="red"
                @click="confirmDelete(survey)"
              >
                删除
              </UButton>
            </div>
          </div>
        </div>
      </div>
    </main>

    <UModal v-model="showCreateModal" :title="editingSurvey ? '编辑问卷' : '创建问卷'">
      <UForm v-model="createForm" @submit="handleCreate">
        <UFormGroup label="问卷标题" name="title">
          <UInput v-model="createForm.title" placeholder="请输入问卷标题" />
        </UFormGroup>

        <UFormGroup label="问卷描述" name="description">
          <UTextarea v-model="createForm.description" placeholder="请输入问卷描述（可选）" rows="3" />
        </UFormGroup>

        <div class="flex justify-end gap-3 mt-6">
          <UButton variant="outline" @click="showCreateModal = false">
            取消
          </UButton>
          <UButton type="submit">
            {{ editingSurvey ? '保存' : '创建' }}
          </UButton>
        </div>
      </UForm>
    </UModal>

    <UConfirmModal
      v-model="showDeleteModal"
      :title="`删除问卷`"
      :description="`确定要删除问卷 \"${deletingSurvey?.title}\" 吗？此操作不可撤销。`"
      :confirm-label="'删除'"
      confirm-color="red"
      @confirm="handleDelete"
    />
  </div>
</template>

<script setup lang="ts">
import type { Survey, SurveyStatus } from '~/types'

const authStore = useAuthStore()
const toast = useToast()

const loading = ref(true)
const surveys = ref<Survey[]>([])
const showCreateModal = ref(false)
const editingSurvey = ref<Survey | null>(null)
const showDeleteModal = ref(false)
const deletingSurvey = ref<Survey | null>(null)

const createForm = ref({
  title: '',
  description: ''
})

const loadSurveys = async () => {
  loading.value = true
  try {
    const api = useApi()
    surveys.value = await api.surveys.list()
  } catch (error: any) {
    toast.add({
      title: '加载失败',
      description: error.message,
      color: 'red'
    })
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  if (!createForm.value.title.trim()) {
    return
  }

  try {
    const api = useApi()
    await api.surveys.create(createForm.value.title)
    toast.add({
      title: '创建成功',
      color: 'green'
    })
    showCreateModal.value = false
    createForm.value = { title: '', description: '' }
    loadSurveys()
  } catch (error: any) {
    toast.add({
      title: '创建失败',
      description: error.message,
      color: 'red'
    })
  }
}

const publishSurvey = async (id: string) => {
  try {
    const api = useApi()
    await api.surveys.update(id, { status: 'active' })
    toast.add({
      title: '发布成功',
      color: 'green'
    })
    loadSurveys()
  } catch (error: any) {
    toast.add({
      title: '发布失败',
      description: error.message,
      color: 'red'
    })
  }
}

const pauseSurvey = async (id: string) => {
  try {
    const api = useApi()
    await api.surveys.update(id, { status: 'paused' })
    toast.add({
      title: '已暂停',
      color: 'orange'
    })
    loadSurveys()
  } catch (error: any) {
    toast.add({
      title: '操作失败',
      description: error.message,
      color: 'red'
    })
  }
}

const copyLink = (id: string) => {
  const link = `${window.location.origin}/fill/${id}`
  navigator.clipboard.writeText(link)
  toast.add({
    title: '链接已复制',
    description: link,
    color: 'green'
  })
}

const confirmDelete = (survey: Survey) => {
  deletingSurvey.value = survey
  showDeleteModal.value = true
}

const handleDelete = async () => {
  if (!deletingSurvey.value) return

  try {
    const api = useApi()
    await api.surveys.delete(deletingSurvey.value.id)
    toast.add({
      title: '删除成功',
      color: 'green'
    })
    loadSurveys()
  } catch (error: any) {
    toast.add({
      title: '删除失败',
      description: error.message,
      color: 'red'
    })
  } finally {
    showDeleteModal.value = false
    deletingSurvey.value = null
  }
}

const getStatusColor = (status: SurveyStatus) => {
  const colors: Record<SurveyStatus, string> = {
    draft: 'gray',
    active: 'green',
    paused: 'orange',
    closed: 'red'
  }
  return colors[status] || 'gray'
}

const getStatusLabel = (status: SurveyStatus) => {
  const labels: Record<SurveyStatus, string> = {
    draft: '草稿',
    active: '发布中',
    paused: '已暂停',
    closed: '已关闭'
  }
  return labels[status] || status
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN')
}

onMounted(() => {
  loadSurveys()
})
</script>
