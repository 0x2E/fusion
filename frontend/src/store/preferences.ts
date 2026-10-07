import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

export const supportedLocales = [
  "en",
  "zh",
  "de",
  "fr",
  "es",
  "ru",
  "pt",
  "sv",
] as const;
export type AppLocale = (typeof supportedLocales)[number];

export const articlePageSizeOptions = [10, 20, 30, 50, 100] as const;
export type ArticlePageSize = (typeof articlePageSizeOptions)[number];

// Closed set mirrored by the backend handler's allow-list. "open" marks an
// article read the moment it is opened; the numeric strings are delays in
// seconds. "off" is a real value (not null): null on the server means "never
// chosen" and the client falls back to "off".
export const autoMarkReadOptions = ["off", "open", "5", "10", "30"] as const;
export type AutoMarkRead = (typeof autoMarkReadOptions)[number];

const localeSet = new Set<AppLocale>(supportedLocales);
const articlePageSizeSet = new Set<number>(articlePageSizeOptions);
const autoMarkReadSet = new Set<AutoMarkRead>(autoMarkReadOptions);

function resolveSupportedLocale(locale: string): AppLocale | null {
  const normalized = locale.toLowerCase().replace("_", "-");

  if (localeSet.has(normalized as AppLocale)) {
    return normalized as AppLocale;
  }

  const languageCode = normalized.split("-")[0];
  if (localeSet.has(languageCode as AppLocale)) {
    return languageCode as AppLocale;
  }

  return null;
}

const defaultLocale: AppLocale =
  (typeof navigator !== "undefined" &&
    resolveSupportedLocale(navigator.language)) ||
  "en";
const defaultArticlePageSize: ArticlePageSize = 10;
const defaultAutoMarkRead: AutoMarkRead = "off";

function normalizeLocale(locale: string): AppLocale {
  const supportedLocale = resolveSupportedLocale(locale);
  if (supportedLocale) {
    return supportedLocale;
  }

  return defaultLocale;
}

function normalizeArticlePageSize(size: number): ArticlePageSize {
  if (articlePageSizeSet.has(size)) {
    return size as ArticlePageSize;
  }

  return defaultArticlePageSize;
}

// Accepts undefined (key missing from an older persisted blob) and any
// garbage; both fall back to "off". A truthiness check would be wrong here:
// every non-"off" member is a meaningful value.
function normalizeAutoMarkRead(value: string | undefined | null): AutoMarkRead {
  return value !== null && autoMarkReadSet.has(value as AutoMarkRead)
    ? (value as AutoMarkRead)
    : defaultAutoMarkRead;
}

// "off" disables the feature; every other member maps to a delay where 0
// means "as soon as the article is open".
export function autoMarkReadDelayMs(option: AutoMarkRead): number | null {
  if (option === "off") {
    return null;
  }
  return option === "open" ? 0 : Number.parseInt(option, 10) * 1000;
}

export interface PreferencesState {
  locale: AppLocale;
  articlePageSize: ArticlePageSize;
  autoMarkRead: AutoMarkRead;
  setLocale: (locale: string) => void;
  setArticlePageSize: (size: number) => void;
  setAutoMarkRead: (option: string) => void;
}

export const usePreferencesStore = create<PreferencesState>()(
  persist(
    (set) => ({
      locale: defaultLocale,
      articlePageSize: defaultArticlePageSize,
      autoMarkRead: defaultAutoMarkRead,
      setLocale: (locale) => set({ locale: normalizeLocale(locale) }),
      setArticlePageSize: (size) =>
        set({ articlePageSize: normalizeArticlePageSize(size) }),
      setAutoMarkRead: (option) =>
        set({ autoMarkRead: normalizeAutoMarkRead(option) }),
    }),
    {
      name: "fusion-preferences",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        locale: state.locale,
        articlePageSize: state.articlePageSize,
        autoMarkRead: state.autoMarkRead,
      }),
      merge: (persistedState, currentState) => {
        const persisted = persistedState as Partial<PreferencesState> | undefined;

        return {
          ...currentState,
          locale: normalizeLocale(persisted?.locale ?? currentState.locale),
          articlePageSize: normalizeArticlePageSize(
            persisted?.articlePageSize ?? currentState.articlePageSize,
          ),
          autoMarkRead: normalizeAutoMarkRead(
            persisted?.autoMarkRead ?? currentState.autoMarkRead,
          ),
        };
      },
    },
  ),
);

export function getPreferredLocale(): AppLocale {
  return usePreferencesStore.getState().locale;
}

export function isSupportedLocale(locale: string): locale is AppLocale {
  return localeSet.has(locale as AppLocale);
}

export function isArticlePageSize(size: number): size is ArticlePageSize {
  return articlePageSizeSet.has(size);
}

export function isAutoMarkRead(value: string): value is AutoMarkRead {
  return autoMarkReadSet.has(value as AutoMarkRead);
}
