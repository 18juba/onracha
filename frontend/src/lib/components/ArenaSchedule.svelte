<script lang="ts">
	import Button from './ui/Button.svelte';
	import DatePicker from './ui/DatePicker.svelte';
	import FormField from './ui/FormField.svelte';
	import Switch from './ui/Switch.svelte';
	import TimePicker from './ui/TimePicker.svelte';
	import { onMount } from 'svelte';
	import {
		AlertCircle,
		ArrowLeft,
		CalendarPlus,
		Check,
		Clock3,
		LoaderCircle,
		Pencil
	} from '@lucide/svelte';
	import { api, type Arena, type ManagedTimeSlot } from '../api';
	import { getUserFacingErrorMessage } from '../errors';

	type Props = {
		arena: Arena;
		onBack: () => void;
		onEdit: () => void;
		onArenaUpdated: (arena: Arena) => void;
	};

	let { arena, onBack, onEdit, onArenaUpdated }: Props = $props();
	let slots = $state<ManagedTimeSlot[]>([]);
	let isLoading = $state(true);
	let isSavingSlot = $state(false);
	let closingSlotID = $state<number | null>(null);
	let isUpdatingActive = $state(false);
	let today = $state('');
	let slotDate = $state('');
	let startTime = $state('');
	let endTime = $state('');
	let loadError = $state('');
	let formError = $state('');
	let activeError = $state('');
	let slotActionError = $state('');
	let successMessage = $state('');

	onMount(() => {
		today = localDateValue(new Date());
		slotDate = today;
		startTime = '19:00';
		endTime = '20:00';
		void loadSlots();
	});

	function localDateValue(date: Date) {
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function sortManagedSlots(values: ManagedTimeSlot[]) {
		return [...values].sort((first, second) => {
			const firstEnded = Date.parse(first.ends_at) <= Date.now();
			const secondEnded = Date.parse(second.ends_at) <= Date.now();
			if (firstEnded !== secondEnded) return Number(firstEnded) - Number(secondEnded);
			return Date.parse(first.starts_at) - Date.parse(second.starts_at);
		});
	}

	async function loadSlots() {
		isLoading = true;
		loadError = '';

		try {
			slots = sortManagedSlots(await api.listManagedArenaTimeSlots(arena.id));
		} catch (error) {
			loadError = getUserFacingErrorMessage(error, 'Não foi possível carregar os horários.');
		} finally {
			isLoading = false;
		}
	}

	async function createSlot(event: SubmitEvent) {
		event.preventDefault();
		formError = '';
		successMessage = '';

		const startsAt = new Date(`${slotDate}T${startTime}:00`);
		const endsAt = new Date(`${slotDate}T${endTime}:00`);
		if (Number.isNaN(startsAt.getTime()) || Number.isNaN(endsAt.getTime())) {
			formError = 'Informe uma data e horários válidos.';
			return;
		}
		if (startsAt.getTime() <= Date.now()) {
			formError = 'Escolha um horário que ainda não tenha começado.';
			return;
		}
		if (endsAt <= startsAt) {
			formError = 'O horário de término precisa ser depois do início.';
			return;
		}

		isSavingSlot = true;
		try {
			const createdSlot = await api.createArenaTimeSlot(arena.id, {
				starts_at: startsAt.toISOString(),
				ends_at: endsAt.toISOString()
			});
			const existingSlot = slots.find((slot) => slot.id === createdSlot.id);
			const openedSlot = { ...createdSlot, game_id: existingSlot?.game_id ?? null };
			slots = sortManagedSlots([...slots.filter((slot) => slot.id !== createdSlot.id), openedSlot]);
			successMessage = existingSlot ? 'Horário reaberto.' : 'Horário liberado para novos jogos.';
			startTime = '';
			endTime = '';
		} catch (error) {
			formError = getUserFacingErrorMessage(error, 'Não foi possível liberar esse horário.');
		} finally {
			isSavingSlot = false;
		}
	}

	async function toggleActive(_nextActive: boolean): Promise<boolean> {
		activeError = '';
		isUpdatingActive = true;

		try {
			const updatedArena = await api.toggleArenaActive(arena.id);
			onArenaUpdated(updatedArena);
			return true;
		} catch (error) {
			activeError = getUserFacingErrorMessage(
				error,
				'Não foi possível atualizar o estado da arena.'
			);
			return false;
		} finally {
			isUpdatingActive = false;
		}
	}

	async function closeTimeSlot(slot: ManagedTimeSlot) {
		if (slot.closed_at || closingSlotID !== null) return;
		const confirmed = window.confirm(
			'Encerrar este horário? Ele deixará de aparecer para os jogadores e ninguém poderá agendar ou entrar nele.'
		);
		if (!confirmed) return;

		slotActionError = '';
		closingSlotID = slot.id;
		try {
			const closedSlot = await api.closeArenaTimeSlot(arena.id, slot.id);
			slots = slots.map((current) =>
				current.id === slot.id ? { ...current, ...closedSlot } : current
			);
		} catch (error) {
			slotActionError = getUserFacingErrorMessage(error, 'Não foi possível encerrar esse horário.');
		} finally {
			closingSlotID = null;
		}
	}

	function timeSlotStatus(slot: ManagedTimeSlot) {
		if (slot.closed_at) return 'Encerrado';
		const startsAt = Date.parse(slot.starts_at);
		const endsAt = Date.parse(slot.ends_at);
		if (endsAt <= Date.now()) return 'Passado';
		if (startsAt <= Date.now()) return 'Em andamento';
		return slot.game_id ? 'Jogo agendado' : 'Disponível';
	}

	function formatDate(value: string) {
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

<section
	class="relative z-10 flex min-h-0 flex-1 flex-col overflow-y-auto bg-[var(--app-background)]"
>
	<header class="shrink-0 border-b border-[var(--border)] bg-[var(--app-background)]">
		<div class="flex items-center gap-3 px-4 pt-4 pb-3">
			<Button
				type="button"
				aria-label="Voltar para arenas"
				onclick={onBack}
				variant="secondary"
				size="icon"
				class="shrink-0"
			>
				<ArrowLeft size={21} aria-hidden="true" />
			</Button>
			<div class="min-w-0 flex-1">
				<h1 class="truncate text-lg font-semibold">{arena.name}</h1>
				<p class="truncate text-xs text-[var(--text-secondary)]">{arena.city}, {arena.state}</p>
			</div>
		</div>
		<div class="flex items-center justify-between gap-3 px-5 pb-3">
			<Switch
				checked={arena.active}
				label="Ativo"
				disabled={isUpdatingActive}
				onCheckedChange={toggleActive}
			/>
			<Button
				type="button"
				onclick={onEdit}
				variant="outline"
				size="default"
				class="justify-center gap-2"
			>
				<Pencil size={15} aria-hidden="true" />
				Editar arena
			</Button>
		</div>
	</header>

	<div class="grid gap-5 px-5 py-5">
		<div>
			<h2 class="text-xl font-semibold tracking-tight">Liberar horários</h2>
			<p class="mt-1 text-sm text-[var(--text-secondary)]">
				Os jogadores poderão marcar jogos nesses períodos.
			</p>
		</div>

		{#if activeError}
			<p class="rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm" role="alert">{activeError}</p>
		{/if}

		<form
			class="grid gap-3 rounded-[24px] border border-[var(--border)] bg-[var(--surface)] p-4"
			onsubmit={createSlot}
		>
			<FormField label="Data" for="slot-date" required>
				<DatePicker id="slot-date" name="date" min={today} bind:value={slotDate} required />
			</FormField>
			<div class="grid grid-cols-2 gap-3">
				<FormField label="Início" for="slot-start-time">
					<TimePicker id="slot-start-time" name="starts_at" bind:value={startTime} required />
				</FormField>
				<FormField label="Término" for="slot-end-time">
					<TimePicker id="slot-end-time" name="ends_at" bind:value={endTime} required />
				</FormField>
			</div>
			{#if formError}
				<p class="text-sm text-[var(--text-primary)]" role="alert">{formError}</p>
			{/if}
			{#if successMessage}
				<p class="flex items-center gap-2 text-sm text-[var(--text-primary)]" role="status">
					<Check size={16} class="text-[var(--accent)]" aria-hidden="true" />
					{successMessage}
				</p>
			{/if}
			<Button
				type="submit"
				disabled={isSavingSlot || !slotDate || !startTime || !endTime}
				variant="primary"
				size="lg"
				class="justify-center gap-2"
			>
				{#if isSavingSlot}
					<LoaderCircle
						size={18}
						class="animate-spin motion-reduce:animate-none"
						aria-hidden="true"
					/>
					<span>Liberando horário</span>
				{:else}
					<CalendarPlus size={18} aria-hidden="true" />
					<span>Liberar horário</span>
				{/if}
			</Button>
		</form>

		<section class="grid gap-3" aria-labelledby="available-slots-title">
			<div class="flex items-center justify-between gap-3">
				<h2 id="available-slots-title" class="text-lg font-semibold">Horários cadastrados</h2>
				{#if !isLoading && slots.length > 0}
					<span class="text-xs text-[var(--text-secondary)]">{slots.length}</span>
				{/if}
			</div>
			{#if slotActionError}
				<p class="rounded-2xl bg-[var(--surface)] px-4 py-3 text-sm" role="alert">
					{slotActionError}
				</p>
			{/if}

			{#if loadError}
				<div class="rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-4" role="alert">
					<p class="flex items-start gap-2 text-sm text-[var(--text-primary)]">
						<AlertCircle
							size={18}
							class="mt-0.5 shrink-0 text-[var(--accent)]"
							aria-hidden="true"
						/>
						{loadError}
					</p>
					<Button type="button" onclick={loadSlots} variant="ghost" size="sm" class="mt-2">
						Tentar novamente
					</Button>
				</div>
			{:else if isLoading}
				<div class="grid min-h-24 place-items-center" role="status" aria-live="polite">
					<LoaderCircle
						size={24}
						class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
						aria-label="Carregando horários"
					/>
				</div>
			{:else if slots.length === 0}
				<div
					class="grid justify-items-center gap-2 rounded-[24px] border border-dashed border-[var(--border)] bg-[var(--surface)] px-5 py-7 text-center"
				>
					<div
						class="grid size-12 place-items-center rounded-2xl bg-[var(--app-background)] text-[var(--accent)]"
					>
						<Clock3 size={23} aria-hidden="true" />
					</div>
					<p class="font-semibold">Nenhum horário cadastrado</p>
					<p class="max-w-xs text-sm text-[var(--text-secondary)]">
						Adicione uma data e um período para liberar sua arena para jogos.
					</p>
				</div>
			{:else}
				<div class="grid gap-2">
					{#each slots as slot (slot.id)}
						<article
							class="flex items-center justify-between gap-3 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3"
						>
							<div class="grid min-w-0 gap-1">
								<h3 class="font-semibold capitalize">{formatDate(slot.starts_at)}</h3>
								<p class="text-sm text-[var(--text-secondary)]">
									{formatTime(slot.starts_at)} – {formatTime(slot.ends_at)}
								</p>
								<p class="text-xs text-[var(--text-secondary)]">{timeSlotStatus(slot)}</p>
							</div>
							{#if slot.closed_at}
								<Check size={17} class="shrink-0 text-[var(--accent)]" aria-hidden="true" />
							{:else}
								<Button
									type="button"
									disabled={closingSlotID !== null}
									aria-label={`Encerrar horário de ${formatDate(slot.starts_at)} das ${formatTime(slot.starts_at)} às ${formatTime(slot.ends_at)}`}
									onclick={() => closeTimeSlot(slot)}
									variant="destructive"
									size="sm"
									class="shrink-0 gap-1.5"
								>
									{#if closingSlotID === slot.id}
										<LoaderCircle
											size={15}
											class="animate-spin motion-reduce:animate-none"
											aria-hidden="true"
										/>
									{:else}
										<Clock3 size={15} aria-hidden="true" />
									{/if}
									<span>Encerrar</span>
								</Button>
							{/if}
						</article>
					{/each}
				</div>
			{/if}
		</section>
	</div>
</section>
