<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Building2, CalendarDays, MapPin, UserRound } from '@lucide/svelte';

	const navigationItems = [
		{ id: 'buscar', path: '/buscar', href: '/(app)/buscar', label: 'Buscar', icon: MapPin },
		{ id: 'jogos', path: '/jogos', href: '/(app)/jogos', label: 'Jogos', icon: CalendarDays },
		{ id: 'arenas', path: '/arenas', href: '/(app)/arenas', label: 'Arenas', icon: Building2 },
		{ id: 'perfil', path: '/perfil', href: '/(app)/perfil', label: 'Perfil', icon: UserRound }
	] as const;

	let activeTab = $derived(
		navigationItems.find((item) => item.path === page.url.pathname)?.id ?? 'buscar'
	);
</script>

<nav
	class="grid grid-cols-4 bg-[var(--bottom-navigation-background)] pb-[calc(env(safe-area-inset-bottom)+0.5rem)] backdrop-blur-xl"
	aria-label="Navegação principal"
>
	{#each navigationItems as item (item.id)}
		{@const Icon = item.icon}
		<a
			class={`flex min-h-[56px] min-w-0 items-center justify-center border-t-2 border-transparent text-[9px] font-medium transition-[color,opacity,border-color] focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === item.id ? 'border-t-[var(--accent)] text-[var(--accent)] opacity-100' : 'text-[var(--text-secondary)] opacity-45 hover:text-[var(--text-primary)] hover:opacity-80'}`}
			href={resolve(item.href)}
			aria-current={activeTab === item.id ? 'page' : undefined}
		>
			<span class="flex flex-col items-center justify-center gap-1 px-1.5 pt-2">
				<Icon size={18} strokeWidth={1.8} aria-hidden="true" />
				<span>{item.label}</span>
			</span>
		</a>
	{/each}
</nav>
