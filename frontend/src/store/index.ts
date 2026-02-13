import { setUnauthorizedCallback } from "@/lib/api";

export { useUIStore } from "./ui";
export { useArticleSessionStore } from "./article-session";
export {
  articlePageSizeOptions,
  supportedLocales,
  usePreferencesStore,
} from "./preferences";

// Setup 401 handler - redirect to login on unauthorized
setUnauthorizedCallback(() => {
  window.location.href = "/login";
});
