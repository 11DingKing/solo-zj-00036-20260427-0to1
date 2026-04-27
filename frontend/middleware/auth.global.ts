export default defineNuxtRouteMiddleware(async (to, from) => {
  const authStore = useAuthStore();

  if (!authStore.isAuthenticated) {
    authStore.initializeFromStorage();
    if (authStore.token) {
      await authStore.fetchCurrentUser();
    }
  }

  if (!authStore.isAuthenticated) {
    return navigateTo("/login");
  }
});
