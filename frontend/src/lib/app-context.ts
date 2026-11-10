import { createContext } from 'svelte';
import type { DiscoveredArena, User } from './api';
import type { Theme } from './theme';

export type AppState = {
	sessionStatus: 'loading' | 'ready';
	user: User | null;
	theme: Theme;
	discoveredArenas: DiscoveredArena[];
	gameAvailabilityVersion: number;
};

type AppContext = {
	appState: AppState;
	refreshGameAvailability: () => void;
	rememberDiscoveredArenas: (arenas: DiscoveredArena[]) => void;
};

export const [getAppContext, setAppContext] = createContext<AppContext>();
