/**
 * Game utility functions
 */

/**
 * Calculates the current round number based on total rounds completed and maximum rounds.
 * Ensures the round counter never exceeds the maximum (e.g., shows "10/10" not "11/10").
 * 
 * @param total - Number of rounds completed so far
 * @param totalRounds - Maximum number of rounds in the game
 * @returns Current round number (capped at totalRounds)
 */
export function calculateCurrentRound(total: number, totalRounds: number): number {
	return Math.min(total + 1, totalRounds);
}
