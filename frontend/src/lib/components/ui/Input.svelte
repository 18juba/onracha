<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	type Props = Omit<HTMLInputAttributes, 'class' | 'value'> & {
		value?: string;
		class?: string;
		bare?: boolean;
	};

	let {
		value = $bindable(''),
		type = 'text',
		class: className = '',
		bare = false,
		...rest
	}: Props = $props();

	const fieldClasses =
		'min-h-11 w-full rounded-lg border border-[var(--border)] bg-[var(--field-background)] px-3.5 py-2 text-sm text-[var(--text-primary)] shadow-sm outline-none transition-[border-color,box-shadow,background-color] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--focus-ring)] disabled:cursor-not-allowed disabled:opacity-50 motion-reduce:transition-none';
	const bareClasses =
		'min-w-0 flex-1 border-0 bg-transparent p-0 text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)] focus:border-transparent focus:ring-0';
</script>

<input {...rest} {type} bind:value class={(bare ? bareClasses : fieldClasses) + ' ' + className} />
