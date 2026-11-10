<script lang="ts">
	import { afterNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		ArrowLeft,
		CalendarDays,
		Check,
		LoaderCircle,
		MapPin,
		RefreshCw,
		UsersRound,
		UserRound
	} from '@lucide/svelte';
	import { api, resolveAssetURL, type GameLobby } from '../../../../lib/api';
	import { getUserFacingErrorMessage } from '../../../../lib/errors';
	import { getAppContext } from '../../../../lib/app-context';
	import type { PageProps } from './$types';

	let { params }: PageProps = $props();
	const { appState, refreshGameAvailability } = getAppContext();
	let lobby = $state<GameLobby | null>(null);
	let isLoading = $state(true);
	let isJoining = $state(false);
	let loadError = $state('');
	let actionError = $state('');
	let currentUserIsParticipant = $derived(
		lobby?.participants.some((participant) => participant.id === appState.user?.id) ?? false
	);
	let currentUserIsOrganizer = $derived(
		lobby !== null && lobby.game.organizer_user_id === appState.user?.id
	);

	afterNavigate(() => {
		void loadLobby(Number(params.id));
	});

	async function loadLobby(gameID: number) {
		isLoading = true;
		loadError = '';

		try {
			lobby = await api.getGameLobby(gameID);
		} catch (error) {
			lobby = null;
			loadError = getUserFacingErrorMessage(error, 'Não foi possível carregar a sala do jogo.');
		} finally {
			isLoading = false;
		}
	}

	async function joinGame() {
		const gameID = Number(params.id);
		isJoining = true;
		actionError = '';

		try {
			await api.joinGame(gameID);
			refreshGameAvailability();
		} catch (error) {
			actionError = getUserFacingErrorMessage(error, 'Não foi possível confirmar sua presença.');
			isJoining = false;
			return;
		}

		isJoining = false;
		await loadLobby(gameID);
	}

	function refreshLobby() {
		void loadLobby(Number(params.id));
	}

	function goBack() {
		if (window.history.length > 1) window.history.back();
		else void goto(resolve('/(app)/jogos'));
	}

	function formatDay(value: string) {
		return new Intl.DateTimeFormat('pt-BR', {
			weekday: 'long',
			day: 'numeric',
			month: 'long'
		}).format(new Date(value));
	}

	function formatTime(value: string) {
		return new Intl.DateTimeFormat('pt-BR', {
			hour: '2-digit',
			minute: '2-digit'
		}).format(new Date(value));
	}

	function initials(name: string) {
		return name
			.trim()
			.split(/\s+/)
			.slice(0, 2)
			.map((part) => part[0])
			.join('')
			.toUpperCase();
	}
</script>

<svelte:head>
	<title>ONRACHA — Sala do jogo</title>
</svelte:head>

