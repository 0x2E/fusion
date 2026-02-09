import { setUnauthorizedCallback } from "@/lib/api";

export { useUIStore } from "./ui";

// Setup 401 handler - redirect to login on unauthorized
setUnauthorizedCallback(() => {
  window.location.href = "/login";
});
