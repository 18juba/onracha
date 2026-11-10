<script lang="ts">
	import { Building2, CalendarDays, MapPin, UserRound } from '@lucide/svelte';
	import type { User } from '../api';
	import BottomNavigation from './BottomNavigation.svelte';

	type MainTab = 'buscar' | 'jogos' | 'arenas' | 'perfil';
	type Props = {
		user: User | null;
	};

	let { user }: Props = $props();
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
			{#if activeTab === 'buscar'}
				<div
					class="grid min-h-0 flex-1 place-items-center overflow-hidden"
					role="region"
					aria-label="Mapa de arenas"
				>
					<div
						class="grid size-[72px] place-items-center rounded-[26px] border border-[var(--border)] bg-[var(--surface)] text-[var(--accent)] shadow-[0_14px_40px_rgba(0,0,0,0.12)]"
						aria-hidden="true"
					>
						<MapPin size={32} strokeWidth={1.8} />
					</div>
				</div>
			{:else if activeTab === 'jogos'}
				<section class="flex min-h-0 flex-1 flex-col items-center justify-center gap-4">
					<CalendarDays
						size={34}
						strokeWidth={1.7}
						class="text-[var(--accent)]"
						aria-hidden="true"
					/>
					<h1 class="text-2xl font-semibold tracking-tight">Jogos</h1>
				</section>
			{:else if activeTab === 'arenas'}
				<section class="flex min-h-0 flex-1 flex-col items-center justify-center gap-4">
					<Building2 size={34} strokeWidth={1.7} class="text-[var(--accent)]" aria-hidden="true" />
					<h1 class="text-2xl font-semibold tracking-tight">Arenas</h1>
				</section>
			{:else}
				<section class="flex min-h-0 flex-1 flex-col items-center justify-center gap-4 text-center">
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
