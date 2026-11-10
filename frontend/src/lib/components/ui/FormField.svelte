<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		label: string;
		for: string;
		required?: boolean;
		hint?: string;
		error?: string;
		class?: string;
		children?: Snippet;
	};

	let {
		label,
		for: htmlFor,
		required = false,
		hint,
		error,
		class: className = '',
		children
	}: Props = $props();
</script>

<div class={'grid gap-2 ' + className}>
	<label for={htmlFor} class="text-sm font-medium text-[var(--text-primary)]">
		{label}
		{#if required}<span class="ml-1 text-[var(--text-muted)]" aria-hidden="true">*</span>{/if}
	</label>
	{@render children?.()}
	{#if hint}<p class="text-xs text-[var(--text-muted)]">{hint}</p>{/if}
	{#if error}<p class="text-sm text-[var(--danger)]" role="alert">{error}</p>{/if}
</div>
