import { t } from '$lib/translations';

/**
 * Must stay aligned with internal/games/flagcolor.deltaPerfect. Raw perceptual ΔE
 * must be at or below this for the “Perfect” tier label.
 */
export const FLAG_COLOR_DELTA_PERFECT = 5;

const TIER_DELTA_EXCELLENT = 14;
const TIER_DELTA_GREAT = 26;
const TIER_DELTA_GOOD = 42;
const TIER_DELTA_PRACTICE = 58;

/**
 * Feedback tier from raw CIE76 ΔE only (matches on-screen “perceptual” distance),
 * independent of hue-relaxed scoring points.
 */
export function flagColorTierFromRawDeltaE(deltaE: number, loc: string): string {
	if (deltaE <= FLAG_COLOR_DELTA_PERFECT) return t('game.flag_color.tier.perfect', undefined, loc);
	if (deltaE <= TIER_DELTA_EXCELLENT) return t('game.flag_color.tier.excellent', undefined, loc);
	if (deltaE <= TIER_DELTA_GREAT) return t('game.flag_color.tier.great', undefined, loc);
	if (deltaE <= TIER_DELTA_GOOD) return t('game.flag_color.tier.good', undefined, loc);
	if (deltaE <= TIER_DELTA_PRACTICE) return t('game.flag_color.tier.practice', undefined, loc);
	return t('game.flag_color.tier.try_again', undefined, loc);
}

/**
 * Maps linear points (0..maxPoints) to a 0.00–10.00 display string per round.
 */
export function pointsToTenDisplay(pointsEarned: number, maxPointsPerRound: number): string {
	if (maxPointsPerRound <= 0) return '0.00';
	const v = (pointsEarned / maxPointsPerRound) * 10;
	return Math.min(10, Math.max(0, v)).toFixed(2);
}

export function parseTenDisplay(s: string): number {
	const n = parseFloat(s);
	return Number.isFinite(n) ? n : 0;
}
