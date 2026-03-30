/**
 * Which culturally themed 404 layout to show (copy + art direction).
 */
export type Error404Variant =
	| 'fr'
	| 'us'
	| 'ja'
	| 'ca'
	| 'it'
	| 'gb'
	| 'de'
	| 'au'
	| 'ru'
	| 'es'
	| 'ch'
	| 'br'
	| 'ie'
	| 'nordic';

const VARIANTS: Error404Variant[] = [
	'fr',
	'us',
	'ja',
	'ca',
	'it',
	'gb',
	'de',
	'au',
	'ru',
	'es',
	'ch',
	'br',
	'ie',
	'nordic'
];

/**
 * Chooses a 404 theme at random (uniform among all variants).
 *
 * @returns {Error404Variant} Variant id used for copy and illustration paths.
 */
export function pickRandom404Variant(): Error404Variant {
	return VARIANTS[Math.floor(Math.random() * VARIANTS.length)];
}

/**
 * Static asset path for the 404 illustration (placeholder SVGs until final art).
 *
 * @param {Error404Variant} variant - Theme id from {@link pickRandom404Variant}.
 * @returns {string} URL path under `/static` for use in `<img src>`.
 */
export function get404IllustrationSrc(variant: Error404Variant): string {
	if (variant === 'fr') return '/assets/errors/surrender-flag.svg';
	return `/assets/errors/404-${variant}.svg`;
}
