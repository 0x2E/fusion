import { setUnauthorizedCallback } from "../lib/api";
import { useAuthStore } from "./auth";

// Initialize 401 handler
setUnauthorizedCallback(() => {
  useAuthStore.getState().setUnauthenticated();
});

export { useAuthStore } from "./auth";
export { useDataStore } from "./data";
export { useUIStore } from "./ui";
