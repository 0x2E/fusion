import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function debounce(func: Function, wait: number): EventListener {
	let timeout: ReturnType<typeof setTimeout>;

	return function (this: HTMLElement, event: Event) {
		const context = this;

		const later = () => {
			func.apply(context, [event]);
		};

		clearTimeout(timeout);
		timeout = setTimeout(later, wait);
	};
}

export function tryAbsURL(url: string, base?: string): string {
	if (!url) return url;

	let parsed = URL.parse(url, base);
	if (!parsed) {
		if (url.startsWith('//')) {
			url = 'https:' + url;
		}
		parsed = URL.parse(url, base);
	}

	return parsed?.href || url;
}
