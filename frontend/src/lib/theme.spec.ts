import { describe, expect, it, vi } from 'vitest';
import { applyTheme, readThemePreference, toggleTheme, type Theme } from './theme';

function createStorage(value: string | null = null) {
	let storedValue = value;
	return {
		getItem: vi.fn(() => storedValue),
		setItem: vi.fn((_key: string, nextValue: string) => {
			storedValue = nextValue;
		}),
		get storedValue() {
			return storedValue;
		}
	};
}

function createRoot() {
	return { dataset: {} as { theme?: string } };
}

describe('theme utilities', () => {
	it('defaults to dark when there is no saved preference', () => {
		expect(readThemePreference(createStorage())).toBe('dark');
	});

	it.each(['dark', 'light'] as Theme[])('reads a saved %s preference', (theme) => {
		expect(readThemePreference(createStorage(theme))).toBe(theme);
	});

	it('ignores an invalid saved preference', () => {
		expect(readThemePreference(createStorage('system'))).toBe('dark');
	});

	it('applies the selected theme to the document root', () => {
		const root = createRoot();
		const storage = createStorage();

		applyTheme('light', root);

		expect(root.dataset.theme).toBe('light');
		expect(storage.setItem).not.toHaveBeenCalled();
	});

	it('toggles between dark and light and returns the new theme', () => {
		const root = createRoot();
		const storage = createStorage();

		const nextTheme = toggleTheme('dark', root, storage);

		expect(nextTheme).toBe('light');
		expect(root.dataset.theme).toBe('light');
		expect(storage.storedValue).toBe('light');
	});
});
