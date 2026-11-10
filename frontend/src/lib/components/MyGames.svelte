<script lang="ts">
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import {
		CalendarDays,
		Clock3,
		LoaderCircle,
		MapPin,
		ChevronRight,
		RefreshCw,
		UsersRound
	} from '@lucide/svelte';
	import { api, type DiscoveredArena, type Game } from '../api';
	import { getUserFacingErrorMessage } from '../errors';

	type Props = {
		arenas: DiscoveredArena[];
		userID: number | null;
		onFindArenas: () => void;
	};

	let { arenas, userID, onFindArenas }: Props = $props();
	let games = $state<Game[]>([]);
	let isLoading = $state(true);
	let loadError = $state('');

	onMount(() => {
		void loadGames();
	});

	async function loadGames() {
		isLoading = true;
		loadError = '';

		try {
			games = await api.listMyGames();
		} catch (error) {
			loadError = getUserFacingErrorMessage(error, 'Não foi possível carregar seus jogos.');
		} finally {
			isLoading = false;
		}
	}

	function findArena(arenaID: number) {
		return arenas.find((arena) => arena.id === arenaID);
	}

	function formatDay(value: string) {
		return new Intl.DateTimeFormat('pt-BR', {
			weekday: 'short',
			day: '2-digit',
			month: 'short'
		}).format(new Date(value));
	}

	function formatTime(value: string) {
		return new Intl.DateTimeFormat('pt-BR', {
			hour: '2-digit',
			minute: '2-digit'
		}).format(new Date(value));
	}
</script>

<section class="relative z-10 flex min-h-0 flex-1 flex-col bg-[var(--app-background)]">
	<header
		class="flex shrink-0 items-center justify-between border-b border-[var(--border)] px-5 py-4"
	>
		<h1 class="text-2xl font-semibold tracking-tight">Jogos</h1>
		{#if games.length > 0}
			<button
				class="grid size-10 place-items-center rounded-full text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface)] hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				aria-label="Atualizar jogos"
				onclick={loadGames}
			>
				<RefreshCw size={17} aria-hidden="true" />
			</button>
		{/if}
	</header>

	{#if isLoading}
		<div class="grid min-h-0 flex-1 place-items-center" role="status" aria-live="polite">
			<LoaderCircle
				size={28}
				class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
				aria-label="Carregando jogos"
			/>
		</div>
	{:else if loadError}
		<div class="grid min-h-0 flex-1 content-center justify-items-center gap-4 px-8 text-center">
			<p class="max-w-xs rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm" role="alert">
				{loadError}
			</p>
			<button
				class="inline-flex min-h-11 items-center gap-2 rounded-2xl bg-[var(--accent)] px-5 font-semibold text-slate-950 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				onclick={loadGames}
			>
				<RefreshCw size={16} aria-hidden="true" />
				Tentar novamente
			</button>
		</div>
	{:else if games.length === 0}
		<div class="grid min-h-0 flex-1 content-center justify-items-center px-8 text-center">
			<div
				class="mb-5 grid size-20 place-items-center rounded-[28px] bg-[var(--surface)] text-[var(--accent)]"
				aria-hidden="true"
			>
				<CalendarDays size={36} strokeWidth={1.6} />
			</div>
			<h2 class="text-xl font-semibold tracking-tight">Nenhum jogo confirmado</h2>
			<p class="mt-2 max-w-xs text-sm leading-6 text-[var(--text-secondary)]">
				Encontre uma arena no mapa e entre em um jogo.
			</p>
			<button
				class="mt-6 inline-flex min-h-12 items-center gap-2 rounded-2xl bg-[var(--accent)] px-5 font-semibold text-slate-950 transition-transform focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.98]"
				type="button"
				onclick={onFindArenas}
			>
				<MapPin size={17} aria-hidden="true" />
				Encontrar arenas
			</button>
		</div>
	{:else}
		<div class="min-h-0 flex-1 overflow-y-auto px-4 py-4">
			<div class="grid gap-3">
				{#each games as game (game.id)}
					{@const arena = findArena(game.arena_id)}
					<article
						class="overflow-hidden rounded-[24px] border border-[var(--border)] bg-[var(--surface)] shadow-sm transition-colors hover:border-[var(--accent)]/45"
					>
						<a
							class="block p-4 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
							href={resolve('/(app)/jogos/[id]', { id: String(game.id) })}
						>
							<div class="flex items-start justify-between gap-3">
								<div class="flex min-w-0 items-center gap-3">
									{#if arena?.photo_url}
										<img
											class="size-14 shrink-0 rounded-2xl object-cover"
											src={arena.photo_url}
											alt={`Foto de ${arena.name}`}
										/>
									{:else}
										<div
											class="grid size-14 shrink-0 place-items-center rounded-2xl bg-[var(--focus-ring)] text-[var(--accent)]"
											aria-hidden="true"
										>
											<MapPin size={21} />
										</div>
									{/if}
									<div class="min-w-0">
										<p class="text-base font-semibold capitalize">{formatDay(game.starts_at)}</p>
										<p class="mt-1 truncate text-sm text-[var(--text-secondary)]">
											{arena?.name ?? `Arena ${game.arena_id}`}
										</p>
										{#if arena}
											<p class="mt-1 truncate text-xs text-[var(--text-secondary)]">
												{arena.city}, {arena.state}
											</p>
										{/if}
									</div>
								</div>
								<span
									class="shrink-0 rounded-full bg-[var(--focus-ring)] px-3 py-1.5 text-xs font-semibold text-[var(--accent)]"
								>
									{userID !== null && game.organizer_user_id === userID
										? 'Organizando'
										: 'Confirmado'}
								</span>
								<ChevronRight
									size={18}
									class="shrink-0 text-[var(--text-secondary)]"
									aria-hidden="true"
								/>
							</div>

							<div
								class="mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-[var(--text-secondary)]"
							>
								<span class="inline-flex items-center gap-1.5">
									<Clock3 size={15} aria-hidden="true" />
									{formatTime(game.starts_at)}–{formatTime(game.ends_at)}
								</span>
								<span class="inline-flex items-center gap-1.5">
									<UsersRound size={15} aria-hidden="true" />
									{game.participant_user_ids.length} confirmados
								</span>
							</div>
						</a>
					</article>
				{/each}
			</div>
		</div>
	{/if}
</section>
