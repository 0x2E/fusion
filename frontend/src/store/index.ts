import { setUnauthorizedCallback } from "@/lib/api";

export { useUIStore, type ArticleFilter } from "./ui";
export { useDataStore } from "./data";

// Setup 401 handler - redirect to login on unauthorized
setUnauthorizedCallback(() => {
  window.location.href = "/login";
});
