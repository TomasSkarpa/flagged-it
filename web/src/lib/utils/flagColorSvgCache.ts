/** Shared SVG text cache for flag-color (survives FlagColorFlag remounts). */
const svgTextByUrl = new Map<string, string>();

export function getCachedFlagSvgText(flagUrl: string): string | undefined {
	return svgTextByUrl.get(flagUrl);
}

export function setCachedFlagSvgText(flagUrl: string, text: string): void {
	svgTextByUrl.set(flagUrl, text);
}
