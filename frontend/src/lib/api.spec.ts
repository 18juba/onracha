import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from './api';

const user = {
	id: 1,
	name: 'Lucas',
	email: 'lucas@example.test',
	created_at: '2026-11-08T12:00:00Z'
};

const response = (status: number, body: unknown) =>
	({
		ok: status >= 200 && status < 300,
		status,
		json: vi.fn().mockResolvedValue(body)
	}) as unknown as Response;

const fetchMock = vi.fn();

describe('API client', () => {
	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('sends credentials when checking the current session', async () => {
		fetchMock.mockResolvedValue(response(200, user));

		await expect(api.me()).resolves.toEqual(user);

		expect(fetchMock).toHaveBeenCalledWith(
			expect.stringMatching(/\/api\/v1\/auth\/me$/),
			expect.objectContaining({ credentials: 'include' })
		);
	});

	it('sends registration and login data as JSON', async () => {
		fetchMock.mockResolvedValueOnce(response(201, user)).mockResolvedValueOnce(response(200, user));

		await api.register({ name: 'Lucas', email: user.email, password: 'Senha123' });
		await api.login({ email: user.email, password: 'Senha123' });

		expect(fetchMock).toHaveBeenNthCalledWith(
			1,
			expect.stringMatching(/\/api\/v1\/auth\/register$/),
			expect.objectContaining({
				method: 'POST',
				credentials: 'include',
				body: JSON.stringify({ name: 'Lucas', email: user.email, password: 'Senha123' })
			})
		);
		expect(fetchMock).toHaveBeenNthCalledWith(
			2,
			expect.stringMatching(/\/api\/v1\/auth\/login$/),
			expect.objectContaining({
				method: 'POST',
				credentials: 'include',
				body: JSON.stringify({ email: user.email, password: 'Senha123' })
			})
		);
	});

	it('exposes API errors with their HTTP status and message', async () => {
		fetchMock.mockResolvedValue(
			response(401, {
				error: { code: 'invalid_credentials', message: 'Email ou senha inválidos.' }
			})
		);

		await expect(api.me()).rejects.toMatchObject({
			status: 401,
			code: 'invalid_credentials',
			message: 'Email ou senha inválidos.'
		});
	});
});
