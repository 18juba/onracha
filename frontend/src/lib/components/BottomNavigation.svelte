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
	class="grid grid-cols-4 gap-1 border-t border-[var(--border)] bg-[var(--surface)] px-3 pt-2 pb-[calc(env(safe-area-inset-bottom)+0.5rem)] backdrop-blur-xl"
	aria-label="Navegação principal"
>
	<button
		class={`flex min-h-[58px] min-w-0 flex-col items-center justify-center gap-1 rounded-2xl px-1 py-2 text-[11px] font-medium transition-colors focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'buscar' ? 'bg-[var(--focus-ring)] text-[var(--accent)]' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]'}`}
		type="button"
		aria-current={activeTab === 'buscar' ? 'page' : undefined}
		onclick={() => onSelect('buscar')}
	>
		<MapPin size={20} strokeWidth={2} aria-hidden="true" />
		<span>Buscar</span>
	</button>
	<button
		class={`flex min-h-[58px] min-w-0 flex-col items-center justify-center gap-1 rounded-2xl px-1 py-2 text-[11px] font-medium transition-colors focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'jogos' ? 'bg-[var(--focus-ring)] text-[var(--accent)]' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]'}`}
		type="button"
		aria-current={activeTab === 'jogos' ? 'page' : undefined}
		onclick={() => onSelect('jogos')}
	>
		<CalendarDays size={20} strokeWidth={2} aria-hidden="true" />
		<span>Jogos</span>
	</button>
	<button
		class={`flex min-h-[58px] min-w-0 flex-col items-center justify-center gap-1 rounded-2xl px-1 py-2 text-[11px] font-medium transition-colors focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'arenas' ? 'bg-[var(--focus-ring)] text-[var(--accent)]' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]'}`}
		type="button"
		aria-current={activeTab === 'arenas' ? 'page' : undefined}
		onclick={() => onSelect('arenas')}
	>
		<Building2 size={20} strokeWidth={2} aria-hidden="true" />
		<span>Arenas</span>
	</button>
	<button
		class={`flex min-h-[58px] min-w-0 flex-col items-center justify-center gap-1 rounded-2xl px-1 py-2 text-[11px] font-medium transition-colors focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none ${activeTab === 'perfil' ? 'bg-[var(--focus-ring)] text-[var(--accent)]' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-raised)] hover:text-[var(--text-primary)]'}`}
		type="button"
		aria-current={activeTab === 'perfil' ? 'page' : undefined}
		onclick={() => onSelect('perfil')}
	>
		<UserRound size={20} strokeWidth={2} aria-hidden="true" />
		<span>Perfil</span>
	</button>
</nav>
