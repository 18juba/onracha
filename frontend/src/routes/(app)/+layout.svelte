<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { LoaderCircle } from '@lucide/svelte';
	import { api, type DiscoveredArena } from '../../lib/api';
	import { setAppContext, type AppState } from '../../lib/app-context';
	import BottomNavigation from '../../lib/components/BottomNavigation.svelte';
	import { applyTheme, readThemePreference } from '../../lib/theme';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
	let appState = $state<AppState>({
		sessionStatus: 'loading',
		user: null,
		theme: 'dark',
		discoveredArenas: [],
		gameAvailabilityVersion: 0
	});

	setAppContext({
		appState,
		refreshGameAvailability() {
			appState.gameAvailabilityVersion += 1;
		},
		rememberDiscoveredArenas(arenas: DiscoveredArena[]) {
			const knownArenas = [...appState.discoveredArenas];
			for (const arena of arenas) {
				const index = knownArenas.findIndex((knownArena) => knownArena.id === arena.id);
				if (index === -1) knownArenas.push(arena);
				else knownArenas[index] = arena;
			}
			appState.discoveredArenas = knownArenas;
		}
	});

	onMount(() => {
		appState.theme = readThemePreference(window.localStorage);
		applyTheme(appState.theme, document.documentElement);
		void loadUser();
	});

	async function loadUser() {
		try {
			appState.user = await api.me();
			appState.sessionStatus = 'ready';
		} catch {
			await goto(resolve('/'), { replaceState: true });
		}
	}
</script>

<div
	class="grid h-svh w-full place-items-center bg-[var(--page-outside)] px-0 py-0 sm:px-6 sm:py-6"
>
	<div
		class="flex h-svh w-full max-w-[430px] flex-col overflow-hidden bg-[var(--app-background)] pt-[env(safe-area-inset-top)] text-[var(--text-primary)] shadow-[0_32px_100px_rgba(0,0,0,0.24)] transition-colors duration-300 motion-reduce:transition-none sm:h-[min(860px,calc(100svh-48px))] sm:rounded-[32px] sm:border sm:border-[var(--border)]"
	>
		{#if appState.sessionStatus === 'loading'}
			<main class="grid min-h-0 flex-1 place-items-center" role="status" aria-live="polite">
				<LoaderCircle
					size={28}
					class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
					aria-label="Validando sessão"
				/>
			</main>
		{:else}
			<main class="relative flex min-h-0 flex-1 bg-[var(--app-background)]">
				{@render children()}
			</main>
			<BottomNavigation />
		{/if}
	</div>
</div>
