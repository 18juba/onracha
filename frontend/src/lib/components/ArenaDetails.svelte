<script lang="ts">
	import Button from './ui/Button.svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import {
		CalendarDays,
		ChevronRight,
		Clock3,
		LoaderCircle,
		MapPin,
		UsersRound,
		X
	} from '@lucide/svelte';
	import { api, type DiscoveredArena, type GameSearchResult, type TimeSlot } from '../api';
	import { getUserFacingErrorMessage } from '../errors';

	type Props = {
		arena: DiscoveredArena;
		onClose: () => void;
		onGamesChanged: () => void;
	};

	let { arena, onClose, onGamesChanged }: Props = $props();
	let timeSlots = $state<TimeSlot[]>([]);
	let availableGames = $state<GameSearchResult[]>([]);
	let isLoading = $state(true);
	let pendingAction = $state('');
	let loadError = $state('');
	let actionError = $state('');
	let arenaGames = $derived(availableGames.filter((result) => result.arena.id === arena.id));

	onMount(() => {
		void loadDetails();
	});

	async function loadDetails() {
		isLoading = true;
		loadError = '';

		const [slotsResult, gamesResult] = await Promise.allSettled([
			api.listArenaTimeSlots(arena.id),
			api.listAvailableGames(arena.city)
		]);

		if (slotsResult.status === 'fulfilled') {
			timeSlots = slotsResult.value;
		} else {
			loadError = getUserFacingErrorMessage(
				slotsResult.reason,
				'Não foi possível carregar os horários da arena.'
			);
		}

		if (gamesResult.status === 'fulfilled') {
			availableGames = gamesResult.value;
		} else if (!loadError) {
			loadError = getUserFacingErrorMessage(
				gamesResult.reason,
				'Não foi possível carregar os jogos disponíveis.'
			);
		}

		isLoading = false;
	}

	async function scheduleGame(slot: TimeSlot) {
		pendingAction = `slot-${slot.id}`;
		actionError = '';
		let createdGame: GameSearchResult['game'] | null = null;

		try {
			createdGame = await api.scheduleGame(slot.id);
		} catch (error) {
			actionError = getUserFacingErrorMessage(error, 'Não foi possível agendar esse jogo.');
		} finally {
			pendingAction = '';
		}

		if (createdGame) {
			onGamesChanged();
			await goto(resolve('/(app)/jogos/[id]', { id: String(createdGame.id) }));
		}
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

<div class="absolute inset-0 z-[600] flex items-end">
	<Button
		type="button"
		aria-label="Fechar detalhes da arena"
		onclick={onClose}
		variant="unstyled"
		size="none"
		class="absolute inset-0 bg-black/35"
	></Button>

	<section
		class="relative flex max-h-[82%] w-full flex-col overflow-hidden rounded-t-[30px] border border-[var(--border)] bg-[var(--app-background)] shadow-2xl"
		aria-labelledby="arena-detail-title"
	>
		<div class="flex shrink-0 items-center justify-between gap-3 px-5 pt-4 pb-3">
			<div class="flex items-center gap-2">
				<span class="h-1.5 w-9 rounded-full bg-[var(--text-secondary)]/35" aria-hidden="true"
				></span>
				<span
					class="text-xs font-semibold tracking-[0.12em] text-[var(--text-secondary)] uppercase"
				>
					Arena
				</span>
			</div>
			<Button
				type="button"
				aria-label="Fechar"
				onclick={onClose}
				variant="secondary"
				size="icon"
				class=""
			>
				<X size={18} aria-hidden="true" />
			</Button>
		</div>

		<div class="min-h-0 overflow-y-auto overscroll-contain px-5 pb-8">
			{#if arena.photo_url}
				<img
					class="h-36 w-full rounded-[22px] object-cover"
					src={arena.photo_url}
					alt={`Foto da arena ${arena.name}`}
				/>
			{:else}
				<div
					class="grid h-28 w-full place-items-center rounded-[22px] bg-[var(--surface-raised)] text-[var(--accent)]"
					aria-hidden="true"
				>
					<MapPin size={34} strokeWidth={1.6} />
				</div>
			{/if}

			<div class="mt-4">
				<h1 id="arena-detail-title" class="text-2xl font-semibold tracking-tight">
					{arena.name}
				</h1>
				<p class="mt-1 flex items-center gap-1.5 text-sm text-[var(--text-secondary)]">
					<MapPin size={15} aria-hidden="true" />
					<span>{arena.city}, {arena.state}</span>
				</p>
				<p class="mt-1 text-sm text-[var(--text-secondary)]">{arena.address}</p>
			</div>

			{#if actionError || loadError}
				<p class="mt-4 rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm" role="alert">
					{actionError || loadError}
				</p>
			{/if}

			{#if isLoading}
				<div class="grid min-h-36 place-items-center" role="status" aria-live="polite">
					<LoaderCircle
						size={24}
						class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
						aria-label="Carregando horários e jogos"
					/>
				</div>
			{:else}
				<section class="mt-6" aria-labelledby="arena-games-title">
					<div class="mb-3 flex items-center gap-2">
						<UsersRound size={18} class="text-[var(--accent)]" aria-hidden="true" />
						<h2 id="arena-games-title" class="text-lg font-semibold">Jogos para entrar</h2>
					</div>

					{#if arenaGames.length > 0}
						<div class="grid gap-2.5">
							{#each arenaGames as result (result.game.id)}
								<a
									class="flex items-center gap-3 rounded-[22px] border border-[var(--border)] bg-[var(--surface)] p-3.5 text-left transition-colors hover:border-[var(--accent)]/45 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
									href={resolve('/(app)/jogos/[id]', { id: String(result.game.id) })}
								>
									<div class="min-w-0 flex-1">
										<p class="text-sm font-semibold capitalize">
											{formatDay(result.game.starts_at)}
										</p>
										<p class="mt-1 flex items-center gap-1.5 text-sm text-[var(--text-secondary)]">
											<Clock3 size={14} aria-hidden="true" />
											{formatTime(result.game.starts_at)}–{formatTime(result.game.ends_at)}
										</p>
										<p class="mt-1 flex items-center gap-1.5 text-xs text-[var(--text-secondary)]">
											<UsersRound size={13} aria-hidden="true" />
											{result.game.participant_user_ids.length} confirmados
										</p>
									</div>
									<span
										class="inline-flex shrink-0 items-center gap-1 text-sm font-semibold text-[var(--accent)]"
									>
										Ver sala
										<ChevronRight size={17} aria-hidden="true" />
									</span>
								</a>
							{/each}
						</div>
					{:else}
						<p
							class="rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm text-[var(--text-secondary)]"
						>
							Ainda não há jogos abertos nesta arena.
						</p>
					{/if}
				</section>

				<section class="mt-6" aria-labelledby="arena-slots-title">
					<div class="mb-3 flex items-center gap-2">
						<CalendarDays size={18} class="text-[var(--accent)]" aria-hidden="true" />
						<h2 id="arena-slots-title" class="text-lg font-semibold">Horários livres</h2>
					</div>

					{#if timeSlots.length > 0}
						<div class="grid gap-2.5">
							{#each timeSlots as slot (slot.id)}
								<article
									class="flex items-center gap-3 rounded-[22px] border border-[var(--border)] bg-[var(--surface)] p-3.5"
								>
									<div class="min-w-0 flex-1">
										<p class="text-sm font-semibold capitalize">{formatDay(slot.starts_at)}</p>
										<p class="mt-1 flex items-center gap-1.5 text-sm text-[var(--text-secondary)]">
											<Clock3 size={14} aria-hidden="true" />
											{formatTime(slot.starts_at)}–{formatTime(slot.ends_at)}
										</p>
									</div>
									<Button
										type="button"
										disabled={pendingAction !== ''}
										onclick={() => scheduleGame(slot)}
										variant="outline"
										size="default"
										class="shrink-0"
									>
										{pendingAction === `slot-${slot.id}` ? 'Agendando…' : 'Criar jogo'}
									</Button>
								</article>
							{/each}
						</div>
					{:else}
						<p
							class="rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm text-[var(--text-secondary)]"
						>
							Nenhum horário livre no momento.
						</p>
					{/if}
				</section>
			{/if}
		</div>
	</section>
</div>
