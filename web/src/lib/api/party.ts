// Multiplayer ("party") API: live rooms and async challenges.
// The player's secret token is kept per browser tab (sessionStorage), so a
// reload keeps your seat and two tabs can be two players.
import { getApiUrl } from './config';

export type PartyMode = 'live' | 'challenge';
export type PartyGame = 'mix' | 'flag' | 'shape' | 'capital' | 'compare' | 'color' | 'fact';
export type PartyKind = Exclude<PartyGame, 'mix'>;
export type PartyPhase = 'lobby' | 'ready' | 'question' | 'reveal' | 'finished';

export interface PartyOption {
	id: string;
	label: string;
	flagUrl?: string;
}

export interface PartyQuestion {
	kind: PartyKind;
	key: string;
	countryName?: string;
	flagUrl?: string;
	fact?: string;
	metric?: 'population' | 'area';
	guessableId?: string;
	options?: PartyOption[];
}

export interface PartyReveal {
	answerId?: string;
	answerLabel?: string;
	countryName?: string;
	cca2?: string;
	targetHex?: string;
	values?: Record<string, number>;
}

export interface PartyAnswer {
	optionId?: string;
	hex?: string;
	correct: boolean;
	points: number;
	timedOut?: boolean;
}

export interface PartyPlayer {
	id: string;
	name: string;
	score: number;
	isHost: boolean;
	isYou: boolean;
	connected: boolean;
	answered: boolean;
	points?: number;
	correct?: boolean;
	done: number;
	finished: boolean;
}

export interface PartyView {
	code: string;
	mode: PartyMode;
	game: PartyGame;
	region: string;
	rounds: number;
	phase: PartyPhase;
	round: number;
	remainingMs: number;
	durationMs: number;
	you: { id: string; name: string; isHost: boolean; score: number; answer?: PartyAnswer };
	players: PartyPlayer[];
	question?: PartyQuestion;
	reveal?: PartyReveal;
	version: number;
}

export interface PartySeat {
	token: string;
	view: PartyView;
}

export class PartyError extends Error {
	constructor(
		message: string,
		public status: number
	) {
		super(message);
	}
}

async function call<T>(path: string, init?: RequestInit): Promise<T> {
	// Only send Content-Type with a body: on a GET it would force a CORS preflight.
	const headers: Record<string, string> = init?.body ? { 'Content-Type': 'application/json' } : {};
	const res = await fetch(getApiUrl(path), { ...init, headers });
	if (!res.ok) {
		let msg = res.statusText;
		try {
			msg = (await res.json()).error || msg;
		} catch {
			/* not JSON */
		}
		throw new PartyError(msg, res.status);
	}
	return res.json();
}

function qs(code: string, token: string, locale: string, extra = ''): string {
	return `?code=${encodeURIComponent(code)}&t=${encodeURIComponent(token)}&locale=${encodeURIComponent(locale)}${extra}`;
}

export function createRoom(
	opts: { mode: PartyMode; game: PartyGame; region: string; name: string },
	locale: string
): Promise<PartySeat> {
	return call(`/party/create?locale=${encodeURIComponent(locale)}`, {
		method: 'POST',
		body: JSON.stringify(opts)
	});
}

export function joinRoom(code: string, name: string, locale: string): Promise<PartySeat> {
	return call(`/party/join?locale=${encodeURIComponent(locale)}`, {
		method: 'POST',
		body: JSON.stringify({ code, name })
	});
}

export function getState(code: string, token: string, locale: string): Promise<PartyView> {
	return call(`/party/state${qs(code, token, locale)}`);
}

export function sendAction(
	code: string,
	token: string,
	locale: string,
	action: 'start' | 'next' | 'again',
	body?: unknown
): Promise<PartyView> {
	return call(`/party/${action}${qs(code, token, locale)}`, {
		method: 'POST',
		body: body ? JSON.stringify(body) : undefined
	});
}

