/** Latin vowels that look identical to Greek; fold so mixed-script input matches Greek names. */
const latinGreekHomoglyph = new Map<number, number>([
	[0x0061, 0x03b1],
	[0x0065, 0x03b5],
	[0x0069, 0x03b9],
	[0x006f, 0x03bf],
	[0x0075, 0x03c5],
]);

/** NFD then strip all Unicode marks (Mn/Mc/Me), same idea as Go stripUnicodeMarks. */
function stripUnicodeMarks(s: string): string {
	return s.normalize('NFD').replace(/\p{M}/gu, '');
}

function foldLatinGreekHomoglyphs(s: string): string {
	let out = '';
	for (const ch of s) {
		const c = ch.codePointAt(0)!;
		const mapped = latinGreekHomoglyph.get(c);
		out += mapped !== undefined ? String.fromCodePoint(mapped) : ch;
	}
	return out;
}

/**
 * Trims, lowercases (Unicode), strips accents/diacritics via NFD + mark removal, and folds Latin vowel
 * homoglyphs into Greek for answer matching.
 *
 * @param s - Raw user or canonical answer text
 * @returns Normalized string for comparison
 */
export function normalizeAnswerForCompare(s: string): string {
	return foldLatinGreekHomoglyphs(stripUnicodeMarks(s.trim().toLowerCase()));
}
