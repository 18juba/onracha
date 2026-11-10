<script lang="ts">
	import { CalendarDays, ChevronLeft, ChevronRight } from '@lucide/svelte';
	import { prefersReducedMotion } from 'svelte/motion';
	import { fly, scale } from 'svelte/transition';
	import Button from './Button.svelte';

	type DayCell = { key: string; value: string | null; day: number; inMonth: boolean };
	type Props = {
		value?: string;
		min?: string;
		id?: string;
		name?: string;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		class?: string;
	};

	const generatedId = $props.id();
	let {
		value = $bindable(''),
		min,
		id = generatedId,
		name,
		placeholder = 'Escolha uma data',
		disabled = false,
		required = false,
		class: className = ''
	}: Props = $props();
	let isOpen = $state(false);
	let month = $state(new Date());
	let root: HTMLDivElement;
	const weekDays = [
		{ short: 'S', full: 'Segunda-feira' },
		{ short: 'T', full: 'Terça-feira' },
		{ short: 'Q', full: 'Quarta-feira' },
		{ short: 'Q', full: 'Quinta-feira' },
		{ short: 'S', full: 'Sexta-feira' },
		{ short: 'S', full: 'Sábado' },
		{ short: 'D', full: 'Domingo' }
	];
	let days = $derived.by((): DayCell[] => {
		const first = new Date(month.getFullYear(), month.getMonth(), 1);
		const offset = (first.getDay() + 6) % 7;
		return Array.from({ length: 42 }, (_, index) => {
			const date = new Date(month.getFullYear(), month.getMonth(), index - offset + 1);
			const inMonth = date.getMonth() === month.getMonth();
			return {
				key: localDateValue(date),
				value: inMonth ? localDateValue(date) : null,
				day: date.getDate(),
				inMonth
			};
		});
	});
	let monthLabel = $derived(
		new Intl.DateTimeFormat('pt-BR', { month: 'long', year: 'numeric' }).format(month)
	);
	let displayValue = $derived(value ? formatDate(value) : placeholder);

	function localDateValue(date: Date) {
		const year = date.getFullYear();
		const monthNumber = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return year + '-' + monthNumber + '-' + day;
	}

	function parseLocalDate(value: string) {
		const parts = value.split('-').map(Number);
		if (parts.length !== 3 || parts.some(Number.isNaN)) return null;
		return new Date(parts[0], parts[1] - 1, parts[2]);
	}

	function formatDate(value: string) {
		const date = parseLocalDate(value);
		return date ? new Intl.DateTimeFormat('pt-BR', { dateStyle: 'long' }).format(date) : value;
	}

	function chooseDate(dateValue: string) {
		value = dateValue;
		isOpen = false;
	}

	function shiftMonth(amount: number) {
		month = new Date(month.getFullYear(), month.getMonth() + amount, 1);
	}

	function toggleCalendar() {
		if (disabled) return;
		if (isOpen) {
			isOpen = false;
			return;
		}
		const selectedDate = parseLocalDate(value);
		if (selectedDate) month = new Date(selectedDate.getFullYear(), selectedDate.getMonth(), 1);
		isOpen = true;
	}

	function closeOnOutside(event: PointerEvent) {
		if (isOpen && root && event.target instanceof Node && !root.contains(event.target))
			isOpen = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (
			event.key === 'Escape' &&
			isOpen &&
			root &&
			event.target instanceof Node &&
			root.contains(event.target)
		) {
			isOpen = false;
		}
	}
</script>

<svelte:window onpointerdown={closeOnOutside} onkeydown={handleKeydown} />

<div
	class={'relative w-full ' + className}
	bind:this={root}
	role="group"
	aria-label="Seletor de data"
>
	<Button
		{id}
		variant="outline"
		class="w-full justify-between text-left font-normal"
		aria-haspopup="dialog"
		aria-expanded={isOpen}
		aria-required={required}
		{disabled}
		onclick={toggleCalendar}
	>
		<span
			class={'min-w-0 flex-1 text-left ' +
				(value ? 'truncate' : 'truncate text-[var(--text-muted)]')}>{displayValue}</span
		>
		<CalendarDays size={16} class="shrink-0 text-[var(--text-muted)]" aria-hidden="true" />
	</Button>
	{#if name}<input type="hidden" {name} {value} {disabled} />{/if}
	{#if isOpen}
		<dialog
			open
			aria-label="Selecione uma data"
			transition:fly={{ y: -5, duration: prefersReducedMotion.current ? 0 : 150 }}
			class="absolute inset-x-0 top-full z-50 m-0 mt-2 w-full max-w-none rounded-xl border border-[var(--border)] bg-[var(--surface-raised)] p-3 text-[var(--text-primary)] shadow-none"
		>
			<header class="mb-3 flex items-center justify-between gap-2">
				<Button
					variant="ghost"
					size="icon"
					aria-label="Mês anterior"
					onclick={() => shiftMonth(-1)}
				>
					<ChevronLeft size={16} aria-hidden="true" />
				</Button>
				<h2 class="text-sm font-medium capitalize text-[var(--text-primary)]">{monthLabel}</h2>
				<Button variant="ghost" size="icon" aria-label="Próximo mês" onclick={() => shiftMonth(1)}>
					<ChevronRight size={16} aria-hidden="true" />
				</Button>
			</header>
			<div class="mb-1 grid grid-cols-7 gap-1 text-center text-[11px] text-[var(--text-muted)]">
				{#each weekDays as weekday (weekday.full)}
					<span aria-label={weekday.full}>{weekday.short}</span>
				{/each}
			</div>
			<div class="grid grid-cols-7 gap-1">
				{#each days as day (day.key)}
					{#if day.value}
						{@const isSelected = day.value === value}
						{@const isToday = day.value === localDateValue(new Date())}
						{@const isBeforeMin = Boolean(min && day.value < min)}
						<Button
							variant={isSelected ? 'primary' : 'ghost'}
							size="none"
							class={'size-9 rounded-lg p-0 text-xs ' +
								(isSelected
									? 'bg-[var(--accent)] text-[var(--accent-ink)] hover:bg-[var(--accent-hover)]'
									: isToday
										? 'border border-[var(--border)] text-[var(--text-primary)] hover:bg-[var(--surface-raised)]'
										: 'text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]')}
							aria-label={formatDate(day.value)}
							aria-pressed={isSelected}
							disabled={isBeforeMin}
							onclick={() => chooseDate(day.value ?? '')}
						>
							{#if isSelected}
								<span
									transition:scale={{
										duration: prefersReducedMotion.current ? 0 : 110,
										start: 0.8
									}}>{day.day}</span
								>
							{:else}{day.day}{/if}
						</Button>
					{:else}
						<span class="size-9" aria-hidden="true"></span>
					{/if}
				{/each}
			</div>
			<div class="mt-3 flex justify-end border-t border-[var(--border)] pt-2">
				<Button
					variant="ghost"
					size="sm"
					disabled={Boolean(min && localDateValue(new Date()) < min)}
					onclick={() => chooseDate(localDateValue(new Date()))}
				>
					Hoje
				</Button>
			</div>
		</dialog>
	{/if}
</div>
