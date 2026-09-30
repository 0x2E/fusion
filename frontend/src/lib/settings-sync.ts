import { useEffect } from "react";
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

// beforeLoad runs outside the ThemeProvider, so the theme application waits
// for a mounted listener under it. RootLayout is mounted for the whole
// document (including /login), and the pull can complete either before or
// after its first mount, so both paths must work.
let pendingRemoteTheme: string | null = null;
const themeListeners = new Set<(theme: string) => void>();

function queueRemoteTheme(theme: string) {
  pendingRemoteTheme = theme;
  for (const listener of themeListeners) {
    listener(theme);
  }
}

// Applies a remotely-pulled theme once a listener is (or becomes) available.
// Values are restricted to next-themes' preference names; resolvedTheme is
// never used, so "system" is not frozen to the OS value at sync time.
export function useRemoteThemeSync() {
  const { setTheme } = useTheme();

  useEffect(() => {
    if (pendingRemoteTheme) {
      setTheme(pendingRemoteTheme);
    }
    const listener = (theme: string) => setTheme(theme);
    themeListeners.add(listener);
    return () => {
      themeListeners.delete(listener);
    };
  }, [setTheme]);
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

  seedUnsetSettings(settings);

  const preferences = usePreferencesStore.getState();
  if (
    settings.locale !== null &&
    isSupportedLocale(settings.locale) &&
    settings.locale !== preferences.locale
  ) {
    // Load the catalog first so the switch does not flash English.
    await ensureLocaleMessages(settings.locale);
    usePreferencesStore.getState().setLocale(settings.locale);
  }
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
    queueRemoteTheme(settings.theme);
  }
}

// A server null means the user never saved a preference. Existing local
// choices predate the backend sync, so upload them once; values this browser
// only defaulted to (no stored key) stay null server-side.
function seedUnsetSettings(settings: Settings): void {
  const seed: UpdateSettingsRequest = {};

  const preferences = usePreferencesStore.getState();
  const hasStoredPreferences = localStorage.getItem("fusion-preferences") !== null;
  if (settings.locale === null && hasStoredPreferences) {
    seed.locale = preferences.locale;
  }
  if (settings.article_page_size === null && hasStoredPreferences) {
    seed.article_page_size = preferences.articlePageSize;
  }
  const storedTheme = localStorage.getItem("theme");
  if (settings.theme === null && storedTheme !== null) {
    seed.theme = storedTheme;
  }

  if (Object.keys(seed).length === 0) {
    return;
  }
  // Best effort: a failed seed simply retries on the next page load.
  settingsAPI.update(seed).catch(() => {});
}
