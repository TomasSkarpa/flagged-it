import { writable, derived, get } from 'svelte/store';
import { getApiUrl, getWebSocketUrl } from '../api/config';
import type { Writable } from 'svelte/store';

// Types matching backend
export type RoomStatus = 'waiting' | 'playing' | 'finished';
export type GameMode = 'flag' | 'shape' | 'capital' | 'higherlower' | 'facts' | 'worldle';
export type DifficultyLevel = 'easy' | 'medium' | 'hard';

export interface RoomConfig {
	gameMode: GameMode;
	numQuestions: number;
	difficulty: DifficultyLevel;
	timeLimit: number; // seconds, 0 = unlimited
	maxPlayers: number;
	region: string;
	categories: string[];
	isPublic: boolean;
	password?: string;
}

export interface Player {
	id: string;
	name: string;
	isHost: boolean;
	isReady: boolean;
	score: number;
	streak: number;
	joinedAt: string;
	lastAnswer?: AnswerSubmission;
}

export interface AnswerSubmission {
	playerId: string;
	questionId: string;
	answer: string;
	isCorrect: boolean;
	points: number;
	timeTaken: number;
	submittedAt: string;
}

export interface Question {
	id: string;
	type: string;
	data: Record<string, unknown>;
	options?: unknown[];
	timeLimit: number;
	startedAt: string;
}

export interface Room {
	id: string;
	code: string;
	hostId: string;
	config: RoomConfig;
	status: RoomStatus;
	players: Record<string, Player>;
	currentQuestion?: Question;
	questionIndex: number;
	answers?: Record<string, Record<string, AnswerSubmission>>;
	createdAt: string;
	startedAt?: string;
	finishedAt?: string;
}

export interface PublicRoomInfo {
	id: string;
	code: string;
	gameMode: GameMode;
	numQuestions: number;
	difficulty: DifficultyLevel;
	playerCount: number;
	maxPlayers: number;
	status: RoomStatus;
	createdAt: string;
}

export interface WSMessage {
	type: string;
	playerId?: string;
	roomId?: string;
	data?: unknown;
	timestamp: string;
}

export interface ErrorMessage {
	type: 'ERROR';
	message?: string;
	code: string;
}

export interface RoomStateMessage {
	type: 'ROOM_STATE';
	room: Room;
	players: Record<string, Player>;
	timestamp: string;
}

export interface QuestionMessage {
	type: 'QUESTION';
	question: Question;
	index: number;
	total: number;
}

export interface AnswerResultMessage {
	type: 'ANSWER_RESULT';
	playerId: string;
	questionId: string;
	isCorrect: boolean;
	points: number;
	correctAnswer: string;
	leaderboard: Record<string, Player>;
}

export interface GameFinishedMessage {
	type: 'GAME_FINISHED';
	leaderboard: Record<string, Player>;
	finalScores: Record<string, number>;
}

type MultiplayerMessage = RoomStateMessage | QuestionMessage | AnswerResultMessage | GameFinishedMessage | ErrorMessage;

// WebSocket connection state
export const wsConnection = writable<WebSocket | null>(null);
export const wsConnected = writable<boolean>(false);
export const wsError = writable<string | null>(null);

// Current room state
export const currentRoom = writable<Room | null>(null);
export const currentPlayers = writable<Record<string, Player>>({});
export const currentQuestion = writable<Question | null>(null);
export const currentPlayerId = writable<string | null>(null);
export const isHost = derived([currentRoom, currentPlayerId], ([room, playerId]) => {
	return room?.hostId === playerId;
});

// Game state
export const gameLeaderboard = writable<Record<string, Player>>({});
export const gameFinished = writable<boolean>(false);
export const finalScores = writable<Record<string, number>>({});

// Message queue for messages received before handlers are set up
const messageQueue: MultiplayerMessage[] = [];
const messageHandlers = new Map<string, Set<(msg: MultiplayerMessage) => void>>();

export function onMessage(type: string, handler: (msg: MultiplayerMessage) => void) {
	if (!messageHandlers.has(type)) {
		messageHandlers.set(type, new Set());
	}
	messageHandlers.get(type)!.add(handler);
	
	// Process queued messages
	messageQueue.forEach(msg => {
		if (msg.type === type) {
			handler(msg);
		}
	});
	
	return () => {
		messageHandlers.get(type)?.delete(handler);
	};
}

function processMessage(msg: MultiplayerMessage) {
	const handlers = messageHandlers.get(msg.type);
	if (handlers) {
		handlers.forEach(handler => handler(msg));
	} else {
		// Queue message if no handlers yet
		messageQueue.push(msg);
	}
}

