<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	type Variant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'unstyled';
	type Size = 'sm' | 'default' | 'lg' | 'icon' | 'none';
	type Props = Omit<HTMLButtonAttributes, 'children' | 'class' | 'size' | 'type'> & {
		children?: Snippet;
		variant?: Variant;
		size?: Size;
		type?: 'button' | 'submit' | 'reset';
		class?: string;
	};

	let {
		children,
		variant = 'primary',
		size = 'default',
		type = 'button',
		class: className = '',
		...rest
	}: Props = $props();

	const variants: Record<Variant, string> = {
		primary:
			'border border-transparent bg-[var(--accent)] text-[var(--accent-ink)] hover:bg-[var(--accent-hover)]',
		secondary:
			'border border-[var(--border)] bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--surface)]',
		outline:
			'border border-[var(--border)] bg-transparent text-[var(--text-primary)] hover:bg-[var(--surface)]',
		ghost:
			'border border-transparent bg-transparent text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]',
		destructive:
			'border border-transparent bg-[var(--danger-surface)] text-[var(--danger)] hover:bg-[var(--danger-surface-hover)]',
		unstyled: 'border-0 text-inherit shadow-none'
	};

	const sizes: Record<Size, string> = {
		sm: 'min-h-9 rounded-lg px-3 text-xs',
		default: 'min-h-11 rounded-lg px-4 text-sm',
		lg: 'min-h-12 rounded-lg px-5 text-sm',
		icon: 'size-10 rounded-lg p-0 text-sm',
		none: 'p-0'
	};
</script>

<button
	{...rest}
	{type}
	data-variant={variant}
	class={'inline-flex shrink-0 items-center justify-center gap-2 font-medium transition-[background-color,color,border-color,box-shadow,transform] duration-150 focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--app-background)] focus-visible:outline-none active:scale-[0.99] disabled:pointer-events-none disabled:opacity-50 motion-reduce:transition-none motion-reduce:active:transform-none ' +
		variants[variant] +
		' ' +
		sizes[size] +
		' ' +
		className}
>
	{@render children?.()}
</button>
