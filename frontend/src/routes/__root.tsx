import { createRootRoute, Outlet, redirect } from "@tanstack/react-router";
import { sessionAPI } from "@/lib/api";

// A positive probe is cached for the lifetime of the document so SPA
// navigations don't re-hit the network; a fresh page load re-checks. A session
// that expires mid-session is still caught by the global 401 interceptor.
let sessionVerified = false;

async function hasValidSession(): Promise<boolean> {
  if (sessionVerified) {
    return true;
  }

  try {
    await sessionAPI.status();
    sessionVerified = true;
    return true;
  } catch {
    return false;
  }
}

export const Route = createRootRoute({
  beforeLoad: async ({ location }) => {
    if (location.pathname === "/login") {
      return;
    }

    if (!(await hasValidSession())) {
      throw redirect({ to: "/login" });
    }
  },
  component: RootLayout,
});

function RootLayout() {
  return <Outlet />;
}