// WebSocket connection management
export function connectWebSocket(roomId: string, playerId: string): Promise<WebSocket> {
	// Ensure we're on the client side
	if (typeof window === 'undefined') {
		return Promise.reject(new Error('WebSocket connections can only be established on the client side'));
	}

	return new Promise((resolve, reject) => {
		// Use consistent API URL helper, then convert to WebSocket protocol
		const wsUrl = getWebSocketUrl(`/ws/rooms/${roomId}`);
		const ws = new WebSocket(wsUrl);
		
		ws.onopen = () => {
			wsConnection.set(ws);
			wsConnected.set(true);
			wsError.set(null);
			resolve(ws);
		};
		
		ws.onerror = (error) => {
			wsError.set('Failed to connect to game room');
			reject(error);
		};
		
		ws.onclose = () => {
			wsConnection.set(null);
			wsConnected.set(false);
		};
		
		ws.onmessage = (event) => {
			try {
				// Backend may send multiple JSON messages separated by newlines
				// Split by newline and process each message
				const data = event.data as string;
				const messages = data.split('\n').filter(line => line.trim().length > 0);
				
				for (const messageStr of messages) {
					try {
						const msg = JSON.parse(messageStr) as MultiplayerMessage;
						processMessage(msg);
						
						// Update stores based on message type
						switch (msg.type) {
							case 'ROOM_STATE':
								currentRoom.set(msg.room);
								currentPlayers.set(msg.players);
								if (msg.room.currentQuestion) {
									currentQuestion.set(msg.room.currentQuestion);
								}
								break;
							case 'QUESTION':
								currentQuestion.set(msg.question);
								break;
							case 'ANSWER_RESULT':
								gameLeaderboard.set(msg.leaderboard);
								break;
							case 'GAME_FINISHED':
								gameFinished.set(true);
								gameLeaderboard.set(msg.leaderboard);
								finalScores.set(msg.finalScores);
								break;
							case 'ERROR':
								wsError.set(msg.code);
								break;
						}
					} catch (parseErr) {
						console.error('Failed to parse individual WebSocket message:', parseErr, 'Message:', messageStr);
					}
				}
			} catch (err) {
				console.error('Failed to process WebSocket message:', err, 'Data:', event.data);
			}
		};
	});
}

export function disconnectWebSocket() {
	const ws = get(wsConnection);
	if (ws) {
		ws.close();
		wsConnection.set(null);
		wsConnected.set(false);
	}
}

export function sendMessage(type: string, data: unknown) {
	const ws = get(wsConnection);
	if (!ws) {
		console.error('WebSocket connection is null');
		throw new Error('WebSocket not connected');
	}
	
	if (ws.readyState !== WebSocket.OPEN) {
		console.error('WebSocket not open. State:', ws.readyState, {
			CONNECTING: WebSocket.CONNECTING,
			OPEN: WebSocket.OPEN,
			CLOSING: WebSocket.CLOSING,
			CLOSED: WebSocket.CLOSED
		});
		throw new Error(`WebSocket not open (state: ${ws.readyState})`);
	}
	
	const msg: WSMessage = {
		type,
		data: data as Record<string, unknown>,
		timestamp: new Date().toISOString()
	};
	
	const msgStr = JSON.stringify(msg);
	console.log('Sending WebSocket message:', type, data, 'Full message:', msgStr);
	ws.send(msgStr);
}

// Convenience functions for common actions
export function joinRoom(roomId: string, roomCode: string, playerId: string, playerName: string, password?: string) {
	sendMessage('JOIN', {
		roomId,
		roomCode,
		playerId,
		playerName,
		password
	});
}

export function setReady(playerId: string, ready: boolean) {
	sendMessage('READY', { playerId, ready });
}

export function startGame(playerId: string) {
	sendMessage('START', { playerId });
}

export function submitAnswer(playerId: string, questionId: string, answer: string, timeTaken: number) {
	sendMessage('ANSWER', {
		playerId,
		questionId,
		answer,
		timeTaken
	});
}

export function updateConfig(playerId: string, config: RoomConfig) {
	sendMessage('CONFIG_UPDATE', { playerId, config });
}

// Reset all multiplayer state
export function resetMultiplayerState() {
	currentRoom.set(null);
	currentPlayers.set({});
	currentQuestion.set(null);
	currentPlayerId.set(null);
	gameLeaderboard.set({});
	gameFinished.set(false);
	finalScores.set({});
	wsError.set(null);
	messageQueue.length = 0;
}
