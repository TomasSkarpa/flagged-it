import { writable } from 'svelte/store';

export type Locale = string;

// Languages with estimated number of active speakers (millions)
// Estimates based on native + second language speakers worldwide (roughly by CCA/population)
const LOCALES_WITH_SPEAKERS = [
	{ code: 'en', name: 'English', speakers: 1500 },      // ~1.5B (most widely spoken)
	{ code: 'zh', name: '中文', speakers: 1100 },         // ~1.1B (Mandarin Chinese)
	{ code: 'hi', name: 'हिन्दी', speakers: 600 },         // ~600M (Hindi)
	{ code: 'es', name: 'Español', speakers: 550 },       // ~550M (Spanish)
	{ code: 'ar', name: 'العربية', speakers: 422 },       // ~422M (Arabic, RTL)
	{ code: 'ms', name: 'Bahasa Melayu', speakers: 290 }, // ~290M (Malay/Indonesian variants)
	{ code: 'fr', name: 'Français', speakers: 280 },      // ~280M (French)
	{ code: 'pt', name: 'Português', speakers: 260 },     // ~260M (Portuguese)
	{ code: 'ru', name: 'Русский', speakers: 258 },       // ~258M (Russian)
	{ code: 'sw', name: 'Kiswahili', speakers: 200 },     // ~200M (mostly L2 speakers)
	{ code: 'id', name: 'Bahasa Indonesia', speakers: 200 }, // ~200M (Indonesian)
	{ code: 'de', name: 'Deutsch', speakers: 130 },       // ~130M (German)
	{ code: 'ja', name: '日本語', speakers: 125 },         // ~125M (Japanese)
	{ code: 'fil', name: 'Filipino', speakers: 110 },     // ~110M (Filipino/Tagalog)
	{ code: 'vi', name: 'Tiếng Việt', speakers: 85 },     // ~85M (Vietnamese)
	{ code: 'it', name: 'Italiano', speakers: 85 },       // ~85M (Italian)
	{ code: 'tr', name: 'Türkçe', speakers: 80 },         // ~80M (Turkish)
	{ code: 'ko', name: '한국어', speakers: 77 },          // ~77M (Korean)
	{ code: 'th', name: 'ไทย', speakers: 60 },             // ~60M (Thai)
	{ code: 'pl', name: 'Polski', speakers: 45 },         // ~45M (Polish)
	{ code: 'uk', name: 'Українська', speakers: 45 },     // ~45M (Ukrainian)
	{ code: 'nl', name: 'Nederlands', speakers: 24 },     // ~24M (Dutch)
	{ code: 'ro', name: 'Română', speakers: 24 },         // ~24M (Romanian)
	{ code: 'hu', name: 'Magyar', speakers: 13 },         // ~13M (Hungarian)
	{ code: 'el', name: 'Ελληνικά', speakers: 13 },       // ~13M (Greek)
	{ code: 'cs', name: 'Čeština', speakers: 10 },        // ~10M (Czech)
	{ code: 'sv', name: 'Svenska', speakers: 10 },        // ~10M (Swedish)
	{ code: 'he', name: 'עברית', speakers: 9 },            // ~9M (Hebrew, RTL)
	{ code: 'el', name: 'Ελληνικά', speakers: 13 },       // ~13M (Greek)
	{ code: 'da', name: 'Dansk', speakers: 6 },           // ~6M (Danish)
	{ code: 'hr', name: 'Hrvatski', speakers: 6 },        // ~6M (Croatian)
	{ code: 'sk', name: 'Slovenčina', speakers: 5 },      // ~5M (Slovak)
	{ code: 'nb', name: 'Norsk Bokmål', speakers: 5 },    // ~5M (Norwegian)
	{ code: 'fi', name: 'Suomi', speakers: 5 }            // ~5M (Finnish)
];

// Sort by speaker count (descending) - most active speakers first
const SUPPORTED_LOCALES = LOCALES_WITH_SPEAKERS.sort((a, b) => b.speakers - a.speakers);

function getStoredLocale(): Locale {
	if (typeof window === 'undefined') return 'en';
	return localStorage.getItem('locale') || getBrowserLocale();
}

function getBrowserLocale(): Locale {
	if (typeof window === 'undefined') return 'en';
	const nav = navigator as Navigator & { userLanguage?: string };
	const browserLang = (nav.language || nav.userLanguage || 'en').split('-')[0];
	// Check if browser language is supported, otherwise default to English
	return SUPPORTED_LOCALES.find(loc => loc.code === browserLang)?.code || 'en';
}

// RTL languages
const RTL_LOCALES = ['ar', 'he'];

function createLocaleStore() {
	const { subscribe, set, update } = writable<Locale>(getStoredLocale());

	return {
		subscribe,
		set: (locale: Locale) => {
			if (typeof window !== 'undefined') {
				localStorage.setItem('locale', locale);
				document.documentElement.lang = locale;
				// Set RTL direction for RTL languages
				if (RTL_LOCALES.includes(locale)) {
					document.documentElement.dir = 'rtl';
				} else {
					document.documentElement.dir = 'ltr';
				}
			}
			set(locale);
		},
		getSupportedLocales: () => SUPPORTED_LOCALES,
		isRTL: (locale: Locale) => RTL_LOCALES.includes(locale)
	};
}

export const locale = createLocaleStore();


