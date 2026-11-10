<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import {
		AlertCircle,
		ArrowLeft,
		ArrowRight,
		Building2,
		Check,
		ImagePlus,
		LoaderCircle,
		MapPin,
		Plus,
		X
	} from '@lucide/svelte';
	import { api, type Arena } from '../api';
	import { getUserFacingErrorMessage } from '../errors';
	import type { Theme } from '../theme';
	import ArenaSchedule from './ArenaSchedule.svelte';
	import MapView from './MapView.svelte';

	type Coordinates = { latitude: number; longitude: number };
	type Screen = 'list' | 'details' | 'location' | 'photo' | 'schedule';
	type EditReturnScreen = 'list' | 'schedule';

	type Props = { theme: Theme };

	const arenaTypes = [
		{ value: 'society', label: 'Society' },
		{ value: 'futsal', label: 'Futsal' },
		{ value: 'campo', label: 'Campo' },
		{ value: 'poliesportiva', label: 'Poliesportiva' }
	];
	const maxPhotoSize = 5 * 1024 * 1024;
	const supportedPhotoTypes = ['image/jpeg', 'image/png'];
	const brazilianStates = [
		{ value: 'AC', label: 'Acre' },
		{ value: 'AL', label: 'Alagoas' },
		{ value: 'AP', label: 'Amapá' },
		{ value: 'AM', label: 'Amazonas' },
		{ value: 'BA', label: 'Bahia' },
		{ value: 'CE', label: 'Ceará' },
		{ value: 'DF', label: 'Distrito Federal' },
		{ value: 'ES', label: 'Espírito Santo' },
		{ value: 'GO', label: 'Goiás' },
		{ value: 'MA', label: 'Maranhão' },
		{ value: 'MT', label: 'Mato Grosso' },
		{ value: 'MS', label: 'Mato Grosso do Sul' },
		{ value: 'MG', label: 'Minas Gerais' },
		{ value: 'PA', label: 'Pará' },
		{ value: 'PB', label: 'Paraíba' },
		{ value: 'PR', label: 'Paraná' },
		{ value: 'PE', label: 'Pernambuco' },
		{ value: 'PI', label: 'Piauí' },
		{ value: 'RJ', label: 'Rio de Janeiro' },
		{ value: 'RN', label: 'Rio Grande do Norte' },
		{ value: 'RS', label: 'Rio Grande do Sul' },
		{ value: 'RO', label: 'Rondônia' },
		{ value: 'RR', label: 'Roraima' },
		{ value: 'SC', label: 'Santa Catarina' },
		{ value: 'SP', label: 'São Paulo' },
		{ value: 'SE', label: 'Sergipe' },
		{ value: 'TO', label: 'Tocantins' }
	];

	let { theme }: Props = $props();
	let arenas = $state<Arena[]>([]);
	let screen = $state<Screen>('list');
	let isLoading = $state(true);
	let isSaving = $state(false);
	let loadError = $state('');
	let formError = $state('');
	let successMessage = $state('');
	let selectedArena = $state<Arena | null>(null);
	let editReturnScreen = $state<EditReturnScreen>('list');
	let showLocationInstructions = $state(true);
	let editingArena = $state<Arena | null>(null);
	let arenaActive = $state(true);
	let name = $state('');
	let arenaType = $state('society');
	let address = $state('');
	let city = $state('');
	let stateCode = $state('');
	let selectedLocation = $state<Coordinates | null>(null);
	let selectedPhoto = $state<File | null>(null);
	let photoPreview = $state('');
	let currentArenaPhotoURL = $state<string | null>(null);
	let photoError = $state('');
	let photoInputKey = $state(0);
	let pendingPhotoArena = $state<Arena | null>(null);

	onMount(() => {
		void loadArenas();
	});

	onDestroy(() => {
		if (photoPreview) URL.revokeObjectURL(photoPreview);
	});

	async function loadArenas() {
		isLoading = true;
		loadError = '';

		try {
			arenas = await api.listArenas();
		} catch (error) {
			loadError = getUserFacingErrorMessage(error, 'Não foi possível carregar suas arenas.');
		} finally {
			isLoading = false;
		}
	}

	function startCreation() {
		resetForm();
		selectedArena = null;
		editReturnScreen = 'list';
		formError = '';
		successMessage = '';
		screen = 'details';
	}

	function startEditing(
		arena: Arena | null = selectedArena,
		returnTo: EditReturnScreen = 'schedule'
	) {
		if (!arena) return;

		resetForm();
		editReturnScreen = returnTo;
		editingArena = arena;
		arenaActive = arena.active;
		name = arena.name;
		arenaType = arena.arena_type;
		address = arena.address;
		city = arena.city;
		stateCode = arena.state;
		selectedLocation =
			arena.latitude !== null && arena.longitude !== null
				? { latitude: arena.latitude, longitude: arena.longitude }
				: null;
		currentArenaPhotoURL = arena.photo_url;
		formError = '';
		successMessage = '';
		screen = 'details';
	}

	function openSchedule(arena: Arena) {
		selectedArena = arena;
		screen = 'schedule';
	}

	function updateSelectedArena(arena: Arena) {
		upsertArena(arena);
		selectedArena = arena;
	}

	function closeSchedule() {
		selectedArena = null;
		screen = 'list';
	}

	function showLocationPicker(event: SubmitEvent) {
		event.preventDefault();
		formError = '';
		photoError = '';
		showLocationInstructions = true;
		screen = 'location';
	}

	function showPhotoStep() {
		formError = '';
		screen = 'photo';
	}

	function selectPhoto(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		if (file.size > maxPhotoSize) {
			photoError = 'A foto deve ter no máximo 5 MB.';
			input.value = '';
			return;
		}

		if (file.type && !supportedPhotoTypes.includes(file.type)) {
			photoError = 'Escolha uma imagem JPEG ou PNG.';
			input.value = '';
			return;
		}

		if (photoPreview) URL.revokeObjectURL(photoPreview);
		selectedPhoto = file;
		photoPreview = URL.createObjectURL(file);
		photoError = '';
		formError = '';
	}

	function clearPhoto() {
		selectedPhoto = null;
		photoError = '';
		if (photoPreview) URL.revokeObjectURL(photoPreview);
		photoPreview = '';
		photoInputKey += 1;
	}

	function selectLocation(location: Coordinates) {
		selectedLocation = location;
		showLocationInstructions = false;
		formError = '';
	}

	function resetForm() {
		editingArena = null;
		arenaActive = true;
		name = '';
		arenaType = 'society';
		address = '';
		city = '';
		stateCode = '';
		selectedLocation = null;
		currentArenaPhotoURL = null;
		formError = '';
		photoError = '';
		clearPhoto();
		pendingPhotoArena = null;
	}

	async function saveArena() {
		const location = selectedLocation;
		const arenaBeingEdited = editingArena;
		if ((!arenaBeingEdited && !location && !pendingPhotoArena) || isSaving) return;

		isSaving = true;
		formError = '';

		try {
			if (pendingPhotoArena) {
				if (!selectedPhoto) return;

				let arenaWithPhoto: Arena;
				try {
					arenaWithPhoto = await api.replaceArenaPhoto(pendingPhotoArena.id, selectedPhoto);
				} catch (error) {
					formError = `Não foi possível enviar a foto. ${getUserFacingErrorMessage(error, 'Tente novamente.')}`;
					return;
				}

				upsertArena(arenaWithPhoto);
				if (arenaBeingEdited) selectedArena = arenaWithPhoto;
				pendingPhotoArena = null;
				successMessage = arenaBeingEdited
					? 'Alterações salvas com foto.'
					: 'Arena cadastrada com foto.';
				resetForm();
				screen = arenaBeingEdited ? 'schedule' : 'list';
				return;
			}

			let savedArena: Arena;
			if (arenaBeingEdited) {
				savedArena = await api.updateArena(arenaBeingEdited.id, {
					name: name.trim(),
					arena_type: arenaType,
					address: address.trim(),
					city: city.trim(),
					state: stateCode.trim().toUpperCase(),
					latitude: location?.latitude ?? null,
					longitude: location?.longitude ?? null,
					active: arenaActive
				});
			} else {
				if (!location) return;
				savedArena = await api.createArena({
					name: name.trim(),
					arena_type: arenaType,
					address: address.trim(),
					city: city.trim(),
					state: stateCode.trim().toUpperCase(),
					latitude: location.latitude,
					longitude: location.longitude
				});
			}

			if (selectedPhoto) {
				try {
					savedArena = await api.replaceArenaPhoto(savedArena.id, selectedPhoto);
				} catch (error) {
					upsertArena(savedArena);
					if (arenaBeingEdited) selectedArena = savedArena;
					pendingPhotoArena = savedArena;
					formError = `Os dados foram salvos, mas a foto não foi enviada. ${getUserFacingErrorMessage(error, 'Tente novamente.')}`;
					return;
				}
			}

			upsertArena(savedArena);
			if (arenaBeingEdited) selectedArena = savedArena;
			successMessage = arenaBeingEdited
				? selectedPhoto
					? 'Alterações salvas com foto.'
					: 'Alterações salvas.'
				: selectedPhoto
					? 'Arena cadastrada com foto.'
					: 'Arena cadastrada.';
			resetForm();
			screen = arenaBeingEdited ? 'schedule' : 'list';
		} catch (error) {
			formError = getUserFacingErrorMessage(
				error,
				arenaBeingEdited
					? 'Não foi possível salvar as alterações.'
					: 'Não foi possível cadastrar a arena.'
			);
		} finally {
			isSaving = false;
		}
	}

	function upsertArena(arena: Arena) {
		if (arenas.some((existingArena) => existingArena.id === arena.id)) {
			arenas = arenas.map((existingArena) =>
				existingArena.id === arena.id ? arena : existingArena
			);
			return;
		}

		arenas = [arena, ...arenas];
	}

	function finishWithoutPhoto() {
		if (!pendingPhotoArena) return;

		const editedExistingArena = editingArena !== null;
		if (editedExistingArena) selectedArena = pendingPhotoArena;
		pendingPhotoArena = null;
		successMessage = editedExistingArena
			? 'Alterações salvas. A foto anterior foi mantida.'
			: 'Arena cadastrada sem foto.';
		resetForm();
		screen = editedExistingArena ? 'schedule' : 'list';
	}

	function cancelArenaForm() {
		if (pendingPhotoArena) {
			finishWithoutPhoto();
			return;
		}

		resetForm();
		screen = editReturnScreen;
	}
