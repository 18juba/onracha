import { describe, expect, it } from 'vitest';
import { ApiError } from './api';
import { getUserFacingErrorMessage } from './errors';

describe('getUserFacingErrorMessage', () => {
	it('returns the message supplied by the API', () => {
		const error = new ApiError('Email já cadastrado.', 409, 'email_taken');

		expect(getUserFacingErrorMessage(error, 'Tente novamente.')).toBe('Email já cadastrado.');
	});

	it('returns the fallback for an unexpected error', () => {
		expect(getUserFacingErrorMessage(new Error('network error'), 'Tente novamente.')).toBe(
			'Tente novamente.'
		);
	});

	it('returns the fallback for non-error values', () => {
		expect(getUserFacingErrorMessage(null, 'Tente novamente.')).toBe('Tente novamente.');
	});
});
