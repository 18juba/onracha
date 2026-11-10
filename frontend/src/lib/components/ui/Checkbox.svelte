<script lang="ts">
	import { Check } from '@lucide/svelte';
	import { prefersReducedMotion } from 'svelte/motion';
	import { scale } from 'svelte/transition';
	import type { HTMLInputAttributes } from 'svelte/elements';

	type Props = Omit<HTMLInputAttributes, 'checked' | 'class' | 'type'> & {
		checked?: boolean;
		label: string;
		class?: string;
	};

	const generatedId = $props.id();
	let {
		checked = $bindable(false),
		label,
		id = generatedId,
		class: className = '',
		disabled = false,
		...rest
	}: Props = $props();
</script>

<label
	for={id}
	class={'inline-flex min-h-11 cursor-pointer items-center gap-2.5 text-sm text-[var(--text-primary)] ' +
		(disabled ? 'cursor-not-allowed opacity-50 ' : '') +
		className}
>
	<input {id} type="checkbox" bind:checked {disabled} {...rest} class="peer sr-only" />
	<span
		class="grid size-[18px] shrink-0 place-items-center rounded-[5px] border border-[var(--border)] bg-[var(--field-background)] text-[var(--accent-ink)] transition-[background-color,border-color,box-shadow] peer-checked:border-[var(--accent)] peer-checked:bg-[var(--accent)] peer-focus-visible:ring-2 peer-focus-visible:ring-[var(--focus-ring)] peer-focus-visible:ring-offset-2 peer-focus-visible:ring-offset-[var(--app-background)] motion-reduce:transition-none"
		aria-hidden="true"
	>
		{#if checked}
			<span transition:scale={{ duration: prefersReducedMotion.current ? 0 : 120, start: 0.65 }}>
				<Check size={13} strokeWidth={2.5} />
			</span>
		{/if}
	</span>
	<span>{label}</span>
</label>
