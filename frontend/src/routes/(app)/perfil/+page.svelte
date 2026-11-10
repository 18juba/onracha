<script lang="ts">
	import { UserRound } from '@lucide/svelte';
	import { getAppContext } from '../../../lib/app-context';

	const { appState } = getAppContext();
</script>

<svelte:head>
	<title>ONRACHA - Perfil</title>
</svelte:head>

<section
	class="relative z-10 flex min-h-0 flex-1 flex-col items-center justify-center gap-4 bg-[var(--app-background)] text-center"
>
	<UserRound size={34} strokeWidth={1.7} class="text-[var(--accent)]" aria-hidden="true" />
	<h1 class="text-2xl font-semibold tracking-tight">{appState.user?.name ?? 'Perfil'}</h1>
	{#if appState.user?.email}
		<p class="text-sm text-[var(--text-secondary)]">{appState.user.email}</p>
	{/if}
</section>
