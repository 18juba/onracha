export type Theme = 'dark' | 'light';

const THEME_STORAGE_KEY = 'onracha-theme';

type ThemeStorage = Pick<Storage, 'getItem' | 'setItem'>;
type ThemeRoot = { dataset: Record<string, string | undefined> };

export function readThemePreference(storage: Pick<Storage, 'getItem'>): Theme {
	const savedTheme = storage.getItem(THEME_STORAGE_KEY);
	return savedTheme === 'light' ? 'light' : 'dark';
}

export function applyTheme(theme: Theme, root: ThemeRoot): void {
	root.dataset.theme = theme;
}

export function toggleTheme(currentTheme: Theme, root: ThemeRoot, storage: ThemeStorage): Theme {
	const nextTheme = currentTheme === 'dark' ? 'light' : 'dark';
	applyTheme(nextTheme, root);
	storage.setItem(THEME_STORAGE_KEY, nextTheme);
	return nextTheme;
}