</script>

{#if screen === 'list'}
	<section
		class="relative z-10 flex min-h-0 flex-1 flex-col overflow-y-auto bg-[var(--app-background)]"
	>
		<header class="flex items-center justify-between gap-3 px-5 pt-6 pb-4">
			<h1 class="text-2xl font-semibold tracking-tight">Arenas</h1>
			{#if arenas.length > 0}
				<button
					class="inline-flex min-h-11 shrink-0 items-center justify-center gap-2 rounded-2xl bg-[var(--accent)] px-4 text-sm font-semibold text-white transition-opacity hover:opacity-90 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
					type="button"
					onclick={startCreation}
				>
					<Plus size={17} aria-hidden="true" />
					<span>Nova arena</span>
				</button>
			{/if}
		</header>

		{#if successMessage}
			<p
				class="mx-5 mb-3 flex items-center gap-2 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm text-[var(--text-primary)]"
				role="status"
			>
				<Check size={17} class="text-[var(--accent)]" aria-hidden="true" />
				{successMessage}
			</p>
		{/if}

		{#if loadError}
			<div
				class="mx-5 mb-4 flex items-start gap-3 rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-4 text-sm text-[var(--text-primary)]"
				role="alert"
			>
				<AlertCircle size={19} class="mt-0.5 shrink-0 text-[var(--accent)]" aria-hidden="true" />
				<div class="min-w-0 flex-1">
					<p>{loadError}</p>
					<button
						class="mt-2 min-h-10 font-semibold text-[var(--accent)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
						type="button"
						onclick={loadArenas}
					>
						Tentar novamente
					</button>
				</div>
			</div>
		{/if}

		{#if isLoading}
			<div class="flex flex-1 items-center justify-center" role="status" aria-live="polite">
				<LoaderCircle
					size={28}
					class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
					aria-label="Carregando arenas"
				/>
			</div>
		{:else if arenas.length === 0 && !loadError}
			<div class="flex flex-1 flex-col items-center justify-center px-7 pb-8 text-center">
				<div
					class="relative mb-7 grid size-56 place-items-center rounded-[40px] bg-[var(--surface)]"
				>
					<svg
						class="absolute inset-5 size-[calc(100%-2.5rem)] text-[var(--border)]"
						viewBox="0 0 180 180"
						fill="none"
						aria-hidden="true"
					>
						<rect
							x="14"
							y="14"
							width="152"
							height="152"
							rx="28"
							stroke="currentColor"
							stroke-width="2"
						/>
						<path d="M90 14v152M14 90h152" stroke="currentColor" stroke-width="2" />
						<circle cx="90" cy="90" r="25" stroke="currentColor" stroke-width="2" />
						<path d="M14 58h27v64H14m152-64h-27v64h27" stroke="currentColor" stroke-width="2" />
					</svg>
					<div
						class="relative grid size-[4.5rem] place-items-center rounded-[24px] bg-[var(--accent)] text-white shadow-[0_12px_32px_rgba(9,176,244,0.3)]"
					>
						<Building2 size={34} strokeWidth={1.8} aria-hidden="true" />
					</div>
					<div
						class="absolute top-5 right-4 grid size-11 place-items-center rounded-2xl border border-[var(--border)] bg-[var(--app-background)] text-[var(--accent)] shadow-lg"
					>
						<MapPin size={20} aria-hidden="true" />
					</div>
				</div>
				<h2 class="max-w-xs text-2xl font-semibold tracking-tight">
					Sua primeira arena começa aqui
				</h2>
				<p class="mt-2 max-w-xs text-sm leading-6 text-[var(--text-secondary)]">
					Cadastre seu espaço e comece a organizar os jogos.
				</p>
				<button
					class="mt-7 inline-flex min-h-12 w-full max-w-xs items-center justify-center gap-2 rounded-2xl bg-[var(--accent)] px-5 text-base font-semibold text-white transition-opacity hover:opacity-90 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
					type="button"
					onclick={startCreation}
				>
					<Plus size={19} aria-hidden="true" />
					Cadastrar primeira arena
				</button>
			</div>
		{:else if arenas.length > 0}
			<div class="grid grid-cols-1 gap-3 px-4 pb-6">
				{#each arenas as arena (arena.id)}
					<article
						class="overflow-hidden rounded-[22px] border border-[var(--border)] bg-[var(--surface)]"
					>
						<button
							class="block w-full text-left transition-colors hover:bg-[var(--surface-raised)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none focus-visible:ring-inset"
							type="button"
							aria-label={`Abrir horários da arena ${arena.name}`}
							onclick={() => openSchedule(arena)}
						>
							<div class="relative aspect-[16/9] overflow-hidden bg-[var(--surface-raised)]">
								<div class="absolute inset-0 grid place-items-center text-[var(--accent)]">
									<Building2 size={34} strokeWidth={1.5} aria-hidden="true" />
								</div>
								{#if arena.photo_url}
									<img
										src={arena.photo_url}
										alt=""
										class="absolute inset-0 size-full object-cover"
										onerror={(event) => event.currentTarget.remove()}
									/>
								{/if}
								<span
									class={`absolute top-2 right-2 rounded-full px-2.5 py-1 text-[10px] font-semibold ${arena.active ? 'bg-[var(--app-background)] text-[var(--accent)]' : 'bg-black/70 text-white'}`}
								>
									{arena.active ? 'Ativa' : 'Fechada'}
								</span>
							</div>
							<div class="p-3">
								<h2 class="truncate text-base font-semibold">{arena.name}</h2>
								<p class="mt-1 truncate text-xs text-[var(--text-secondary)]">
									{arena.city}, {arena.state}
								</p>
							</div>
						</button>
					</article>
				{/each}
			</div>
		{/if}
	</section>
{:else if screen === 'details'}
	<section
		class="relative z-10 flex min-h-0 flex-1 flex-col overflow-y-auto bg-[var(--app-background)]"
	>
		<header class="flex items-center gap-3 border-b border-[var(--border)] px-4 py-3">
			<button
				class="grid size-11 shrink-0 place-items-center rounded-2xl text-[var(--text-primary)] transition-colors hover:bg-[var(--surface)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				aria-label={editingArena ? 'Cancelar edição' : 'Voltar para arenas'}
				onclick={cancelArenaForm}
			>
				<ArrowLeft size={21} aria-hidden="true" />
			</button>
			<div>
				<h1 class="text-lg font-semibold">{editingArena ? 'Editar arena' : 'Nova arena'}</h1>
				<p class="text-xs text-[var(--text-secondary)]">
					{editingArena ? 'Dados da arena' : '1 de 3 · Dados do espaço'}
				</p>
			</div>
		</header>

		<form class="flex flex-1 flex-col px-5 py-5" onsubmit={showLocationPicker}>
			<div class="mb-6 h-1 overflow-hidden rounded-full bg-[var(--surface-raised)]">
				<div class="h-full w-1/3 rounded-full bg-[var(--accent)]"></div>
			</div>

			<div class="grid gap-4">
				<label class="grid gap-2 text-sm font-medium">
					<span>Nome da arena</span>
					<input
						class="min-h-12 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--focus-ring)]"
						name="name"
						autocomplete="organization"
						placeholder="Ex.: Arena Central"
						maxlength="120"
						bind:value={name}
						required
					/>
				</label>

				<label class="grid gap-2 text-sm font-medium">
					<span>Tipo de arena</span>
					<select
						class="min-h-12 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 text-base text-[var(--text-primary)] outline-none focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--focus-ring)]"
						name="arena_type"
						bind:value={arenaType}
					>
						{#each arenaTypes as type (type.value)}
							<option value={type.value}>{type.label}</option>
						{/each}
					</select>
				</label>

				<label class="grid gap-2 text-sm font-medium">
					<span>Endereço</span>
					<input
						class="min-h-12 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--focus-ring)]"
						name="address"
						autocomplete="street-address"
						placeholder="Rua, número e bairro"
						maxlength="255"
						bind:value={address}
						required
					/>
				</label>

				<div class="grid gap-4">
					<label class="grid gap-2 text-sm font-medium">
						<span>Cidade</span>
						<input
							class="min-h-12 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--focus-ring)]"
							name="city"
							autocomplete="address-level2"
							placeholder="Sua cidade"
							maxlength="120"
							bind:value={city}
							required
						/>
					</label>
					<label class="grid gap-2 text-sm font-medium">
						<span>UF</span>
						<select
							class="min-h-12 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 text-base text-[var(--text-primary)] outline-none focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--focus-ring)]"
							name="state"
							autocomplete="address-level1"
							bind:value={stateCode}
							required
						>
							<option value="" disabled>Selecione a UF</option>
							{#each brazilianStates as state (state.value)}
								<option value={state.value}>{state.label} ({state.value})</option>
							{/each}
						</select>
					</label>
				</div>
			</div>

			{#if formError}
				<p
					class="mt-4 rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm text-[var(--text-primary)]"
					role="alert"
				>
					{formError}
				</p>
			{/if}

			<div class="mt-auto pt-6">
				<button
					class="inline-flex min-h-12 w-full items-center justify-center gap-2 rounded-2xl bg-[var(--accent)] px-5 text-base font-semibold text-white transition-opacity hover:opacity-90 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
					type="submit"
				>
					<span>{editingArena ? 'Continuar para o mapa' : 'Escolher local no mapa'}</span>
					<ArrowRight size={19} aria-hidden="true" />
				</button>
			</div>
		</form>
	</section>
{:else if screen === 'location'}
	<section class="relative z-10 flex min-h-0 flex-1 flex-col bg-[var(--app-background)]">
		<header class="flex items-center gap-3 border-b border-[var(--border)] px-4 py-3">
			<button
				class="grid size-11 shrink-0 place-items-center rounded-2xl text-[var(--text-primary)] transition-colors hover:bg-[var(--surface)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				aria-label="Voltar para os dados da arena"
				onclick={() => (screen = 'details')}
			>
				<ArrowLeft size={21} aria-hidden="true" />
			</button>
			<div class="min-w-0 flex-1">
				<h1 class="text-lg font-semibold">Local da arena</h1>
				<p class="text-xs text-[var(--text-secondary)]">
					{editingArena ? '2 de 3 · Toque para ajustar' : '2 de 3 · Toque para marcar'}
				</p>
			</div>
			<span
				class="grid size-9 place-items-center rounded-xl bg-[var(--surface)] text-[var(--accent)]"
			>
				<MapPin size={19} aria-hidden="true" />
			</span>
		</header>

		<div class="relative flex min-h-0 flex-1">
			<MapView {theme} selectable={true} {selectedLocation} onSelectLocation={selectLocation} />
			{#if showLocationInstructions}
				<div class="absolute inset-x-4 top-4 z-[600] flex justify-center">
					<div
						class="w-full max-w-sm rounded-[24px] border border-[var(--border)] bg-[var(--surface)] p-4 shadow-xl backdrop-blur-xl"
						role="dialog"
						aria-labelledby="arena-map-instructions-title"
					>
						<div class="flex items-start gap-3">
							<span
								class="grid size-10 shrink-0 place-items-center rounded-xl bg-[var(--focus-ring)] text-[var(--accent)]"
							>
								<MapPin size={20} aria-hidden="true" />
							</span>
							<div class="min-w-0 flex-1">
								<h2 id="arena-map-instructions-title" class="font-semibold">
									Selecione o local da arena
								</h2>
								<p class="mt-1 text-sm leading-5 text-[var(--text-secondary)]">
									Toque no mapa no ponto onde a arena fica.
								</p>
							</div>
							<button
								class="grid size-9 shrink-0 place-items-center rounded-xl text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-raised)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
								type="button"
								aria-label="Entendi"
								onclick={() => (showLocationInstructions = false)}
							>
								<X size={18} aria-hidden="true" />
							</button>
						</div>
					</div>
				</div>
			{/if}
		</div>

		<footer
			class="border-t border-[var(--border)] bg-[var(--app-background)] px-5 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]"
		>
			{#if formError}
				<p class="mb-3 flex items-center gap-2 text-sm text-[var(--text-primary)]" role="alert">
					<AlertCircle size={17} class="shrink-0 text-[var(--accent)]" aria-hidden="true" />
					{formError}
				</p>
			{/if}
			{#if selectedLocation}
				<p
					class="mb-3 flex items-center gap-2 text-xs text-[var(--text-secondary)]"
					aria-live="polite"
				>
					<MapPin size={15} class="shrink-0 text-[var(--accent)]" aria-hidden="true" />
					<span class="truncate"
						>{selectedLocation.latitude.toFixed(5)}, {selectedLocation.longitude.toFixed(5)}</span
					>
				</p>
			{:else if editingArena}
				<p class="mb-3 text-xs text-[var(--text-secondary)]">Nenhum local informado · opcional</p>
			{/if}
			<button
				class="inline-flex min-h-12 w-full items-center justify-center gap-2 rounded-2xl bg-[var(--accent)] px-5 text-base font-semibold text-white transition-opacity hover:opacity-90 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-45"
				type="button"
				disabled={!editingArena && !selectedLocation}
				onclick={showPhotoStep}
			>
				<span>Continuar para a foto</span>
				<ArrowRight size={19} aria-hidden="true" />
			</button>
		</footer>
	</section>
{:else if screen === 'photo'}
	<section class="relative z-10 flex min-h-0 flex-1 flex-col bg-[var(--app-background)]">
		<header class="flex items-center gap-3 border-b border-[var(--border)] px-4 py-3">
			<button
				class="grid size-11 shrink-0 place-items-center rounded-2xl text-[var(--text-primary)] transition-colors hover:bg-[var(--surface)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none disabled:opacity-40"
				type="button"
				aria-label="Voltar para o mapa"
				disabled={pendingPhotoArena !== null}
				onclick={() => (screen = 'location')}
			>
				<ArrowLeft size={21} aria-hidden="true" />
			</button>
			<div class="min-w-0 flex-1">
				<h1 class="text-lg font-semibold">Foto da arena</h1>
				<p class="text-xs text-[var(--text-secondary)]">3 de 3 · Opcional</p>
			</div>
			<span
				class="grid size-9 place-items-center rounded-xl bg-[var(--surface)] text-[var(--accent)]"
			>
				<ImagePlus size={19} aria-hidden="true" />
			</span>
		</header>

		<div class="min-h-0 flex-1 overflow-y-auto px-5 py-5">
			<div class="mx-auto grid w-full max-w-sm gap-4">
				<div>
					<h2 class="text-xl font-semibold tracking-tight">Mostre sua arena</h2>
					<p class="mt-1 text-sm leading-6 text-[var(--text-secondary)]">
						Adicione uma foto para os jogadores reconhecerem o espaço.
					</p>
				</div>

				<div
					class="relative grid aspect-[4/3] place-items-center overflow-hidden rounded-[28px] border border-[var(--border)] bg-[var(--surface)] text-[var(--accent)]"
				>
					{#if photoPreview || currentArenaPhotoURL}
						<img
							src={photoPreview || currentArenaPhotoURL || ''}
							alt={photoPreview ? 'Pré-visualização da foto da arena' : `Foto atual de ${name}`}
							class="absolute inset-0 size-full object-cover"
						/>
					{:else}
						<div class="grid size-20 place-items-center rounded-[26px] bg-[var(--app-background)]">
							<ImagePlus size={36} strokeWidth={1.5} aria-hidden="true" />
						</div>
					{/if}
				</div>

				<div class="flex flex-wrap items-center gap-2">
					<label
						class="inline-flex min-h-11 cursor-pointer items-center justify-center rounded-2xl bg-[var(--surface)] px-4 text-sm font-semibold text-[var(--accent)] transition-colors focus-within:ring-4 focus-within:ring-[var(--focus-ring)] hover:bg-[var(--surface-raised)]"
					>
						{#key photoInputKey}
							<input
								class="sr-only"
								type="file"
								name="photo"
								accept="image/jpeg,image/png"
								onchange={selectPhoto}
							/>
						{/key}
						<span>{selectedPhoto || currentArenaPhotoURL ? 'Trocar foto' : 'Escolher foto'}</span>
					</label>
					{#if selectedPhoto}
						<button
							class="inline-flex min-h-11 items-center gap-2 rounded-2xl px-3 text-sm font-medium text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
							type="button"
							onclick={clearPhoto}
						>
							<X size={15} aria-hidden="true" />
							Remover foto
						</button>
					{/if}
				</div>
				<p class="text-xs text-[var(--text-secondary)]">JPEG ou PNG · até 5 MB</p>
				{#if photoError}
					<p class="text-sm text-[var(--text-primary)]" role="alert">{photoError}</p>
				{/if}
			</div>
		</div>

		<footer
			class="border-t border-[var(--border)] bg-[var(--app-background)] px-5 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]"
		>
			{#if formError}
				<p class="mb-3 flex items-center gap-2 text-sm text-[var(--text-primary)]" role="alert">
					<AlertCircle size={17} class="shrink-0 text-[var(--accent)]" aria-hidden="true" />
					{formError}
				</p>
			{/if}
			<button
				class="inline-flex min-h-12 w-full items-center justify-center gap-2 rounded-2xl bg-[var(--accent)] px-5 text-base font-semibold text-white transition-opacity hover:opacity-90 focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-45"
				type="button"
				disabled={isSaving || (pendingPhotoArena !== null && !selectedPhoto)}
				onclick={saveArena}
			>
				{#if isSaving}
					<LoaderCircle
						size={19}
						class="animate-spin motion-reduce:animate-none"
						aria-hidden="true"
					/>
					<span
						>{pendingPhotoArena
							? 'Enviando foto'
							: editingArena
								? 'Salvando alterações'
								: 'Cadastrando arena'}</span
					>
				{:else}
					<Check size={19} aria-hidden="true" />
					<span>
						{pendingPhotoArena
							? 'Tentar enviar foto'
							: editingArena
								? 'Salvar alterações'
								: selectedPhoto
									? 'Cadastrar arena'
									: 'Cadastrar sem foto'}
					</span>
				{/if}
			</button>
			{#if pendingPhotoArena}
				<button
					class="mt-2 min-h-11 w-full rounded-2xl px-4 text-sm font-medium text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface)] hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
					type="button"
					onclick={finishWithoutPhoto}
				>
					Concluir sem foto
				</button>
			{/if}
		</footer>
	</section>
{:else if screen === 'schedule' && selectedArena}
	<ArenaSchedule
		arena={selectedArena}
		onBack={closeSchedule}
		onEdit={() => startEditing(selectedArena, 'schedule')}
		onArenaUpdated={updateSelectedArena}
	/>
{/if}
