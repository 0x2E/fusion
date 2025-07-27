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

	try {
		const parsed = new URL(url, base);
		return parsed.href;
	} catch {
		if (url.startsWith('//')) {
			try {
				const parsed = new URL('https:' + url, base);
				return parsed.href;
			} catch {
				return url;
			}
		}
		return url;
	}
}
