export default defineNuxtRouteMiddleware(async (to, from) => {
  const publicPaths = ["/login", "/register", "/fill/"];
  const isPublicPath = publicPaths.some((path) => to.path.startsWith(path));

  if (isPublicPath) {
    return;
  }

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
