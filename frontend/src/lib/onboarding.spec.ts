import { describe, expect, it } from 'vitest';
import { getChoiceScreen, passwordsMatch } from './onboarding';

describe('onboarding flow', () => {
	it('starts the email step for an existing account', () => {
		expect(getChoiceScreen(true)).toBe('email');
	});

	it('starts registration for a new account', () => {
		expect(getChoiceScreen(false)).toBe('register');
	});

	it('accepts a matching password confirmation', () => {
		expect(passwordsMatch('racha2026', 'racha2026')).toBe(true);
	});

	it('rejects a mismatched password confirmation', () => {
		expect(passwordsMatch('racha2026', 'outra2026')).toBe(false);
	});
});
