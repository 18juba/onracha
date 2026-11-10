<script lang="ts">
	import { ChevronDown, Clock3 } from '@lucide/svelte';
	import { tick } from 'svelte';
	import { prefersReducedMotion } from 'svelte/motion';
	import { fly } from 'svelte/transition';
	import Button from './Button.svelte';

	type Props = {
		value?: string;
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
		id = generatedId,
		name,
		placeholder = 'Escolha um horário',
		disabled = false,
		required = false,
		class: className = ''
	}: Props = $props();
	let isOpen = $state(false);
	let draftHour = $state('19');
	let draftMinute = $state('00');
	let root: HTMLDivElement;
	const hours = Array.from({ length: 24 }, (_, index) => String(index).padStart(2, '0'));
	const minutes = Array.from({ length: 60 }, (_, index) => String(index).padStart(2, '0'));
	let displayValue = $derived(value || placeholder);

	async function openPicker() {
		if (disabled) return;
		const [hour, minute] = value.split(':');
		draftHour = hour || '19';
		draftMinute = minute || '00';
		isOpen = true;
		await tick();
		document.getElementById(optionID('hour', draftHour))?.focus();
	}

	function optionID(kind: 'hour' | 'minute', part: string) {
		return id + '-' + kind + '-' + part;
	}

	function chooseHour(hour: string) {
		draftHour = hour;
		value = hour + ':' + draftMinute;
	}

	function chooseMinute(minute: string) {
		draftMinute = minute;
		value = draftHour + ':' + minute;
	}

	function moveOption(event: KeyboardEvent, kind: 'hour' | 'minute', index: number) {
		const items = kind === 'hour' ? hours : minutes;
		let nextIndex: number;
		if (event.key === 'ArrowDown') nextIndex = (index + 1) % items.length;
		else if (event.key === 'ArrowUp') nextIndex = (index - 1 + items.length) % items.length;
		else if (event.key === 'Home') nextIndex = 0;
		else if (event.key === 'End') nextIndex = items.length - 1;
		else return;

		event.preventDefault();
		const nextValue = items[nextIndex];
		if (kind === 'hour') chooseHour(nextValue);
		else chooseMinute(nextValue);
		document.getElementById(optionID(kind, nextValue))?.focus();
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
	aria-label="Seletor de horário"
>
	<Button
		{id}
		variant="outline"
		class="w-full justify-between text-left font-normal"
		aria-haspopup="dialog"
		aria-expanded={isOpen}
		aria-required={required}
		{disabled}
		onclick={() => (isOpen ? (isOpen = false) : openPicker())}
	>
		<span
			class={'min-w-0 flex-1 text-left ' +
				(value ? 'truncate' : 'truncate text-[var(--text-muted)]')}>{displayValue}</span
		>
		<span class="flex shrink-0 items-center gap-2 text-[var(--text-muted)]">
			<Clock3 size={16} aria-hidden="true" />
			<ChevronDown
				size={14}
				class={'transition-transform duration-150 ' + (isOpen ? 'rotate-180' : '')}
				aria-hidden="true"
			/>
		</span>
	</Button>
	{#if name}<input type="hidden" {name} {value} {disabled} />{/if}
	{#if isOpen}
		<dialog
			open
			aria-label="Selecione um horário"
			transition:fly={{ y: -5, duration: prefersReducedMotion.current ? 0 : 150 }}
			class="absolute inset-x-0 top-full z-50 m-0 mt-2 w-full max-w-none rounded-xl border border-[var(--border)] bg-[var(--surface-raised)] p-3 text-[var(--text-primary)] shadow-none"
		>
			<div class="mb-2 flex items-center justify-between">
				<h2 class="text-sm font-medium text-[var(--text-primary)]">Escolha o horário</h2>
				<span class="font-mono text-sm text-[var(--text-secondary)]">{draftHour}:{draftMinute}</span
				>
			</div>
			<div class="grid grid-cols-2 gap-3">
				<div>
					<p class="mb-1 text-center text-[11px] text-[var(--text-muted)]">Hora</p>
					<div
						class="max-h-40 overflow-y-auto rounded-lg border border-[var(--border)] p-1"
						role="listbox"
						aria-label="Hora"
					>
						{#each hours as hour (hour)}
							<Button
								variant={draftHour === hour ? 'secondary' : 'ghost'}
								size="sm"
								class="w-full justify-center font-mono"
								id={optionID('hour', hour)}
								role="option"
								aria-selected={draftHour === hour}
								tabindex={draftHour === hour ? 0 : -1}
								onkeydown={(event) => moveOption(event, 'hour', Number(hour))}
								onclick={() => chooseHour(hour)}>{hour}</Button
							>
						{/each}
					</div>
				</div>
				<div>
					<p class="mb-1 text-center text-[11px] text-[var(--text-muted)]">Minuto</p>
					<div
						class="max-h-40 overflow-y-auto rounded-lg border border-[var(--border)] p-1"
						role="listbox"
						aria-label="Minuto"
					>
						{#each minutes as minute (minute)}
							<Button
								variant={draftMinute === minute ? 'secondary' : 'ghost'}
								size="sm"
								class="w-full justify-center font-mono"
								id={optionID('minute', minute)}
								role="option"
								aria-selected={draftMinute === minute}
								tabindex={draftMinute === minute ? 0 : -1}
								onkeydown={(event) => moveOption(event, 'minute', Number(minute))}
								onclick={() => chooseMinute(minute)}>{minute}</Button
							>
						{/each}
					</div>
				</div>
			</div>
			<div class="mt-3 flex justify-end border-t border-[var(--border)] pt-2">
				<Button variant="primary" size="sm" onclick={() => (isOpen = false)}>Concluir</Button>
			</div>
		</dialog>
	{/if}
</div>
