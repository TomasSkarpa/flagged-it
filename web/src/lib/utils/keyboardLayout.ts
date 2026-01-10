import type { KeyboardLayout } from '$lib/components/ui/Keyboard.svelte';

/**
 * Maps locale codes to keyboard layouts
 * Falls back to 'english' if no mapping exists
 * 
 * Note: Some keyboard layouts (Arabic, Greek, Hebrew, Thai) don't have corresponding
 * translations yet, but are kept for future use or for users who want to type in those scripts.
 */
export function getKeyboardLayoutForLocale(loc: string): KeyboardLayout {
	const localeToLayout: Record<string, KeyboardLayout> = {
		'en': 'english',
		'es': 'spanish',
		'fr': 'english', // French uses QWERTY with accents (no special layout needed)
		'de': 'english', // German uses QWERTZ but we use QWERTY for simplicity
		'nl': 'english',
		'nb': 'english', // Norwegian Bokmål
		'da': 'english',
		'sv': 'english',
		'fi': 'english',
		'pt': 'english', // Portuguese uses QWERTY with accents
		'tr': 'turkish',
		'ro': 'english', // Romanian uses QWERTZ but we use QWERTY
		'hu': 'english', // Hungarian uses QWERTZ
		'hr': 'english', // Croatian
		'cs': 'czech',
		'sk': 'english', // Slovak uses QWERTZ
		'pl': 'polish',
		'it': 'english',
		'id': 'indonesian',
		'ms': 'english', // Malay uses QWERTY
		'fil': 'english', // Filipino uses QWERTY
		'sw': 'english', // Swahili uses QWERTY
		'vi': 'vietnamese',
		'ru': 'russian',
		'zh': 'english', // Chinese uses Pinyin input method with QWERTY layout
		'ko': 'korean',
		'ja': 'japanese',
		'ar': 'arabic',
		'hi': 'english', // Hindi uses QWERTY with Devanagari input method
		'th': 'thai',
		'uk': 'russian', // Ukrainian uses Cyrillic keyboard similar to Russian
		'he': 'hebrew',
		'el': 'greek' // Greek keyboard layout
	};
	return localeToLayout[loc] || 'english';
}
