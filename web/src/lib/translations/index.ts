import { get } from 'svelte/store';
import { locale } from '$lib/stores/locale';
import type { Locale } from '$lib/stores/locale';

// Import all translation files
import en from './locales/en.json';
import es from './locales/es.json';
import fr from './locales/fr.json';
import de from './locales/de.json';
import nl from './locales/nl.json';
import nb from './locales/nb.json';
import da from './locales/da.json';
import sv from './locales/sv.json';
import fi from './locales/fi.json';
import pt from './locales/pt.json';
import tr from './locales/tr.json';
import ro from './locales/ro.json';
import hu from './locales/hu.json';
import hr from './locales/hr.json';
import cs from './locales/cs.json';
import sk from './locales/sk.json';
import pl from './locales/pl.json';
import it from './locales/it.json';
import id from './locales/id.json';
import ms from './locales/ms.json';
import fil from './locales/fil.json';
import sw from './locales/sw.json';
import vi from './locales/vi.json';
import ru from './locales/ru.json';
import zh from './locales/zh.json';
import ko from './locales/ko.json';
import ja from './locales/ja.json';
import ar from './locales/ar.json';
import hi from './locales/hi.json';
import th from './locales/th.json';
import uk from './locales/uk.json';
import he from './locales/he.json';
import el from './locales/el.json';

// Translation dictionary
const translations: Record<string, Record<string, string>> = {
	en,
	es,
	fr,
	de,
	nl,
	nb,
	da,
	sv,
	fi,
	pt,
	tr,
	ro,
	hu,
	hr,
	cs,
	sk,
	pl,
	it,
	id,
	ms,
	fil,
	sw,
	vi,
	ru,
	zh,
	ko,
	ja,
	ar,
	hi,
	th,
	uk,
	he,
	el
};

// Fallback to English if translation is missing
const defaultLocale: Locale = 'en';

/**
 * Get translation string by key
 * Supports template variables like {{.Variable}} and printf-style %d, %s, %.0f%%
 * 
 * @param key - Translation key (e.g., "game.flag.title")
 * @param params - Optional parameters for template substitution
 *   - Object: { Country: "USA", Score: 5 } for {{.Country}} and %s
 *   - Array: [5, 10] for multiple %d, %s placeholders in order
 *   - Single value: 5 or "text" for single placeholder
 * @param currentLocale - Optional locale override (defaults to store locale)
 * @returns Translated string
 */
export function t(
	key: string,
	params?: Record<string, string | number> | Array<string | number> | string | number,
	currentLocale?: Locale
): string {
	const loc = currentLocale || get(locale);
	const translationDict = translations[loc] || translations[defaultLocale];
	
	let translation = translationDict[key] || translations[defaultLocale][key] || key;
	
	// Handle template variables {{.Variable}} or {{Variable}}
	if (params && typeof params === 'object' && !Array.isArray(params)) {
		for (const [paramKey, paramValue] of Object.entries(params)) {
			// Replace {{.Variable}} format
			translation = translation.replace(
				new RegExp(`{{\\.${paramKey}}}`, 'g'),
				String(paramValue)
			);
			// Replace {{Variable}} format (without dot)
			translation = translation.replace(
				new RegExp(`{{${paramKey}}}`, 'g'),
				String(paramValue)
			);
		}
	}
	
	// Handle printf-style formatting (%d, %s, %.0f%%, etc.)
	if (params !== undefined) {
		// If params is an array, replace placeholders in order
		if (Array.isArray(params)) {
			let placeholderIndex = 0;
			translation = translation.replace(/%\.?\d*f?%?[sdif%]/g, () => {
				const value = params[placeholderIndex++];
				return value !== undefined ? String(value) : '';
			});
		}
		// If params is a single value (not an object), use it for first placeholder
		else if (typeof params === 'string' || typeof params === 'number') {
			translation = translation.replace(/%\.?\d*f?%?[sdif%]/, String(params));
		}
		// If params is an object, use values in order (for printf-style)
		else if (typeof params === 'object') {
			let placeholderIndex = 0;
			const values = Object.values(params);
			translation = translation.replace(/%\.?\d*f?%?[sdif%]/g, () => {
				const value = values[placeholderIndex++];
				return value !== undefined ? String(value) : '';
			});
		}
	}
	
	return translation;
}

/**
 * Reactive translation function for Svelte components
 * Automatically updates when locale changes
 * 
 * @param key - Translation key
 * @param params - Optional parameters
 * @returns Reactive translation string
 */
export function $t(
	key: string, 
	params?: Record<string, string | number> | Array<string | number> | string | number
): string {
	// This will be used in Svelte components with $: reactive statements
	// The locale store subscription will trigger re-evaluation
	const loc = get(locale);
	return t(key, params, loc);
}

/**
 * Check if a translation key exists
 */
export function hasTranslation(key: string, currentLocale?: Locale): boolean {
	const loc = currentLocale || get(locale);
	const translationDict = translations[loc] || translations[defaultLocale];
	return key in translationDict;
}

/**
 * Get all available translation keys for a locale
 */
export function getTranslationKeys(currentLocale?: Locale): string[] {
	const loc = currentLocale || get(locale);
	const translationDict = translations[loc] || translations[defaultLocale];
	return Object.keys(translationDict);
}
