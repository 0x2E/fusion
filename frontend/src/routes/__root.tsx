import { createRootRoute, Outlet, redirect } from "@tanstack/react-router";
import { APIError, sessionAPI } from "@/lib/api";

// A positive probe is cached for the lifetime of the document so SPA
// navigations don't re-hit the network; a fresh page load re-checks. A session
// that expires mid-session is still caught by the global 401 interceptor.
let sessionVerified = false;

// Sign-out must clear the cached verdict: it lives in the same document as the
// SPA navigation to /login, and a stale "verified" would wave later
// navigations through without re-checking.
export function resetSessionGuard() {
  sessionVerified = false;
}

async function hasValidSession(): Promise<boolean> {
  if (sessionVerified) {
    return true;
  }

  try {
    await sessionAPI.status();
    sessionVerified = true;
    return true;
  } catch (error) {
    // Only an explicit 401 means "not signed in". Server errors and network
    // failures leave the user where they are; the global 401 interceptor
    // still handles a session that turns out to be invalid.
    return error instanceof APIError && error.status !== 401;
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
