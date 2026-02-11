import { useEffect, useState } from "react";
import { createLazyFileRoute, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { oidcAPI, sessionAPI } from "@/lib/api";
import { defaultArticleFilter } from "@/lib/article-filter";

export const Route = createLazyFileRoute("/login")({
  component: LoginPage,
});

function LoginPage() {
  const navigate = useNavigate();
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [oidcEnabled, setOidcEnabled] = useState(false);
  const [oidcLoading, setOidcLoading] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("error") === "oidc_failed") {
      toast.error("OIDC login failed. Please try again.");
      window.history.replaceState({}, "", "/login");
    }

    oidcAPI
      .status()
      .then((res) => {
        if (res.data?.enabled) {
          setOidcEnabled(true);
        }
      })
      .catch(() => {
        // Keep password login as fallback when OIDC status is unavailable.
      });
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setIsLoading(true);
    try {
      await sessionAPI.login({ password });
      navigate({
        to: "/$filter",
        params: { filter: defaultArticleFilter },
      });
    } catch {
      toast.error("Invalid password");
    } finally {
      setIsLoading(false);
    }
  };

  const handleOIDCLogin = async () => {
    setOidcLoading(true);
    try {
      const res = await oidcAPI.login();
      if (res.data?.auth_url) {
        window.location.href = res.data.auth_url;
      }
    } catch {
      toast.error("Failed to start OIDC login");
      setOidcLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-sm space-y-6 p-4">
        <div className="flex flex-col items-center gap-2">
          <img
            src="/icon-96.png"
            alt="Fusion logo"
            className="h-12 w-12 rounded-xl"
          />
          <h1 className="text-2xl font-bold">Fusion</h1>
          <p className="text-sm text-muted-foreground">
            Enter your password to continue
          </p>
        </div>

        {oidcEnabled && (
          <>
            <Button
              type="button"
              variant="outline"
              className="w-full"
              onClick={handleOIDCLogin}
              disabled={oidcLoading}
            >
              {oidcLoading ? "Redirecting..." : "Sign in with OIDC"}
            </Button>
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  or
                </span>
              </div>
            </div>
          </>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={isLoading}
            autoFocus
          />
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "Signing in..." : "Sign in"}
          </Button>
        </form>
      </div>
    </div>
  );
}
