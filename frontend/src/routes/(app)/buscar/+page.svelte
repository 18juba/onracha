<script lang="ts">
	import ArenaDetails from '../../../lib/components/ArenaDetails.svelte';
	import MapView from '../../../lib/components/MapView.svelte';
	import { getAppContext } from '../../../lib/app-context';
	import type { DiscoveredArena } from '../../../lib/api';

	const { appState, refreshGameAvailability, rememberDiscoveredArenas } = getAppContext();
	let selectedArena = $state<DiscoveredArena | null>(null);

	function selectArena(arena: DiscoveredArena) {
		selectedArena = arena;
	}
</script>

<svelte:head>
	<title>ONRACHA — Buscar</title>
</svelte:head>

<div class="absolute inset-0 flex">
	<MapView
		theme={appState.theme}
		discoverArenas
		gameAvailabilityVersion={appState.gameAvailabilityVersion}
		onArenaSelect={selectArena}
		onArenasChange={rememberDiscoveredArenas}
	/>
</div>

{#if selectedArena}
	{#key selectedArena.id}
		<ArenaDetails
			arena={selectedArena}
			onClose={() => (selectedArena = null)}
			onGamesChanged={refreshGameAvailability}
		/>
	{/key}
{/if}
