import { getApiUrl } from './config';
import type { Room, RoomConfig, PublicRoomInfo } from '../stores/multiplayer';

export interface CreateRoomRequest {
	hostName: string;
	config: RoomConfig;
}

export interface CreateRoomResponse {
	roomId: string;
	roomCode: string;
	hostId: string;
	shareUrl: string;
	room: Room;
}

/**
 * Create a new multiplayer room
 */
export async function createRoom(request: CreateRoomRequest): Promise<CreateRoomResponse> {
	const response = await fetch(getApiUrl('/rooms'), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(request),
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Failed to create room');
	}

	return response.json();
}

/**
 * Get room information by ID
 */
export async function getRoom(roomId: string): Promise<Room> {
	const response = await fetch(getApiUrl(`/rooms/${roomId}`), {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json',
		},
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Room not found');
	}

	return response.json();
}

/**
 * Get room information by code
 */
export async function getRoomByCode(code: string): Promise<Room> {
	const response = await fetch(getApiUrl(`/rooms/code/${code}`), {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json',
		},
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Room not found');
	}

	return response.json();
}

/**
 * Get list of public rooms
 */
export async function getPublicRooms(): Promise<PublicRoomInfo[]> {
	const response = await fetch(getApiUrl('/rooms/public'), {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json',
		},
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Failed to fetch public rooms');
	}

	const data = await response.json();
	return data.rooms || [];
}

/**
 * Delete a room (host only)
 */
export async function deleteRoom(roomId: string): Promise<void> {
	const response = await fetch(getApiUrl(`/rooms/${roomId}`), {
		method: 'DELETE',
		headers: {
			'Content-Type': 'application/json',
		},
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Failed to delete room');
	}
}
