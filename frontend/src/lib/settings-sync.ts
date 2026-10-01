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

  // Classify before the first await: a failed GET must not defer this to a
  // later load, after user edits have created the blob and turned this
  // browser's defaults into "user" values. Nothing else can write the blob
  // yet — beforeLoad has not rendered any UI.
  try {
    if (readPreferencesOrigin() === null) {
      markPreferencesOrigin(
        localStorage.getItem("fusion-preferences") !== null ? "user" : "sync",
      );
    }
  } catch {
    // Storage failure: classification retries on the next page load.
  }

  let settings: Settings | undefined;
  try {
    settings = (await settingsAPI.get()).data;
  } catch {
    return;
  }
  if (!settings) {
    return;
  }

  try {
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
      // Load the catalog first so the switch does not flash English. A failed
      // catalog load keeps the local locale; the sync retries on the next
      // page load and must not fail the navigation.
      await ensureLocaleMessages(settings.locale);
      usePreferencesStore.getState().setLocale(settings.locale);
    }
  } catch {
    // Storage failures (quota exceeded, read-only storage) and anything else
    // after a successful GET must keep local values and not fail the route;
    // the sync retries on the next page load.
  }
}

// A server null means the user never saved a preference. Only a blob that
// already existed when sync first ran counts as user choices: one written by
// applying server values holds defaults this browser never chose. The theme
// key only ever holds a value next-themes wrote for an explicit pick (a
// failed save deliberately keeps the pick instead of writing the implicit
// default back), so its presence alone qualifies the value.
function seedUnsetSettings(settings: Settings): void {
  const seed: UpdateSettingsRequest = {};

  if (readPreferencesOrigin() === "user") {
    const preferences = usePreferencesStore.getState();
    if (settings.locale === null) {
      seed.locale = preferences.locale;
    }
    if (settings.article_page_size === null) {
      seed.article_page_size = preferences.articlePageSize;
    }
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
