import ca from './langs/ca';
import de from './langs/de';
import en from './langs/en';
import es from './langs/es';
import fr from './langs/fr';
import pt from './langs/pt';
import ptBR from './langs/pt-BR';
import ru from './langs/ru';
import sv from './langs/sv';
import zhHans from './langs/zh-Hans';
import zhHant from './langs/zh-Hant';

export type TranslationKey = keyof typeof en;

export type Translation = {
	[key in TranslationKey]: string;
};

export const languages = [
	{ id: 'en', name: 'English', translation: en },
	{ id: 'zh-Hans', name: '简体中文', translation: zhHans },
	{ id: 'zh-Hant', name: '繁體中文', translation: zhHant },
	{ id: 'ca', name: 'Català', translation: ca },
	{ id: 'de', name: 'Deutsch', translation: de },
	{ id: 'es', name: 'Español', translation: es },
	{ id: 'fr', name: 'Français', translation: fr },
	{ id: 'sv', name: 'Svenska', translation: sv },
	{ id: 'ru', name: 'Русский', translation: ru },
	{ id: 'pt', name: 'Português', translation: pt },
	{ id: 'pt-BR', name: 'Português do Brasil', translation: ptBR }
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
	const browserLang = navigator.language;

	// Chinese
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

	// Catalan
	if (browserLang.startsWith('ca')) {
		return 'ca';
	}
	
	// French
	if (browserLang.startsWith('fr')) {
		return 'fr';
	}

	// Spanish
	if (browserLang.startsWith('es')) {
		return 'es';
	}

	// Swedish
	if (browserLang.startsWith('sv')) {
		return 'sv';
	}

	// Russian
	if (browserLang.startsWith('ru')) {
		return 'ru';
	}

	// German
	if (browserLang.startsWith('de')) {
		return 'de';
	}

	// Portuguese
	if (browserLang.startsWith('pt')) {
		if (browserLang === 'pt-BR') {
			return 'pt-BR';
		}
		return 'pt';
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

	if (!text) {
		text = key;
	}

	return text;
}

export function setLanguage(langId: Language): void {
	localStorage.setItem(LANGUAGE_STORAGE_KEY, langId);
}

export function getAvailableLanguages(): { id: Language; name: string }[] {
	return languages.map(({ id, name }) => ({ id, name }));
}
