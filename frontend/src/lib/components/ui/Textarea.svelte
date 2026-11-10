<script lang="ts">
	import type { HTMLTextareaAttributes } from 'svelte/elements';

	type Props = Omit<HTMLTextareaAttributes, 'class' | 'value'> & {
		value?: string;
		class?: string;
	};

	let { value = $bindable(''), class: className = '', ...rest }: Props = $props();
</script>

<textarea
	{...rest}
	bind:value
	class={'min-h-24 w-full resize-y rounded-lg border border-[var(--border)] bg-[var(--field-background)] px-3.5 py-3 text-sm text-[var(--text-primary)] shadow-sm transition-[border-color,box-shadow] outline-none placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--focus-ring)] disabled:cursor-not-allowed disabled:opacity-50 motion-reduce:transition-none ' +
		className}></textarea>
