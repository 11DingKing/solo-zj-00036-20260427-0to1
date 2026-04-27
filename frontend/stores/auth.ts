import { defineStore } from "pinia";
import type { User } from "~/types";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null as User | null,
    token: null as string | null,
  }),

  getters: {
    isAuthenticated: (state) => !!state.token && !!state.user,
  },

  actions: {
    async login(email: string, password: string) {
      const api = useApi();
      const result = await api.auth.login(email, password);

      this.token = result.token;
      this.user = result.user;

      if (import.meta.client) {
        localStorage.setItem("auth_token", result.token);
      }

      return result;
    },

    async register(email: string, password: string, name: string) {
      const api = useApi();
      const result = await api.auth.register(email, password, name);

      this.token = result.token;
      this.user = result.user;

      if (import.meta.client) {
        localStorage.setItem("auth_token", result.token);
      }

      return result;
    },

    logout() {
      this.token = null;
      this.user = null;

      if (import.meta.client) {
        localStorage.removeItem("auth_token");
      }

      navigateTo("/login");
    },

    async fetchCurrentUser() {
      if (!this.token && import.meta.client) {
        this.token = localStorage.getItem("auth_token");
      }

      if (!this.token) {
        return;
      }

      try {
        const api = useApi();
        this.user = await api.auth.getCurrentUser();
      } catch (error) {
        this.token = null;
        this.user = null;
        if (import.meta.client) {
          localStorage.removeItem("auth_token");
        }
      }
    },

    initializeFromStorage() {
      if (import.meta.client) {
        const token = localStorage.getItem("auth_token");
        if (token) {
          this.token = token;
        }
      }
    },
  },
});
