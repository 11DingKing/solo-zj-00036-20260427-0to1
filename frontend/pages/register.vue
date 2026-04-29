<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900"
  >
    <div
      class="w-full max-w-md p-8 space-y-8 bg-white dark:bg-gray-800 rounded-xl shadow-lg"
    >
      <div class="text-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">注册</h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          创建您的问卷调查平台账户
        </p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-6">
        <UFormGroup label="姓名" name="name">
          <UInput v-model="state.name" placeholder="请输入姓名" />
        </UFormGroup>

        <UFormGroup label="邮箱" name="email">
          <UInput v-model="state.email" type="email" placeholder="请输入邮箱" />
        </UFormGroup>

        <UFormGroup label="密码" name="password">
          <UInput
            v-model="state.password"
            type="password"
            placeholder="请输入密码（至少6位）"
          />
        </UFormGroup>

        <UFormGroup label="确认密码" name="confirmPassword">
          <UInput
            v-model="state.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
          />
        </UFormGroup>

        <UButton type="submit" :loading="loading" class="w-full">
          注册
        </UButton>
      </form>

      <div class="text-center">
        <p class="text-gray-600 dark:text-gray-400">
          已有账户？
          <NuxtLink
            to="/login"
            class="text-blue-600 hover:text-blue-800 dark:text-blue-400"
          >
            立即登录
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
  name: "",
  email: "",
  password: "",
  confirmPassword: "",
});

const loading = ref(false);
const toast = useToast();

const handleRegister = async () => {
  loading.value = true;
  try {
    const authStore = useAuthStore();
    await authStore.register(
      state.value.email,
      state.value.password,
      state.value.name,
    );

    toast.add({
      title: "注册成功",
      color: "green",
    });

    await navigateTo("/");
  } catch (error: any) {
    toast.add({
      title: "注册失败",
      description: error.message || "请检查您的输入信息",
      color: "red",
    });
  } finally {
    loading.value = false;
  }
};
</script>
