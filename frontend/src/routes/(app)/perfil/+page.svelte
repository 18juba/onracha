<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ArrowRight, BookOpen, LoaderCircle, LogOut, Moon, Sun, UserRound } from '@lucide/svelte';
	import { api } from '../../../lib/api';
	import { getAppContext } from '../../../lib/app-context';
	import Button from '../../../lib/components/ui/Button.svelte';
	import { getUserFacingErrorMessage } from '../../../lib/errors';
	import { toggleTheme } from '../../../lib/theme';

	const { appState } = getAppContext();
	let isLoggingOut = $state(false);
	let errorMessage = $state('');

	function handleThemeToggle() {
		appState.theme = toggleTheme(appState.theme, document.documentElement, window.localStorage);
	}

	function openOnboarding() {
		void goto(resolve('/(app)/onboarding'), { state: { onboardingReturnToProfile: true } });
	}

	async function handleLogout() {
		if (isLoggingOut) return;

		isLoggingOut = true;
		errorMessage = '';
		try {
			await api.logout();
			appState.user = null;
			appState.sessionStatus = 'loading';
			await goto(resolve('/'), { replaceState: true });
		} catch (error) {
			errorMessage = getUserFacingErrorMessage(
				error,
				'Não foi possível encerrar a sessão. Tente novamente.'
			);
		} finally {
			isLoggingOut = false;
		}
	}
</script>

<svelte:head>
	<title>ONRACHA - Perfil</title>
</svelte:head>

<section
	class="relative z-10 flex min-h-0 flex-1 flex-col overflow-y-auto bg-[var(--app-background)]"
>
	<header class="border-b border-[var(--border)] px-5 pt-7 pb-4">
		<h1 class="text-2xl font-semibold tracking-tight">Perfil</h1>
		<p class="mt-1 text-sm text-[var(--text-secondary)]">Conta e preferências</p>
	</header>

	<div class="grid flex-1 content-start gap-6 px-5 py-5">
		<div
			class="flex items-center gap-4 rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-4"
		>
			<div
				class="grid size-14 shrink-0 place-items-center rounded-2xl bg-[var(--surface-raised)] text-[var(--accent)]"
			>
				<UserRound size={25} strokeWidth={1.7} aria-hidden="true" />
			</div>
			<div class="min-w-0">
				<h2 class="truncate font-semibold">{appState.user?.name ?? 'Perfil'}</h2>
				{#if appState.user?.email}
					<p class="mt-1 truncate text-sm text-[var(--text-secondary)]">{appState.user.email}</p>
				{/if}
			</div>
		</div>

		<section class="grid gap-3" aria-labelledby="profile-preferences-title">
			<h2 id="profile-preferences-title" class="text-sm font-semibold text-[var(--text-secondary)]">
				Preferências
			</h2>
			<Button
				variant="outline"
				class="w-full justify-start"
				disabled={isLoggingOut}
				aria-label={appState.theme === 'dark' ? 'Ativar tema claro' : 'Ativar tema escuro'}
				onclick={handleThemeToggle}
			>
				{#if appState.theme === 'dark'}
					<Sun size={18} aria-hidden="true" />
					<span>Ativar tema claro</span>
				{:else}
					<Moon size={18} aria-hidden="true" />
					<span>Ativar tema escuro</span>
				{/if}
			</Button>
		</section>
		<section class="grid gap-3" aria-labelledby="profile-onboarding-title">
			<h2 id="profile-onboarding-title" class="text-sm font-semibold text-[var(--text-secondary)]">
				Primeiros passos
			</h2>
			<Button variant="outline" class="w-full justify-start" onclick={openOnboarding}>
				<BookOpen size={18} aria-hidden="true" />
				<span>Refazer tutorial de início</span>
				<ArrowRight size={16} class="ml-auto text-[var(--text-muted)]" aria-hidden="true" />
			</Button>
		</section>

		<div class="mt-auto grid gap-3 pt-3">
			{#if errorMessage}
				<p
					class="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm"
					role="alert"
				>
					{errorMessage}
				</p>
			{/if}
			<Button
				variant="destructive"
				class="w-full justify-start"
				disabled={isLoggingOut}
				onclick={handleLogout}
			>
				{#if isLoggingOut}
					<LoaderCircle
						size={18}
						class="animate-spin motion-reduce:animate-none"
						aria-hidden="true"
					/>
					<span>Encerrando sessão</span>
				{:else}
					<LogOut size={18} aria-hidden="true" />
					<span>Sair da conta</span>
				{/if}
			</Button>
		</div>
	</div>
</section>
