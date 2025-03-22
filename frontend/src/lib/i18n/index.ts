import en from './langs/en';
import zhHans from './langs/zh-Hans';
import zhHant from './langs/zh-Hant';

export type TranslationKey = keyof typeof en;

export type Translation = {
	[key in TranslationKey]: string;
};

export const languages = [
	{ id: 'en', name: 'English', translation: en },
	{ id: 'zh-Hans', name: '简体中文', translation: zhHans },
	{ id: 'zh-Hant', name: '繁體中文', translation: zhHant }
] as const;

export type Language = (typeof languages)[number]['id'];

const LANGUAGE_STORAGE_KEY = 'app_language';

export function getCurrentLanguage(): Language {
	// get the language from localStorage
	const savedLanguage = localStorage.getItem(LANGUAGE_STORAGE_KEY) as Language | null;
	if (savedLanguage && languages.some((lang) => lang.id === savedLanguage)) {
		return savedLanguage;
	}

	// get the language from the browser

	// Chinese
	const browserLang = navigator.language;
	if (browserLang.startsWith('zh')) {
		if (browserLang.includes('Hans') || browserLang === 'zh-CN') {
			return 'zh-Hans';
		}
		if (browserLang.includes('Hant') || browserLang === 'zh-TW' || browserLang === 'zh-HK') {
			return 'zh-Hant';
		}
		// fallback
		return 'zh-Hans';
	}

	// fallback
	return 'en';
}

export function getTranslation(): Translation {
	const currentLang = getCurrentLanguage();
	const langObj = languages.find((lang) => lang.id === currentLang);
	return langObj?.translation || en;
}

export function t(key: TranslationKey, params?: Record<string, any>): string {
	const translations = getTranslation();
	let text = translations[key] ?? en[key];

	if (params) {
		Object.entries(params).forEach(([paramKey, value]) => {
			text = text.replace(new RegExp(`{${paramKey}}`, 'g'), String(value));
		});
	}

	return text;
}

export function setLanguage(langId: Language): void {
	localStorage.setItem(LANGUAGE_STORAGE_KEY, langId);
}

export function getAvailableLanguages() {
	return languages.map(({ id, name }) => ({ id, name }));
}
