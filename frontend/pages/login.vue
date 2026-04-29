<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900"
  >
    <div
      class="w-full max-w-md p-8 space-y-8 bg-white dark:bg-gray-800 rounded-xl shadow-lg"
    >
      <div class="text-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">登录</h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          登录您的问卷调查平台账户
        </p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-6">
        <div>
          <label
            for="email"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
          >
            邮箱
          </label>
          <UInput
            id="email"
            v-model="state.email"
            type="email"
            placeholder="请输入邮箱"
          />
        </div>

        <div>
          <label
            for="password"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
          >
            密码
          </label>
          <UInput
            id="password"
            v-model="state.password"
            type="password"
            placeholder="请输入密码"
          />
        </div>

        <UButton type="submit" :loading="loading" class="w-full">
          登录
        </UButton>
      </form>

      <div class="text-center">
        <p class="text-gray-600 dark:text-gray-400">
          还没有账户？
          <NuxtLink
            to="/register"
            class="text-blue-600 hover:text-blue-800 dark:text-blue-400"
          >
            立即注册
          </NuxtLink>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: false,
});

const state = ref({
  email: "",
  password: "",
});

const loading = ref(false);
const toast = useToast();

const validate = () => {
  if (!state.value.email.trim()) {
    toast.add({
      title: "请输入邮箱",
      color: "orange",
    });
    return false;
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(state.value.email)) {
    toast.add({
      title: "请输入有效的邮箱地址",
      color: "orange",
    });
    return false;
  }

  if (!state.value.password) {
    toast.add({
      title: "请输入密码",
      color: "orange",
    });
    return false;
  }

  if (state.value.password.length < 6) {
    toast.add({
      title: "密码至少6位",
      color: "orange",
    });
    return false;
  }

  return true;
};

const handleLogin = async () => {
  if (!validate()) {
    return;
  }

  loading.value = true;
  try {
    const authStore = useAuthStore();
    await authStore.login(state.value.email, state.value.password);

    toast.add({
      title: "登录成功",
      color: "green",
    });

    await navigateTo("/");
  } catch (error: any) {
    toast.add({
      title: "登录失败",
      description: error.message || "请检查您的邮箱和密码",
      color: "red",
    });
  } finally {
    loading.value = false;
  }
};

const authStore = useAuthStore();
onMounted(() => {
  if (authStore.isAuthenticated) {
    navigateTo("/");
  }
});
</script>
