/**
 * Converts HSV (hue degrees 0–360, saturation/value 0–1) to sRGB bytes.
 */
export function hsvToRgb(h: number, s: number, v: number): { r: number; g: number; b: number } {
	const hh = ((h % 360) + 360) % 360;
	const c = v * s;
	const x = c * (1 - Math.abs(((hh / 60) % 2) - 1));
	const m = v - c;
	let rp = 0;
	let gp = 0;
	let bp = 0;
	if (hh < 60) {
		rp = c;
		gp = x;
	} else if (hh < 120) {
		rp = x;
		gp = c;
	} else if (hh < 180) {
		gp = c;
		bp = x;
	} else if (hh < 240) {
		gp = x;
		bp = c;
	} else if (hh < 300) {
		rp = x;
		bp = c;
	} else {
		rp = c;
		bp = x;
	}
	return {
		r: Math.round((rp + m) * 255),
		g: Math.round((gp + m) * 255),
		b: Math.round((bp + m) * 255),
	};
}

/** Parse #RGB or #RRGGBB to sRGB bytes. */
export function hexToRgb(hex: string): { r: number; g: number; b: number } {
	let h = hex.trim().replace(/^#/, '');
	if (h.length === 3) {
		h = h
			.split('')
			.map((c) => c + c)
			.join('');
	}
	const n = parseInt(h, 16);
	if (h.length !== 6 || Number.isNaN(n)) {
		return { r: 0, g: 0, b: 0 };
	}
	return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
}

/** Human-readable HSB label: H360 S100 B100 (S/B as 0–100). */
export function formatHsbLabel(h: number, s: number, v: number): string {
	const H = Math.round(((h % 360) + 360) % 360);
	const S = Math.round(s * 100);
	const B = Math.round(v * 100);
	return `H${H} S${S} B${B}`;
}

export function rgbToHsv(r: number, g: number, b: number): { h: number; s: number; v: number } {
	const rn = r / 255;
	const gn = g / 255;
	const bn = b / 255;
	const max = Math.max(rn, gn, bn);
	const min = Math.min(rn, gn, bn);
	const d = max - min;
	let h = 0;
	if (d !== 0) {
		if (max === rn) h = 60 * (((gn - bn) / d) % 6);
		else if (max === gn) h = 60 * ((bn - rn) / d + 2);
		else h = 60 * ((rn - gn) / d + 4);
	}
	if (h < 0) h += 360;
	const s = max === 0 ? 0 : d / max;
	return { h, s, v: max };
}
