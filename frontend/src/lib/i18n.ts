import { useCallback, useEffect, useState } from "react";
import {
  getPreferredLocale,
  type AppLocale,
  usePreferencesStore,
} from "@/store/preferences";
import {
  ensureLocaleMessages,
  getLocaleMessages,
  localeLabels,
  subscribeLocaleMessagesUpdated,
  type TranslationKey,
} from "@/lib/i18n/messages";

type TranslationParams = Record<string, string | number>;

export { localeLabels, type TranslationKey };

function formatMessage(template: string, params?: TranslationParams): string {
  if (!params) {
    return template;
  }

  return template.replace(/\{(\w+)\}/g, (match, key: string) => {
    const value = params[key];
    return value === undefined ? match : String(value);
  });
}

export function translate(
  key: TranslationKey,
  params?: TranslationParams,
  locale: AppLocale = getPreferredLocale(),
): string {
  if (locale !== "en") {
    void ensureLocaleMessages(locale);
  }

  const localeMessages = getLocaleMessages(locale);
  const fallbackMessages = getLocaleMessages("en");
  return formatMessage(localeMessages[key] ?? fallbackMessages[key], params);
}

export function preloadLocaleMessages(locale: AppLocale = getPreferredLocale()) {
  return ensureLocaleMessages(locale);
}

export function useI18n() {
  const locale = usePreferencesStore((state) => state.locale);

  const [, setMessageVersion] = useState(0);
  useEffect(
    () =>
      subscribeLocaleMessagesUpdated(() => {
        setMessageVersion((version) => version + 1);
      }),
    [],
  );

  useEffect(() => {
    void ensureLocaleMessages(locale);
  }, [locale]);

  const t = useCallback(
    (key: TranslationKey, params?: TranslationParams) =>
      translate(key, params, locale),
    [locale],
  );

  return {
    locale,
    t,
  };
}
