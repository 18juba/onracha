<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import MyGames from '../../../lib/components/MyGames.svelte';
	import { getAppContext } from '../../../lib/app-context';

	const { appState } = getAppContext();

	function findArenas() {
		void goto(resolve('/(app)/buscar'));
	}
</script>

<svelte:head>
	<title>ONRACHA - Jogos</title>
</svelte:head>

<MyGames
	arenas={appState.discoveredArenas}
	userID={appState.user?.id ?? null}
	onFindArenas={findArenas}
/>
