<script lang="ts">
	import ArenaManagement from '../../../lib/components/ArenaManagement.svelte';
	import { getAppContext } from '../../../lib/app-context';

	const { appState } = getAppContext();
</script>

<svelte:head>
	<title>ONRACHA — Arenas</title>
</svelte:head>

<ArenaManagement theme={appState.theme} />
