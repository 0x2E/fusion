import { useEffect, useRef } from "react";
import { useTheme } from "next-themes";
import { settingsAPI, type Settings, type UpdateSettingsRequest } from "@/lib/api";
import { ensureLocaleMessages } from "@/lib/i18n/messages";
import {
  isArticlePageSize,
  isSupportedLocale,
  usePreferencesStore,
} from "@/store/preferences";

const THEMES = ["light", "dark", "system"] as const;

// The remote settings pull runs once per document, after the session probe.
// Server values win over the localStorage cache (which stays the first-paint /
// offline fallback), so preferences follow the user across devices and
// browsers (#167). A null server value means "never chosen": keep local. A
// failed pull keeps local values and is not retried until the next page load.
let settingsPulled = false;

// Sign-out does a full page navigation, which discards module state anyway;
// resetting here keeps the pull honest even if that ever changes.
export function resetSettingsPull() {
  settingsPulled = false;
}

// The persisted preferences blob is written by both user choices and this
// sync (any store write persists both keys). The marker records which side
// owns it: only a user-owned blob may seed the server, so values that merely
// defaulted on this browser are never promoted to explicit choices on a
// later load. Evaluated once, when the pull first runs.
const PREFERENCES_ORIGIN_KEY = "fusion-preferences-origin";

function readPreferencesOrigin(): "user" | "sync" | null {
  const value = localStorage.getItem(PREFERENCES_ORIGIN_KEY);
  return value === "user" || value === "sync" ? value : null;
}

function markPreferencesOrigin(origin: "user" | "sync") {
  localStorage.setItem(PREFERENCES_ORIGIN_KEY, origin);
}

// Called when the user explicitly changes locale or page size, making the
// current blob values eligible for seeding.
export function markPreferencesAsUserOwned() {
  markPreferencesOrigin("user");
}

// A first-load seed may still be in flight when the user changes a setting;
// user writes wait for it so the seed cannot land on top of them.
let seedPromise: Promise<unknown> = Promise.resolve();

export function waitForSettingsSeed(): Promise<unknown> {
  return seedPromise;
}

// beforeLoad runs outside the ThemeProvider, so the theme application waits
// for a mounted listener under it. The pending value is one-shot: next-themes
// recreates its setTheme callback on every theme change, so re-running an
// effect that re-applies a stale pending value would fight the user's own
// theme picks and cross-tab storage sync.
let pendingRemoteTheme: string | null = null;
const themeListeners = new Set<(theme: string) => void>();

function dispatchRemoteTheme(theme: string) {
  if (themeListeners.size > 0) {
    for (const listener of themeListeners) {
      listener(theme);
    }
    return;
  }
  pendingRemoteTheme = theme;
}

export function useRemoteThemeSync() {
  const setThemeRef = useRef(useTheme().setTheme);
  setThemeRef.current = useTheme().setTheme;

  useEffect(() => {
    if (pendingRemoteTheme !== null) {
      const theme = pendingRemoteTheme;
      pendingRemoteTheme = null;
      setThemeRef.current(theme);
    }
    const listener = (theme: string) => setThemeRef.current(theme);
    themeListeners.add(listener);
    return () => {
      themeListeners.delete(listener);
    };
  }, []);
}

export async function pullRemoteSettings(): Promise<void> {
  if (settingsPulled) {
    return;
  }
  settingsPulled = true;

  let settings: Settings | undefined;
  try {
    settings = (await settingsAPI.get()).data;
  } catch {
    return;
  }
  if (!settings) {
    return;
  }

  if (readPreferencesOrigin() === null) {
    markPreferencesOrigin(
      localStorage.getItem("fusion-preferences") !== null ? "user" : "sync",
    );
  }

  seedUnsetSettings(settings);

  const preferences = usePreferencesStore.getState();
  if (
    settings.article_page_size !== null &&
    isArticlePageSize(settings.article_page_size) &&
    settings.article_page_size !== preferences.articlePageSize
  ) {
    usePreferencesStore
      .getState()
      .setArticlePageSize(settings.article_page_size);
  }

  const storedTheme = localStorage.getItem("theme");
  if (
    settings.theme !== null &&
    (THEMES as readonly string[]).includes(settings.theme) &&
    settings.theme !== storedTheme
  ) {
    dispatchRemoteTheme(settings.theme);
  }

  if (
    settings.locale !== null &&
    isSupportedLocale(settings.locale) &&
    settings.locale !== preferences.locale
  ) {
    try {
      // Load the catalog first so the switch does not flash English. A failed
      // catalog load keeps the local locale; the sync retries on the next
      // page load and must not fail the navigation.
      await ensureLocaleMessages(settings.locale);
      usePreferencesStore.getState().setLocale(settings.locale);
    } catch {
      // keep local locale
    }
  }
}

// A server null means the user never saved a preference. A user-owned local
// blob predates (or was chosen after) the backend sync, so upload its values
// once; values this browser only defaulted to stay null server-side.
function seedUnsetSettings(settings: Settings): void {
  if (readPreferencesOrigin() !== "user") {
    return;
  }

  const seed: UpdateSettingsRequest = {};
  const preferences = usePreferencesStore.getState();
  if (settings.locale === null) {
    seed.locale = preferences.locale;
  }
  if (settings.article_page_size === null) {
    seed.article_page_size = preferences.articlePageSize;
  }
  const storedTheme = localStorage.getItem("theme");
  if (
    settings.theme === null &&
    storedTheme !== null &&
    (THEMES as readonly string[]).includes(storedTheme)
  ) {
    seed.theme = storedTheme;
  }

  if (Object.keys(seed).length === 0) {
    return;
  }
  // Best effort: a failed seed simply retries on the next page load.
  seedPromise = settingsAPI.update(seed).catch(() => undefined);
}
