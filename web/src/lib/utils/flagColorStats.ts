const STORAGE_KEY = 'flag_color_rounds';

export type FlagColorRoundStat = {
	cca2: string;
	deltaE: number;
	pointsEarned: number;
	difficulty: string;
	at: number;
};

/**
 * Persists a single round outcome for future reports / adaptive practice (best-effort).
 */
export function recordFlagColorRound(entry: Omit<FlagColorRoundStat, 'at'>): void {
	if (typeof localStorage === 'undefined') return;
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		const prev: FlagColorRoundStat[] = raw ? JSON.parse(raw) : [];
		prev.push({ ...entry, at: Date.now() });
		localStorage.setItem(STORAGE_KEY, JSON.stringify(prev.slice(-200)));
	} catch {
		/* ignore quota / parse errors */
	}
}
