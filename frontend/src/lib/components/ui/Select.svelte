<script lang="ts">
	import { Check, ChevronDown } from '@lucide/svelte';
	import { prefersReducedMotion } from 'svelte/motion';
	import { fly, scale } from 'svelte/transition';
	import Button from './Button.svelte';

	export type SelectOption = { value: string; label: string };
	type Props = {
		value?: string;
		options: SelectOption[];
		label: string;
		placeholder?: string;
		name?: string;
		id?: string;
		required?: boolean;
		disabled?: boolean;
		class?: string;
	};

	const generatedId = $props.id();
	let {
		value = $bindable(''),
		options,
		label,
		placeholder = 'Selecione uma opção',
		name,
		id = generatedId,
		required = false,
		disabled = false,
		class: className = ''
	}: Props = $props();
	let isOpen = $state(false);
	let activeIndex = $state(0);
	let root: HTMLDivElement;
	let selectedIndex = $derived(options.findIndex((option) => option.value === value));
	let selectedOption = $derived(selectedIndex >= 0 ? options[selectedIndex] : undefined);

	function optionID(index: number) {
		return id + '-option-' + index;
	}

	function open() {
		if (disabled || options.length === 0) return;
		activeIndex = selectedIndex >= 0 ? selectedIndex : 0;
		isOpen = true;
	}

	function choose(option: SelectOption) {
		value = option.value;
		isOpen = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!isOpen && ['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(event.key)) {
			event.preventDefault();
			open();
			return;
		}
		if (!isOpen) {
			if (event.key.length === 1 && /[a-z0-9]/i.test(event.key)) {
				const found = options.findIndex((option) =>
					option.label.toLowerCase().startsWith(event.key.toLowerCase())
				);
				if (found >= 0) {
					value = options[found].value;
					activeIndex = found;
				}
			}
			return;
		}

		if (event.key === 'Escape' || event.key === 'Tab') {
			isOpen = false;
			return;
		}
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			activeIndex = (activeIndex + 1) % options.length;
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			activeIndex = (activeIndex - 1 + options.length) % options.length;
		} else if (event.key === 'Home') {
			event.preventDefault();
			activeIndex = 0;
		} else if (event.key === 'End') {
			event.preventDefault();
			activeIndex = options.length - 1;
		} else if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			if (options[activeIndex]) choose(options[activeIndex]);
		} else if (event.key.length === 1 && /[a-z0-9]/i.test(event.key)) {
			const found = options.findIndex((option) =>
				option.label.toLowerCase().startsWith(event.key.toLowerCase())
			);
			if (found >= 0) activeIndex = found;
		}
	}

	function closeOnOutside(event: PointerEvent) {
		if (isOpen && root && event.target instanceof Node && !root.contains(event.target))
			isOpen = false;
	}
</script>

<svelte:window onpointerdown={closeOnOutside} />

<div class={'relative w-full ' + className} bind:this={root}>
	<Button
		{id}
		variant="outline"
		size="default"
		class="w-full justify-between text-left font-normal"
		role="combobox"
		aria-haspopup="listbox"
		aria-controls={id + '-listbox'}
		aria-expanded={isOpen}
		aria-required={required}
		aria-activedescendant={isOpen && options.length > 0 ? optionID(activeIndex) : undefined}
		aria-label={label}
		{disabled}
		onclick={() => (isOpen ? (isOpen = false) : open())}
		onkeydown={handleKeydown}
	>
		<span
			class={'min-w-0 flex-1 text-left ' +
				(selectedOption ? 'truncate' : 'truncate text-[var(--text-muted)]')}
			>{selectedOption?.label ?? placeholder}</span
		>
		<ChevronDown
			size={16}
			class={'shrink-0 transition-transform duration-150 ' + (isOpen ? 'rotate-180' : '')}
			aria-hidden="true"
		/>
	</Button>
	{#if name}<input type="hidden" {name} {value} {disabled} />{/if}
	{#if isOpen}
		<div
			id={id + '-listbox'}
			role="listbox"
			aria-label={label}
			transition:fly={{ y: -5, duration: prefersReducedMotion.current ? 0 : 140 }}
			class="absolute inset-x-0 top-full z-50 mt-2 max-h-64 overflow-y-auto rounded-xl border border-[var(--border)] bg-[var(--surface-raised)] p-1.5 shadow-none"
		>
			{#each options as option, index (option.value)}
				<Button
					id={optionID(index)}
					type="button"
					variant="ghost"
					size="sm"
					role="option"
					aria-selected={option.value === value}
					tabindex="-1"
					onclick={() => choose(option)}
					onpointerenter={() => (activeIndex = index)}
					class={'w-full justify-start rounded-lg text-left text-sm transition-colors ' +
						(activeIndex === index
							? 'bg-[var(--surface-raised)] text-[var(--text-primary)]'
							: 'text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]')}
				>
					<span class="min-w-0 flex-1 text-left">{option.label}</span>
					{#if option.value === value}
						<span
							class="ml-auto shrink-0"
							transition:scale={{ duration: prefersReducedMotion.current ? 0 : 100, start: 0.7 }}
						>
							<Check size={15} class="text-[var(--accent)]" aria-hidden="true" />
						</span>
					{/if}
				</Button>
			{/each}
		</div>
	{/if}
</div>
