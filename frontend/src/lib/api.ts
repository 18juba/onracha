export type User = {
	id: number;
	name: string;
	email: string;
	created_at: string;
};

export type Arena = {
	id: number;
	owner_user_id: number;
	name: string;
	arena_type: string;
	address: string;
	city: string;
	state: string;
	latitude: number | null;
	longitude: number | null;
	photo_url: string | null;
	active: boolean;
	created_at: string;
};

export type DiscoveredArena = Pick<
	Arena,
	| 'id'
	| 'name'
	| 'arena_type'
	| 'address'
	| 'city'
	| 'state'
	| 'latitude'
	| 'longitude'
	| 'photo_url'
>;

export type CreateArenaInput = {
	name: string;
	arena_type: string;
	address: string;
	city: string;
	state: string;
	latitude: number;
	longitude: number;
};

export type UpdateArenaInput = Omit<CreateArenaInput, 'latitude' | 'longitude'> & {
	latitude: number | null;
	longitude: number | null;
	active: boolean;
};

export type TimeSlot = {
	id: number;
	arena_id: number;
	starts_at: string;
	ends_at: string;
	created_at: string;
	closed_at: string | null;
};

export type ManagedTimeSlot = TimeSlot & {
	game_id: number | null;
};

export type CreateTimeSlotInput = {
	starts_at: string;
	ends_at: string;
};

export type Game = {
	id: number;
	time_slot_id: number;
	arena_id: number;
	organizer_user_id: number;
	starts_at: string;
	ends_at: string;
	participant_user_ids: number[];
	created_at: string;
};

export type GameSearchResult = {
	game: Game;
	arena: Pick<
		DiscoveredArena,
		'id' | 'name' | 'address' | 'city' | 'state' | 'latitude' | 'longitude' | 'photo_url'
	>;
};

export type GameParticipant = {
	id: number;
	name: string;
};

export type GameLobby = {
	game: Game;
	arena: GameSearchResult['arena'];
	participants: GameParticipant[];
};

export class ApiError extends Error {
	constructor(
		message: string,
		readonly status: number,
		readonly code: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}

type ApiErrorResponse = {
	error?: {
		code?: string;
		message?: string;
	};
};

const apiBaseURL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '');

export function resolveAssetURL(value: string | null): string | null {
	if (!value) return null;

	try {
		const apiURL = new URL(apiBaseURL);
		const assetURL = new URL(value, apiURL);
		const localHosts = ['localhost', '127.0.0.1', '::1'];

		if (localHosts.includes(assetURL.hostname) && !localHosts.includes(apiURL.hostname)) {
			assetURL.protocol = apiURL.protocol;
			assetURL.host = apiURL.host;
		}

		return assetURL.href;
	} catch {
		return value;
	}
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const headers = new Headers(init.headers);
	const isFormData = typeof FormData !== 'undefined' && init.body instanceof FormData;
	if (init.body && !isFormData && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	const response = await fetch(`${apiBaseURL}${path}`, {
		...init,
		headers,
		credentials: 'include'
	});

	if (response.status === 204) {
		return undefined as T;
	}

	let body: unknown;
	try {
		body = await response.json();
	} catch {
		body = null;
	}

	if (!response.ok) {
		const error = (body as ApiErrorResponse | null)?.error;
		throw new ApiError(
			error?.message ?? 'Não foi possível concluir a solicitação.',
			response.status,
			error?.code ?? 'request_failed'
		);
	}

	return body as T;
}

export const api = {
	register(input: { name: string; email: string; password: string }): Promise<User> {
		return request('/api/v1/auth/register', {
			method: 'POST',
			body: JSON.stringify(input)
		});
	},

	login(input: { email: string; password: string }): Promise<User> {
		return request('/api/v1/auth/login', {
			method: 'POST',
			body: JSON.stringify(input)
		});
	},

	me(): Promise<User> {
		return request('/api/v1/auth/me');
	},

	listArenas(): Promise<Arena[]> {
		return request('/api/v1/arenas');
	},

	discoverArenas(input: {
		city?: string;
		latitude?: number;
		longitude?: number;
		radius_km?: number;
	}): Promise<DiscoveredArena[]> {
		const query = new URLSearchParams();
		if (input.city) query.set('city', input.city);
		if (input.latitude !== undefined) query.set('latitude', String(input.latitude));
		if (input.longitude !== undefined) query.set('longitude', String(input.longitude));
		if (input.radius_km !== undefined) query.set('radius_km', String(input.radius_km));

		return request(`/api/v1/arenas/discover?${query.toString()}`);
	},

	createArena(input: CreateArenaInput): Promise<Arena> {
		return request('/api/v1/arenas', {
			method: 'POST',
			body: JSON.stringify(input)
		});
	},

	updateArena(arenaID: number, input: UpdateArenaInput): Promise<Arena> {
		return request(`/api/v1/arenas/${arenaID}`, {
			method: 'PUT',
			body: JSON.stringify(input)
		});
	},

	toggleArenaActive(arenaID: number): Promise<Arena> {
		return request(`/api/v1/arenas/${arenaID}/active`, { method: 'PATCH' });
	},

	listArenaTimeSlots(arenaID: number): Promise<TimeSlot[]> {
		return request(`/api/v1/arenas/${arenaID}/time-slots`);
	},

	listManagedArenaTimeSlots(arenaID: number): Promise<ManagedTimeSlot[]> {
		return request(`/api/v1/arenas/${arenaID}/time-slots/manage`);
	},

	closeArenaTimeSlot(arenaID: number, timeSlotID: number): Promise<TimeSlot> {
		return request(`/api/v1/arenas/${arenaID}/time-slots/${timeSlotID}/close`, { method: 'PATCH' });
	},

	listMyGames(): Promise<Game[]> {
		return request('/api/v1/me/games');
	},

	listAvailableGames(city?: string): Promise<GameSearchResult[]> {
		const query = new URLSearchParams();
		if (city) query.set('city', city);
		const queryString = query.toString();
		const suffix = queryString ? `?${queryString}` : '';
		return request(`/api/v1/games${suffix}`);
	},

	getGameLobby(gameID: number): Promise<GameLobby> {
		return request(`/api/v1/games/${gameID}/lobby`);
	},

	createArenaTimeSlot(arenaID: number, input: CreateTimeSlotInput): Promise<TimeSlot> {
		return request(`/api/v1/arenas/${arenaID}/time-slots`, {
			method: 'POST',
			body: JSON.stringify(input)
		});
	},

	scheduleGame(timeSlotID: number): Promise<Game> {
		return request(`/api/v1/time-slots/${timeSlotID}/games`, { method: 'POST' });
	},

	joinGame(gameID: number): Promise<Game> {
		return request(`/api/v1/games/${gameID}/join`, { method: 'POST' });
	},

	replaceArenaPhoto(arenaID: number, file: File): Promise<Arena> {
		const formData = new FormData();
		formData.append('file', file);

		return request(`/api/v1/arenas/${arenaID}/photo`, {
			method: 'PUT',
			body: formData
		});
	}
};
