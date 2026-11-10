<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import PartySetup from '$lib/components/party/PartySetup.svelte';
	import PartyJoin from '$lib/components/party/PartyJoin.svelte';
	import PartyLobby from '$lib/components/party/PartyLobby.svelte';
	import PartyChallengeReady from '$lib/components/party/PartyChallengeReady.svelte';
	import PartyRound from '$lib/components/party/PartyRound.svelte';
	import PartyFinished from '$lib/components/party/PartyFinished.svelte';
	import {
		createRoom,
		joinRoom,
		getState,
		sendAction,
		sendAnswer,
		leaveRoom,
		watchRoom,
		saveSeat,
		loadSeat,
		clearSeat,
		loadName,
		saveName,
		PartyError,
		type PartyView,
		type PartyGame,
		type PartyMode
	} from '$lib/api/party';

	type Screen = 'loading' | 'setup' | 'join' | 'room';

	let screen: Screen = 'loading';
	let code = '';
	let token = '';
	let name = '';
	let view: PartyView | null = null;
	let busy = false;
	let error: string | null = null;
	let stopWatching: (() => void) | null = null;
	let watchedLocale = '';

	$: L = $locale;
	$: shareUrl = typeof window !== 'undefined' && code ? `${window.location.origin}/party?code=${code}` : '';

	function errorText(e: unknown): string {
		if (e instanceof PartyError) {
			if (e.status === 404 || e.status === 403) return t('party.error.not_found', undefined, L);
			if (e.message === 'room is full') return t('party.error.full', undefined, L);
		}
		return t('party.error.generic', undefined, L);
	}

	// Keep the newest view: action replies and the live stream can cross.
	function applyView(v: PartyView) {
		if (view && v.code === view.code && v.version < view.version) return;
		// New screen (game starts, next round in a challenge, results): start at the top.
		if (view && (v.phase !== view.phase || v.round !== view.round) && v.phase !== 'reveal') {
			window.scrollTo({ top: 0, behavior: 'smooth' });
		}
		view = v;
	}

	function watch() {
		stopWatching?.();
		watchedLocale = L;
		stopWatching = watchRoom(code, token, () => L, applyView, () => {
			clearSeat(code);
			error = t('party.error.not_found', undefined, L);
			toSetup();
		});
	}

	// Reopen the stream in the new language so names are translated.
	$: if (screen === 'room' && token && L !== watchedLocale) watch();

	async function setUrl(c: string) {
		await goto(c ? `/party?code=${c}` : '/party', { replaceState: true, keepFocus: true, noScroll: true });
	}

	function enterRoom(seatToken: string, v: PartyView) {
		code = v.code;
		token = seatToken;
		saveSeat(code, token);
		view = v;
		screen = 'room';
		error = null;
		window.scrollTo({ top: 0 });
		setUrl(code);
		watch();
	}

	onMount(async () => {
		name = loadName();
		const c = new URLSearchParams(window.location.search).get('code')?.toUpperCase() ?? '';
		if (!c) {
			screen = 'setup';
			return;
		}
		code = c;
		const saved = loadSeat(c);
		if (!saved) {
			screen = 'join';
			return;
		}
		try {
			enterRoom(saved, await getState(c, saved, L));
		} catch (e) {
			clearSeat(c);
			if (e instanceof PartyError && e.status === 404) {
				error = errorText(e);
				toSetup();
			} else {
				screen = 'join';
			}
		}
	});

	onDestroy(() => stopWatching?.());

	function toSetup() {
		stopWatching?.();
		stopWatching = null;
		view = null;
		token = '';
		code = '';
		screen = 'setup';
		setUrl('');
	}

	async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
		if (busy) return;
		busy = true;
		error = null;
		try {
			return await fn();
		} catch (e) {
			error = errorText(e);
		} finally {
			busy = false;
		}
	}

	async function handleCreate(e: CustomEvent<{ mode: PartyMode; game: PartyGame; region: string; name: string }>) {
		saveName(e.detail.name);
		name = e.detail.name;
		const seat = await run(() => createRoom(e.detail, L));
		if (!seat) return;
		enterRoom(seat.token, seat.view);
		// The creator of a challenge plays straight away, then shares the link.
		if (seat.view.mode === 'challenge') await handleAction('start');
	}

	async function handleJoin(e: CustomEvent<{ code: string; name: string }>) {
		saveName(e.detail.name);
		name = e.detail.name;
		const seat = await run(() => joinRoom(e.detail.code, e.detail.name, L));
		if (seat) enterRoom(seat.token, seat.view);
	}

	async function handleAction(action: 'start' | 'next' | 'again') {
		const v = await run(() => sendAction(code, token, L, action));
		if (v) applyView(v);
	}

	async function handleAnswer(e: CustomEvent<{ optionId?: string; r?: number; g?: number; b?: number }>) {
		if (!view) return;
		const round = view.round;
		const v = await run(() => sendAnswer(code, token, L, { round, ...e.detail }));
		if (v) applyView(v);
	}

	async function handleLeave() {
		const c = code;
		const tk = token;
		clearSeat(c);
		toSetup();
		await leaveRoom(c, tk);
	}
</script>

<svelte:head>
	<title>{t('party.title', undefined, L)} - Flagged It</title>
	<meta name="description" content={t('home.game.party.description', undefined, L)} />
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if screen === 'loading'}
			<div class="flex justify-center py-20"><LoadingSpinner /></div>
		{:else if screen === 'setup'}
			<PartySetup bind:name isLoading={busy} {error} on:create={handleCreate} on:join={handleJoin} />
		{:else if screen === 'join'}
			<PartyJoin {code} bind:name isLoading={busy} {error} on:join={handleJoin} on:back={toSetup} />
		{:else if view}
			{#if view.phase === 'lobby'}
				<PartyLobby {view} {shareUrl} isLoading={busy} on:start={() => handleAction('start')} on:leave={handleLeave} />
			{:else if view.phase === 'ready'}
				<PartyChallengeReady {view} isLoading={busy} on:start={() => handleAction('start')} />
			{:else if (view.phase === 'question' || view.phase === 'reveal') && view.question}
				<PartyRound {view} {code} {token} isSubmitting={busy} on:answer={handleAnswer} on:next={() => handleAction('next')} />
			{:else if view.phase === 'finished'}
				<PartyFinished {view} {shareUrl} isLoading={busy} on:again={() => handleAction('again')} on:leave={handleLeave} on:newGame={toSetup} />
			{/if}
			{#if error}
				<div class="mt-6 p-4 bg-error/20 border border-error rounded-lg max-w-md mx-auto text-center">
					<p class="text-error font-semibold">{error}</p>
				</div>
			{/if}
		{/if}
	</div>
</div>
