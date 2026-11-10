<script lang="ts">
	import { UserRound } from '@lucide/svelte';
	import type { DiscoveredArena, User } from '../api';
	import type { Theme } from '../theme';
	import ArenaDetails from './ArenaDetails.svelte';
	import ArenaManagement from './ArenaManagement.svelte';
	import BottomNavigation from './BottomNavigation.svelte';
	import MapView from './MapView.svelte';
	import MyGames from './MyGames.svelte';

	type MainTab = 'buscar' | 'jogos' | 'arenas' | 'perfil';
	type Props = {
		user: User | null;
		theme: Theme;
	};

	let { user, theme }: Props = $props();
	let activeTab = $state<MainTab>('buscar');
	let selectedArena = $state<DiscoveredArena | null>(null);
	let discoveredArenas = $state<DiscoveredArena[]>([]);
	let gameAvailabilityVersion = $state(0);

	function selectTab(tab: MainTab) {
		selectedArena = null;
		activeTab = tab;
	}

	function selectArena(arena: DiscoveredArena) {
		selectedArena = arena;
	}

	function refreshGameAvailability() {
		gameAvailabilityVersion += 1;
	}

	function rememberDiscoveredArenas(arenas: DiscoveredArena[]) {
		const knownArenas = [...discoveredArenas];
		for (const arena of arenas) {
			const existingIndex = knownArenas.findIndex((knownArena) => knownArena.id === arena.id);
			if (existingIndex === -1) {
				knownArenas.push(arena);
			} else {
				knownArenas[existingIndex] = arena;
			}
		}
		discoveredArenas = knownArenas;
	}
</script>

<div
	class="grid h-svh w-full place-items-center bg-[var(--page-outside)] px-0 py-0 sm:px-6 sm:py-6"
>
	<div
		class="flex h-svh w-full max-w-[430px] flex-col overflow-hidden bg-[var(--app-background)] pt-[env(safe-area-inset-top)] text-[var(--text-primary)] shadow-[0_32px_100px_rgba(0,0,0,0.24)] transition-colors duration-300 motion-reduce:transition-none sm:h-[min(860px,calc(100svh-48px))] sm:rounded-[32px] sm:border sm:border-[var(--border)]"
	>
		<main class="relative flex min-h-0 flex-1 bg-[var(--app-background)]">
			<div
				class={`absolute inset-0 flex ${activeTab === 'buscar' ? 'visible z-0' : 'pointer-events-none invisible -z-10'}`}
				inert={activeTab !== 'buscar'}
			>
				<MapView
					{theme}
					discoverArenas
					{gameAvailabilityVersion}
					onArenaSelect={selectArena}
					onArenasChange={rememberDiscoveredArenas}
				/>
			</div>
			{#if activeTab === 'jogos'}
				<MyGames
					arenas={discoveredArenas}
					userID={user?.id ?? null}
					onFindArenas={() => selectTab('buscar')}
				/>
			{:else if activeTab === 'arenas'}
				<ArenaManagement {theme} />
			{:else if activeTab === 'perfil'}
				<section
					class="relative z-10 flex min-h-0 flex-1 flex-col items-center justify-center gap-4 bg-[var(--app-background)] text-center"
				>
					<UserRound size={34} strokeWidth={1.7} class="text-[var(--accent)]" aria-hidden="true" />
					<h1 class="text-2xl font-semibold tracking-tight">{user?.name ?? 'Perfil'}</h1>
					{#if user?.email}
						<p class="text-sm text-[var(--text-secondary)]">{user.email}</p>
					{/if}
				</section>
			{/if}
			{#if activeTab === 'buscar' && selectedArena}
				{#key selectedArena.id}
					<ArenaDetails
						arena={selectedArena}
						onClose={() => (selectedArena = null)}
						onGamesChanged={refreshGameAvailability}
					/>
				{/key}
			{/if}
		</main>

		<BottomNavigation {activeTab} onSelect={selectTab} />
	</div>
</div>