export function sendAnswer(
	code: string,
	token: string,
	locale: string,
	answer: { round: number; optionId?: string; r?: number; g?: number; b?: number }
): Promise<PartyView> {
	return call(`/party/answer${qs(code, token, locale)}`, {
		method: 'POST',
		body: JSON.stringify(answer)
	});
}

export async function leaveRoom(code: string, token: string): Promise<void> {
	try {
		await call(`/party/leave${qs(code, token, 'en')}`, { method: 'POST' });
	} catch {
		/* leaving a room that is already gone is fine */
	}
}

/** Flag image for a flag round, served by the room so the URL doesn't name the country. */
export function partyFlagUrl(code: string, token: string, round: number): string {
	return getApiUrl(`/party/flag${qs(code, token, 'en', `&round=${round}`)}`);
}

export function fetchPartyShape(code: string, token: string, round: number): Promise<any> {
	return call(`/party/shape${qs(code, token, 'en', `&round=${round}`)}`);
}

/**
 * Follow the room live. Uses server-sent events and falls back to polling
 * if the stream keeps failing. Returns a function that stops listening.
 */
export function watchRoom(
	code: string,
	token: string,
	getLocale: () => string,
	onView: (v: PartyView) => void,
	onGone: () => void
): () => void {
	let stopped = false;
	let es: EventSource | null = null;
	let poll: ReturnType<typeof setInterval> | null = null;
	let failures = 0;
	let gotMessage = false;
	let lastSeen = Date.now();

	const startPolling = () => {
		if (poll || stopped) return;
		poll = setInterval(async () => {
			try {
				onView(await getState(code, token, getLocale()));
			} catch (e) {
				if (e instanceof PartyError && (e.status === 403 || e.status === 404)) {
					stop();
					onGone();
				}
			}
		}, 1500);
	};

	const open = () => {
		if (stopped) return;
		es = new EventSource(getApiUrl(`/party/events${qs(code, token, getLocale())}`));
		es.onmessage = (ev) => {
			failures = 0;
			gotMessage = true;
			lastSeen = Date.now();
			try {
				onView(JSON.parse(ev.data));
			} catch {
				/* ignore a bad frame */
			}
		};
		es.addEventListener('ping', () => {
			lastSeen = Date.now();
		});
		es.addEventListener('gone', () => {
			stop();
			onGone();
		});
		es.onerror = () => {
			failures++;
			if (failures >= 3) {
				es?.close();
				es = null;
				startPolling();
			}
		};
	};

	// Watchdog: if a proxy buffers the stream, nothing arrives and nothing
	// errors. The server sends the state at once and pings every 15s, so a
	// quiet stream means we should poll instead.
	const watchdog = setInterval(() => {
		const quiet = Date.now() - lastSeen;
		if (!poll && (quiet > 25000 || (!gotMessage && quiet > 6000))) startPolling();
	}, 2000);

	const stop = () => {
		stopped = true;
		es?.close();
		if (poll) clearInterval(poll);
		clearInterval(watchdog);
	};

	open();
	return stop;
}

const seatKey = (code: string) => `party-seat-${code.toUpperCase()}`;

export function saveSeat(code: string, token: string) {
	try {
		sessionStorage.setItem(seatKey(code), token);
	} catch {
		/* private mode */
	}
}

export function loadSeat(code: string): string | null {
	try {
		return sessionStorage.getItem(seatKey(code));
	} catch {
		return null;
	}
}

export function clearSeat(code: string) {
	try {
		sessionStorage.removeItem(seatKey(code));
	} catch {
		/* private mode */
	}
}

const NAME_KEY = 'party-name';

export function loadName(): string {
	try {
		return localStorage.getItem(NAME_KEY) || '';
	} catch {
		return '';
	}
}

export function saveName(name: string) {
	try {
		localStorage.setItem(NAME_KEY, name);
	} catch {
		/* private mode */
	}
}
