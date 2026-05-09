import { describe, expect, it } from 'vitest';
import { normalizeAnswerForCompare } from './answerNormalize';

describe('normalizeAnswerForCompare', () => {
	const cases: [string, string][] = [
		['Κόσοβο', 'κοσοβο'],
		['Κοσοβο', 'κοσοβο'],
		['Κoσοβο', 'κοσοβο'],
		['κόσοβο', 'κοσοβο'],
		['  Αθήνα ', 'αθηνα'],
		['Αθήνα', 'αθηνα'],
		['United States', 'υnιtεd stαtεs'],
		['jizni korea', 'jιznι kοrεα'],
		['Jižní Korea', 'jιznι kοrεα'],
		['JIZNI KOREA', 'jιznι kοrεα'],
		['réunion', 'rευnιοn'],
		['RÉUNION', 'rευnιοn']
	];

	for (const [input, want] of cases) {
		it(`${JSON.stringify(input)} → ${JSON.stringify(want)}`, () => {
			expect(normalizeAnswerForCompare(input)).toBe(want);
		});
	}
});
