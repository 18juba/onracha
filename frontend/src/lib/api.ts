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

export type CreateArenaInput = {
	name: string;
	arena_type: string;
	address: string;
	city: string;
	state: string;
	latitude: number;
	longitude: number;
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

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const headers = new Headers(init.headers);
	if (init.body && !headers.has('Content-Type')) {
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

	createArena(input: CreateArenaInput): Promise<Arena> {
		return request('/api/v1/arenas', {
			method: 'POST',
			body: JSON.stringify(input)
		});
	}
};
