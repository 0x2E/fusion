import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

export const supportedLocales = ["en", "zh"] as const;
export type AppLocale = (typeof supportedLocales)[number];

export const articlePageSizeOptions = [10, 20, 30, 50, 100] as const;
export type ArticlePageSize = (typeof articlePageSizeOptions)[number];

const defaultLocale: AppLocale = "en";
const defaultArticlePageSize: ArticlePageSize = 10;

const localeSet = new Set<AppLocale>(supportedLocales);
const articlePageSizeSet = new Set<number>(articlePageSizeOptions);

function normalizeLocale(locale: string): AppLocale {
  if (localeSet.has(locale as AppLocale)) {
    return locale as AppLocale;
  }

  return defaultLocale;
}

function normalizeArticlePageSize(size: number): ArticlePageSize {
  if (articlePageSizeSet.has(size)) {
    return size as ArticlePageSize;
  }

  return defaultArticlePageSize;
}

export interface PreferencesState {
  locale: AppLocale;
  articlePageSize: ArticlePageSize;
  setLocale: (locale: string) => void;
  setArticlePageSize: (size: number) => void;
}

export const usePreferencesStore = create<PreferencesState>()(
  persist(
    (set) => ({
      locale: defaultLocale,
      articlePageSize: defaultArticlePageSize,
      setLocale: (locale) => set({ locale: normalizeLocale(locale) }),
      setArticlePageSize: (size) =>
        set({ articlePageSize: normalizeArticlePageSize(size) }),
    }),
    {
      name: "fusion-preferences",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        locale: state.locale,
        articlePageSize: state.articlePageSize,
      }),
      merge: (persistedState, currentState) => {
        const persisted = persistedState as Partial<PreferencesState> | undefined;

        return {
          ...currentState,
          locale: normalizeLocale(persisted?.locale ?? currentState.locale),
          articlePageSize: normalizeArticlePageSize(
            persisted?.articlePageSize ?? currentState.articlePageSize,
          ),
        };
      },
    },
  ),
);

export function getPreferredLocale(): AppLocale {
  return usePreferencesStore.getState().locale;
}
