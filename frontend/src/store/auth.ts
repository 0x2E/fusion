import { toast } from "sonner";
import { create } from "zustand";
import { sessionAPI } from "../lib/api";

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  setUnauthenticated: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  isLoading: false,

  login: async (password: string) => {
    set({ isLoading: true });
    try {
      await sessionAPI.login({ password });
      set({ isAuthenticated: true, isLoading: false });
      toast.success("Login successful");
    } catch (error) {
      set({ isLoading: false });
      toast.error(error instanceof Error ? error.message : "Login failed");
      throw error;
    }
  },

  logout: async () => {
    set({ isLoading: true });
    try {
      await sessionAPI.logout();
      set({ isAuthenticated: false, isLoading: false });
      toast.success("Logged out successfully");
    } catch (error) {
      set({ isLoading: false });
      toast.error(error instanceof Error ? error.message : "Logout failed");
      throw error;
    }
  },

  checkAuth: async () => {
    set({ isLoading: true });
    try {
      // Try to fetch groups to verify auth - if 401, we're not authenticated
      const response = await fetch("/api/groups", { credentials: "include" });
      set({
        isAuthenticated: response.ok,
        isLoading: false,
      });
    } catch (error) {
      set({ isAuthenticated: false, isLoading: false });
    }
  },

  setUnauthenticated: () => {
    set({ isAuthenticated: false });
  },
}));
