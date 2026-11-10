<script lang="ts">
	import { Building2, CalendarDays, MapPin, UserRound } from '@lucide/svelte';

	type MainTab = 'buscar' | 'jogos' | 'arenas' | 'perfil';
	type Props = {
		activeTab: MainTab;
		onSelect: (tab: MainTab) => void;
	};

	let { activeTab, onSelect }: Props = $props();
</script>

<nav
	class="grid grid-cols-4 bg-[var(--bottom-navigation-background)] pb-[calc(env(safe-area-inset-bottom)+0.5rem)] backdrop-blur-xl"
	aria-label="Navegação principal"
>
	<button
		class={`flex min-h-[56px] min-w-0 items-center justify-center border-t-2 border-transparent text-[9px] font-medium transition-[color,opacity,border-color] focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'buscar' ? 'border-t-[var(--accent)] text-[var(--accent)] opacity-100' : 'text-[var(--text-secondary)] opacity-45 hover:text-[var(--text-primary)] hover:opacity-80'}`}
		type="button"
		aria-current={activeTab === 'buscar' ? 'page' : undefined}
		onclick={() => onSelect('buscar')}
	>
		<span class="flex flex-col items-center justify-center gap-1 px-1.5 pt-2">
			<MapPin size={18} strokeWidth={1.8} aria-hidden="true" />
			<span>Buscar</span>
		</span>
	</button>
	<button
		class={`flex min-h-[56px] min-w-0 items-center justify-center border-t-2 border-transparent text-[9px] font-medium transition-[color,opacity,border-color] focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'jogos' ? 'border-t-[var(--accent)] text-[var(--accent)] opacity-100' : 'text-[var(--text-secondary)] opacity-45 hover:text-[var(--text-primary)] hover:opacity-80'}`}
		type="button"
		aria-current={activeTab === 'jogos' ? 'page' : undefined}
		onclick={() => onSelect('jogos')}
	>
		<span class="flex flex-col items-center justify-center gap-1 px-1.5 pt-2">
			<CalendarDays size={18} strokeWidth={1.8} aria-hidden="true" />
			<span>Jogos</span>
		</span>
	</button>
	<button
		class={`flex min-h-[56px] min-w-0 items-center justify-center border-t-2 border-transparent text-[9px] font-medium transition-[color,opacity,border-color] focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'arenas' ? 'border-t-[var(--accent)] text-[var(--accent)] opacity-100' : 'text-[var(--text-secondary)] opacity-45 hover:text-[var(--text-primary)] hover:opacity-80'}`}
		type="button"
		aria-current={activeTab === 'arenas' ? 'page' : undefined}
		onclick={() => onSelect('arenas')}
	>
		<span class="flex flex-col items-center justify-center gap-1 px-1.5 pt-2">
			<Building2 size={18} strokeWidth={1.8} aria-hidden="true" />
			<span>Arenas</span>
		</span>
	</button>
	<button
		class={`flex min-h-[56px] min-w-0 items-center justify-center border-t-2 border-transparent text-[9px] font-medium transition-[color,opacity,border-color] focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'perfil' ? 'border-t-[var(--accent)] text-[var(--accent)] opacity-100' : 'text-[var(--text-secondary)] opacity-45 hover:text-[var(--text-primary)] hover:opacity-80'}`}
		type="button"
		aria-current={activeTab === 'perfil' ? 'page' : undefined}
		onclick={() => onSelect('perfil')}
	>
		<span class="flex flex-col items-center justify-center gap-1 px-1.5 pt-2">
			<UserRound size={18} strokeWidth={1.8} aria-hidden="true" />
			<span>Perfil</span>
		</span>
	</button>
</nav>