<section class="relative z-10 flex min-h-0 flex-1 flex-col bg-[var(--app-background)]">
	<header
		class="flex shrink-0 items-center justify-between border-b border-[var(--border)] px-4 py-3"
	>
		<button
			class="grid size-10 place-items-center rounded-full text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface)] hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
			type="button"
			aria-label="Voltar"
			onclick={goBack}
		>
			<ArrowLeft size={19} aria-hidden="true" />
		</button>
		<h1 class="text-base font-semibold">Sala do jogo</h1>
		<button
			class="grid size-10 place-items-center rounded-full text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface)] hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none disabled:opacity-50"
			type="button"
			aria-label="Atualizar sala"
			disabled={isLoading || isJoining}
			onclick={refreshLobby}
		>
			<RefreshCw size={17} class={isLoading ? 'animate-spin' : ''} aria-hidden="true" />
		</button>
	</header>

	{#if isLoading}
		<div class="grid min-h-0 flex-1 place-items-center" role="status" aria-live="polite">
			<LoaderCircle
				size={28}
				class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
				aria-label="Carregando sala"
			/>
		</div>
	{:else if loadError || !lobby}
		<div class="grid min-h-0 flex-1 content-center justify-items-center gap-4 px-8 text-center">
			<p class="max-w-xs rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm" role="alert">
				{loadError || 'Este jogo não foi encontrado.'}
			</p>
			<button
				class="inline-flex min-h-11 items-center gap-2 rounded-2xl bg-[var(--accent)] px-5 font-semibold text-slate-950 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				onclick={refreshLobby}
			>
				<RefreshCw size={16} aria-hidden="true" />
				Tentar novamente
			</button>
		</div>
	{:else}
		<div class="min-h-0 flex-1 overflow-y-auto px-4 py-4">
			<article
				class="overflow-hidden rounded-[26px] border border-[var(--border)] bg-[var(--surface)]"
			>
				{#if lobby.arena.photo_url}
					<img
						class="h-40 w-full object-cover"
						src={resolveAssetURL(lobby.arena.photo_url) ?? undefined}
						alt={`Foto da arena ${lobby.arena.name}`}
					/>
				{:else}
					<div
						class="grid h-32 place-items-center bg-[linear-gradient(135deg,var(--surface-raised),var(--focus-ring))] text-[var(--accent)]"
						aria-hidden="true"
					>
						<MapPin size={36} strokeWidth={1.5} />
					</div>
				{/if}

				<div class="p-4">
					<h2 class="text-xl font-semibold tracking-tight">{lobby.arena.name}</h2>
					<p class="mt-1 flex items-center gap-1.5 text-sm text-[var(--text-secondary)]">
						<MapPin size={15} aria-hidden="true" />
						<span>{lobby.arena.city}, {lobby.arena.state}</span>
					</p>
					<p class="mt-1 text-sm text-[var(--text-secondary)]">{lobby.arena.address}</p>

					<div class="mt-4 grid gap-3 border-t border-[var(--border)] pt-4">
						<div class="flex items-center gap-3">
							<div
								class="grid size-10 shrink-0 place-items-center rounded-2xl bg-[var(--focus-ring)] text-[var(--accent)]"
							>
								<CalendarDays size={18} aria-hidden="true" />
							</div>
							<div>
								<p class="text-sm font-semibold capitalize">{formatDay(lobby.game.starts_at)}</p>
								<p class="mt-0.5 text-sm text-[var(--text-secondary)]">
									{formatTime(lobby.game.starts_at)}–{formatTime(lobby.game.ends_at)}
								</p>
							</div>
						</div>
						<div class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
							<UsersRound size={17} aria-hidden="true" />
							{lobby.participants.length} confirmados
						</div>
					</div>
				</div>
			</article>

			<section class="mt-5" aria-labelledby="lobby-players-title">
				<div class="mb-3 flex items-center justify-between">
					<h2 id="lobby-players-title" class="text-lg font-semibold">Jogadores</h2>
					<span class="text-sm text-[var(--text-secondary)]"
						>{lobby.participants.length} confirmados</span
					>
				</div>
				<div
					class="overflow-hidden rounded-[24px] border border-[var(--border)] bg-[var(--surface)]"
				>
					{#each lobby.participants as participant (participant.id)}
						{@const isOrganizer = participant.id === lobby.game.organizer_user_id}
						<div
							class="flex items-center gap-3 border-b border-[var(--border)] px-4 py-3 last:border-b-0"
						>
							<div
								class="grid size-11 shrink-0 place-items-center rounded-full bg-[var(--focus-ring)] text-sm font-semibold text-[var(--accent)]"
							>
								{initials(participant.name) || '?'}
							</div>
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm font-semibold">
									{participant.name}
									{#if participant.id === appState.user?.id}
										<span class="font-normal text-[var(--text-secondary)]">(você)</span>
									{/if}
								</p>
								<p class="mt-0.5 text-xs text-[var(--text-secondary)]">
									{isOrganizer ? 'Organizador' : 'Presença confirmada'}
								</p>
							</div>
							{#if isOrganizer}
								<span
									class="rounded-full bg-[var(--focus-ring)] px-2.5 py-1 text-[11px] font-semibold text-[var(--accent)]"
								>
									Organiza
								</span>
							{/if}
						</div>
					{/each}
					{#if lobby.participants.length === 0}
						<div class="flex items-center gap-3 px-4 py-4 text-sm text-[var(--text-secondary)]">
							<UserRound size={18} aria-hidden="true" />
							Ainda não há jogadores confirmados.
						</div>
					{/if}
				</div>
			</section>

			{#if actionError}
				<p
					class="mt-4 rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm text-[var(--danger)]"
					role="alert"
				>
					{actionError}
				</p>
			{/if}

			{#if currentUserIsParticipant}
				<div
					class="mt-5 flex min-h-12 items-center justify-center gap-2 rounded-2xl border border-[var(--accent)]/25 bg-[var(--focus-ring)] px-4 text-sm font-semibold text-[var(--accent)]"
					role="status"
				>
					<Check size={18} aria-hidden="true" />
					{currentUserIsOrganizer ? 'Você organiza este jogo' : 'Você está confirmado'}
				</div>
			{:else}
				<button
					class="mt-5 flex min-h-12 w-full items-center justify-between rounded-2xl bg-[var(--accent)] px-5 text-sm font-semibold text-slate-950 transition-transform focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.99] disabled:cursor-wait disabled:opacity-65"
					type="button"
					disabled={isJoining}
					onclick={joinGame}
				>
					<span>{isJoining ? 'Confirmando presença…' : 'Entrar neste jogo'}</span>
					{#if isJoining}
						<LoaderCircle
							size={18}
							class="animate-spin motion-reduce:animate-none"
							aria-hidden="true"
						/>
					{:else}
						<Check size={19} aria-hidden="true" />
					{/if}
				</button>
			{/if}
		</div>
	{/if}
</section>
