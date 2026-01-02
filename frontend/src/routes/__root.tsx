import { createRootRoute, Outlet, redirect } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/router-devtools";
import { useAuthStore } from "@/store";

export const Route = createRootRoute({
  beforeLoad: async ({ location }) => {
    // Allow access to login page
    if (location.pathname === "/login") {
      return;
    }

    // Check authentication for all other routes
    await useAuthStore.getState().checkAuth();
    const { isAuthenticated } = useAuthStore.getState();

    if (!isAuthenticated) {
      throw redirect({
        to: "/login",
        search: {
          redirect: location.pathname,
        },
      });
    }
  },
  component: () => (
    <>
      <Outlet />
      <TanStackRouterDevtools />
    </>
  ),
});
