<script lang="ts">
	import { CalendarDays, UserRound } from '@lucide/svelte';
	import type { User } from '../api';
	import type { Theme } from '../theme';
	import ArenaManagement from './ArenaManagement.svelte';
	import BottomNavigation from './BottomNavigation.svelte';
	import MapView from './MapView.svelte';

	type MainTab = 'buscar' | 'jogos' | 'arenas' | 'perfil';
	type Props = {
		user: User | null;
		theme: Theme;
	};

	let { user, theme }: Props = $props();
	let activeTab = $state<MainTab>('buscar');

	function selectTab(tab: MainTab) {
		activeTab = tab;
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
				<MapView {theme} />
			</div>
			{#if activeTab === 'jogos'}
				<section
					class="relative z-10 flex min-h-0 flex-1 flex-col items-center justify-center gap-4 bg-[var(--app-background)]"
				>
					<CalendarDays
						size={34}
						strokeWidth={1.7}
						class="text-[var(--accent)]"
						aria-hidden="true"
					/>
					<h1 class="text-2xl font-semibold tracking-tight">Jogos</h1>
				</section>
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
		</main>

		<BottomNavigation {activeTab} onSelect={selectTab} />
	</div>
</div>
