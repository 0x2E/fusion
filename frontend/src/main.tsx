import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import { QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { routeTree } from "./routeTree.gen";
import { Toaster } from "@/components/ui/sonner";
import { setUnauthorizedCallback } from "@/lib/api";
import { preloadLocaleMessages } from "@/lib/i18n";
import { queryClient } from "@/lib/query-client";
import { registerPWA } from "@/lib/pwa";
import { usePreferencesStore } from "@/store";
import "./index.css";

// Create a new router instance
const router = createRouter({ routeTree });

registerPWA();
setUnauthorizedCallback(() => {
  window.location.href = "/login";
});

const initialLocale = usePreferencesStore.getState().locale;
document.documentElement.lang = initialLocale;
void preloadLocaleMessages(initialLocale);
usePreferencesStore.subscribe((state) => {
  document.documentElement.lang = state.locale;
  void preloadLocaleMessages(state.locale);
});

// Register the router instance for type safety
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
        <RouterProvider router={router} />
        <Toaster position="top-center" />
      </ThemeProvider>
    </QueryClientProvider>
  </StrictMode>,
);
