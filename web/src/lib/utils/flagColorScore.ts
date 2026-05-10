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
