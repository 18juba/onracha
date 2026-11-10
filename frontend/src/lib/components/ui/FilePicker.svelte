<script lang="ts">
	import { ImagePlus } from '@lucide/svelte';

	type Props = {
		inputKey: number;
		label: string;
		accept: string;
		name?: string;
		onchange: (event: Event) => void;
	};

	let { inputKey, label, accept, name, onchange }: Props = $props();
</script>

<label
	class="inline-flex min-h-11 cursor-pointer items-center justify-center gap-2 rounded-lg border border-[var(--border)] bg-[var(--surface-raised)] px-4 text-sm font-medium text-[var(--text-primary)] transition-colors hover:bg-[var(--surface)] focus-within:ring-2 focus-within:ring-[var(--focus-ring)] motion-reduce:transition-none"
>
	{#key inputKey}
		<input class="sr-only" type="file" {accept} {name} {onchange} />
	{/key}
	<ImagePlus size={16} class="text-[var(--text-muted)]" aria-hidden="true" />
	<span>{label}</span>
</label>
