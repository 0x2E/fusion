import type { AppLocale } from "@/store/preferences";
import { enMessages, type TranslationKey } from "./en";
import type { PartialMessages } from "./types";

type LoadedMessages = Record<TranslationKey, string>;
type NonEnglishLocale = Exclude<AppLocale, "en">;
type LocaleModule = Partial<Record<string, PartialMessages>>;

const fallbackMessages: LoadedMessages = { ...enMessages };

const localeExportNames: Record<NonEnglishLocale, string> = {
  zh: "zhMessages",
  de: "deMessages",
  fr: "frMessages",
  es: "esMessages",
  ru: "ruMessages",
  pt: "ptMessages",
  sv: "svMessages",
};

const localeLoaders: Record<NonEnglishLocale, () => Promise<LocaleModule>> = {
  zh: () => import("./zh"),
  de: () => import("./de"),
  fr: () => import("./fr"),
  es: () => import("./es"),
  ru: () => import("./ru"),
  pt: () => import("./pt"),
  sv: () => import("./sv"),
};

const cachedMessages: Partial<Record<AppLocale, LoadedMessages>> = {
  en: fallbackMessages,
};
const loadingPromises: Partial<Record<NonEnglishLocale, Promise<void>>> = {};
const listeners = new Set<() => void>();

function notifyLocaleMessagesUpdated(): void {
  for (const listener of listeners) {
    listener();
  }
}

function extractLocaleMessages(
  locale: NonEnglishLocale,
  localeModule: LocaleModule,
): PartialMessages {
  const exportName = localeExportNames[locale];
  const messages = localeModule[exportName];

  if (!messages) {
    throw new Error(`Missing export ${exportName} in locale module ${locale}`);
  }

  return messages;
}

async function loadLocaleMessages(locale: NonEnglishLocale): Promise<void> {
  if (cachedMessages[locale]) {
    return;
  }

  const localeModule = await localeLoaders[locale]();
  const localeMessages = extractLocaleMessages(locale, localeModule);
  cachedMessages[locale] = {
    ...fallbackMessages,
    ...localeMessages,
  };
  notifyLocaleMessagesUpdated();
}

export function getLocaleMessages(locale: AppLocale): LoadedMessages {
  return cachedMessages[locale] ?? fallbackMessages;
}

export function subscribeLocaleMessagesUpdated(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export async function ensureLocaleMessages(locale: AppLocale): Promise<void> {
  if (locale === "en" || cachedMessages[locale]) {
    return;
  }

  const nonEnglishLocale = locale as NonEnglishLocale;
  const currentPromise = loadingPromises[nonEnglishLocale];
  if (currentPromise) {
    await currentPromise;
    return;
  }

  const promise = loadLocaleMessages(nonEnglishLocale).finally(() => {
    delete loadingPromises[nonEnglishLocale];
  });
  loadingPromises[nonEnglishLocale] = promise;
  await promise;
}

export const localeLabels: Record<AppLocale, string> = {
  en: "English",
  zh: "简体中文",
  de: "Deutsch",
  fr: "Français",
  es: "Español",
  ru: "Русский",
  pt: "Português",
  sv: "Svenska",
};

export type { TranslationKey };
