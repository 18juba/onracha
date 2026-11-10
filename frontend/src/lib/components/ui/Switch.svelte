<script lang="ts">
	type Props = {
		checked: boolean;
		label: string;
		disabled?: boolean;
		class?: string;
		onCheckedChange?: (checked: boolean) => Promise<boolean | void> | boolean | void;
	};

	let {
		checked = $bindable(false),
		label,
		disabled = false,
		class: className = '',
		onCheckedChange
	}: Props = $props();

	async function handleChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const previous = checked;
		checked = input.checked;
		try {
			const accepted = await onCheckedChange?.(input.checked);
			if (accepted === false) checked = previous;
		} catch {
			checked = previous;
		}
	}
</script>

<label
	class={'inline-flex min-h-11 cursor-pointer items-center gap-2.5 text-sm text-[var(--text-primary)] ' +
		(disabled ? 'cursor-not-allowed opacity-50 ' : '') +
		className}
>
	<input
		type="checkbox"
		{checked}
		{disabled}
		aria-label={label}
		onchange={handleChange}
		class="peer sr-only"
	/>
	<span
		class="relative h-[22px] w-10 shrink-0 rounded-full border border-[var(--border)] bg-[var(--surface-raised)] transition-[background-color,border-color] peer-checked:border-[var(--accent)] peer-checked:bg-[var(--accent)] peer-focus-visible:ring-2 peer-focus-visible:ring-[var(--focus-ring)] peer-focus-visible:ring-offset-2 peer-focus-visible:ring-offset-[var(--app-background)] motion-reduce:transition-none"
		aria-hidden="true"
	>
		<span
			class={'absolute top-[3px] size-3.5 rounded-full bg-[var(--text-secondary)] shadow-sm transition-[transform,background-color] peer-checked:bg-[var(--accent-ink)] ' +
				(checked ? 'translate-x-[19px] bg-[var(--accent-ink)] ' : 'translate-x-[3px]')}
		></span>
	</span>
	<span>{label}</span>
</label>
