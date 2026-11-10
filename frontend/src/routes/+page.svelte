<script lang="ts">
	import { onMount } from 'svelte';
	import { prefersReducedMotion } from 'svelte/motion';
	import { fly } from 'svelte/transition';
	import {
		ArrowLeft,
		ArrowRight,
		Eye,
		EyeOff,
		LoaderCircle,
		LockKeyhole,
		Mail,
		Moon,
		Sun,
		UserRound
	} from '@lucide/svelte';
	import MainApp from '../lib/components/MainApp.svelte';
	import { api, type User } from '../lib/api';
	import { getUserFacingErrorMessage } from '../lib/errors';
	import { getChoiceScreen, passwordsMatch, type OnboardingScreen } from '../lib/onboarding';
	import { applyTheme, readThemePreference, toggleTheme, type Theme } from '../lib/theme';

	let screen = $state<OnboardingScreen>('welcome');
	let theme = $state<Theme>('dark');
	let direction = $state<1 | -1>(1);
	let email = $state('');
	let name = $state('');
	let password = $state('');
	let passwordConfirmation = $state('');
	let showPassword = $state(false);
	let showPasswordConfirmation = $state(false);
	let submitting = $state(false);
	let errorMessage = $state('');
	let noticeMessage = $state('');
	let currentUser = $state<User | null>(null);

	onMount(() => {
		theme = readThemePreference(window.localStorage);
		applyTheme(theme, document.documentElement);
	});

	function handleThemeToggle() {
		theme = toggleTheme(theme, document.documentElement, window.localStorage);
	}

	function moveTo(nextScreen: OnboardingScreen, backwards = false) {
		direction = backwards ? -1 : 1;
		screen = nextScreen;
		errorMessage = '';
	}

	function startFlow(hasAccount: boolean) {
		email = '';
		name = '';
		password = '';
		passwordConfirmation = '';
		currentUser = null;
		noticeMessage = '';
		showPassword = false;
		showPasswordConfirmation = false;
		moveTo(getChoiceScreen(hasAccount));
	}

	function goBack() {
		if (screen === 'password') {
			moveTo('email', true);
		} else if (screen === 'email' || screen === 'register') {
			noticeMessage = '';
			moveTo('welcome', true);
		}
	}

	function submitEmail(event: SubmitEvent) {
		event.preventDefault();
		email = email.trim();
		noticeMessage = '';
		moveTo('password');
	}

	async function submitLogin(event: SubmitEvent) {
		event.preventDefault();
		errorMessage = '';
		submitting = true;

		try {
			currentUser = await api.login({ email: email.trim(), password });
			password = '';
			moveTo('success');
		} catch (error) {
			errorMessage = getUserFacingErrorMessage(
				error,
				'Não foi possível conectar à API. Tente novamente.'
			);
		} finally {
			submitting = false;
		}
	}

	async function submitRegistration(event: SubmitEvent) {
		event.preventDefault();
		errorMessage = '';
		noticeMessage = '';

		if (!passwordsMatch(password, passwordConfirmation)) {
			errorMessage = 'As senhas não coincidem.';
			return;
		}

		const registration = { name: name.trim(), email: email.trim(), password };
		submitting = true;
		let accountWasCreated = false;

		try {
			await api.register(registration);
			accountWasCreated = true;
			currentUser = await api.login({ email: registration.email, password });
			password = '';
			passwordConfirmation = '';
			moveTo('success');
		} catch (error) {
			if (accountWasCreated) {
				name = '';
				password = '';
				passwordConfirmation = '';
				email = registration.email;
				noticeMessage = 'Conta criada. Entre com seu email e senha.';
				moveTo('email');
			} else {
				errorMessage = getUserFacingErrorMessage(
					error,
					'Não foi possível conectar à API. Tente novamente.'
				);
			}
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{screen === 'success' ? 'ONRACHA — Buscar' : 'ONRACHA — entrar ou criar conta'}</title>
	<meta name="description" content="Entre no ONRACHA para jogar junto." />
</svelte:head>

{#if screen === 'success'}
	<MainApp user={currentUser} />
{:else}
<main
	class="grid min-h-svh w-full place-items-center bg-[var(--page-outside)] px-0 py-0 sm:px-6 sm:py-6"
>
	<div
		class="relative flex min-h-svh w-full max-w-[430px] flex-col overflow-hidden bg-[var(--app-background)] px-6 pt-[calc(18px+env(safe-area-inset-top))] pb-[calc(20px+env(safe-area-inset-bottom))] text-[var(--text-primary)] shadow-[0_32px_100px_rgba(0,0,0,0.24)] transition-colors duration-300 motion-reduce:transition-none sm:min-h-[min(860px,calc(100svh-48px))] sm:rounded-[32px] sm:border sm:border-[var(--border)]"
	>
		<header class="relative z-10 grid h-11 grid-cols-[44px_1fr_44px] items-center">
			<div class="grid size-11 place-items-center">
				{#if screen !== 'welcome'}
					<button
						class="grid size-11 place-items-center rounded-full text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface)] hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
						aria-label="Voltar"
						type="button"
						onclick={goBack}
					>
						<ArrowLeft size={20} aria-hidden="true" />
					</button>
				{/if}
			</div>

			<img
				class={`mx-auto h-auto w-[174px] max-w-full object-contain transition-[filter] duration-300 ${theme === 'light' ? 'hue-rotate-180 invert' : ''}`}
				src="/images/logo.png"
				alt="ONRACHA"
			/>

			<button
				class="grid size-11 place-items-center rounded-full border border-[var(--border)] bg-[var(--surface)] text-[var(--text-secondary)] transition-colors hover:text-[var(--accent)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				aria-label={theme === 'dark' ? 'Ativar tema claro' : 'Ativar tema escuro'}
				aria-pressed={theme === 'light'}
				type="button"
				onclick={handleThemeToggle}
			>
				{#if theme === 'dark'}
					<Sun size={19} aria-hidden="true" />
				{:else}
					<Moon size={19} aria-hidden="true" />
				{/if}
			</button>
		</header>

		{#if screen === 'email' || screen === 'password'}
			<div
				class="mt-6 flex gap-2"
				role="progressbar"
				aria-label="Progresso do login"
				aria-valuemin={1}
				aria-valuemax={2}
				aria-valuenow={screen === 'email' ? 1 : 2}
			>
				<span class="h-1 flex-1 rounded-full bg-[var(--accent)]"></span>
				<span
					class={`h-1 flex-1 rounded-full transition-colors duration-300 ${screen === 'password' ? 'bg-[var(--accent)]' : 'bg-[var(--border)]'}`}
				></span>
			</div>
		{/if}

		<div class={`relative z-10 flex flex-1 flex-col ${screen === 'welcome' ? '' : 'pt-8'}`}>
			{#key screen}
				<section
					class={`flex flex-1 flex-col ${screen === 'welcome' ? 'justify-between pb-2' : ''}`}
					aria-labelledby="screen-title"
					aria-live="polite"
					in:fly={{
						x: prefersReducedMotion.current ? 0 : direction * 28,
						duration: prefersReducedMotion.current ? 0 : 230
					}}
					out:fly={{
						x: prefersReducedMotion.current ? 0 : direction * -20,
						duration: prefersReducedMotion.current ? 0 : 160
					}}
				>
					{#if screen === 'welcome'}
						<div class="flex flex-1 flex-col items-center justify-center">
							<div class="h-[230px] w-full max-w-[320px]" aria-hidden="true">
								<svg viewBox="0 0 320 250" class="h-full w-full">
									<ellipse
										cx="160"
										cy="125"
										rx="129"
										ry="47"
										transform="rotate(-22 160 125)"
										class="origin-center animate-[spin_24s_linear_infinite] fill-none stroke-[var(--accent)] opacity-30 motion-reduce:animate-none"
									/>
									<ellipse
										cx="160"
										cy="125"
										rx="105"
										ry="69"
										transform="rotate(36 160 125)"
										class="fill-none stroke-[var(--text-secondary)] opacity-15"
									/>
									<circle cx="160" cy="125" r="64" class="fill-[var(--accent)] opacity-[0.07]" />
									<g
										class="origin-center animate-[spin_28s_linear_infinite] motion-reduce:animate-none"
									>
										<circle
											cx="160"
											cy="121"
											r="48"
											class="fill-[var(--surface-raised)] stroke-[var(--accent)] stroke-[1.5]"
										/>
										<path
											d="M160 99 179 113 172 135 148 135 141 113Z"
											class="fill-[var(--accent)] opacity-70"
										/>
										<path
											d="m160 99 1-25m18 39 24-8m-31 30 15 19m-39-19-15 19m-7-41-24-8"
											class="fill-none stroke-[var(--text-secondary)] stroke-[1.5] opacity-45"
										/>
									</g>
									<circle
										cx="57"
										cy="74"
										r="5"
										class="animate-pulse fill-[var(--accent)] motion-reduce:animate-none"
									/>
									<circle cx="270" cy="171" r="3.5" class="fill-[var(--accent)] opacity-70" />
								</svg>
							</div>

							<h1
								id="screen-title"
								class="mt-3 max-w-[340px] text-center text-[32px] leading-[1.06] font-bold tracking-[-0.055em] sm:text-[36px]"
							>
								Você já tem conta?
							</h1>
						</div>

						<div class="grid w-full gap-3">
							<button
								class="flex h-[60px] w-full items-center justify-between rounded-2xl bg-[var(--accent)] px-5 text-left text-[15px] font-semibold text-[var(--accent-ink)] transition-[background-color,transform] duration-150 hover:bg-[var(--accent-hover)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.99]"
								type="button"
								onclick={() => startFlow(true)}
							>
								<span>Já tenho conta</span>
								<ArrowRight size={20} aria-hidden="true" />
							</button>
							<button
								class="flex h-[60px] w-full items-center justify-between rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-5 text-left text-[15px] font-semibold transition-[background-color,transform] duration-150 hover:bg-[var(--surface-raised)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.99]"
								type="button"
								onclick={() => startFlow(false)}
							>
								<span>Criar conta</span>
								<UserRound size={19} class="text-[var(--text-secondary)]" aria-hidden="true" />
							</button>
						</div>
					{:else if screen === 'email'}
						<div class="mb-8">
							<div
								class="mb-6 grid size-[52px] place-items-center rounded-[18px] border border-[var(--border)] bg-[var(--surface)] text-[var(--accent)]"
								aria-hidden="true"
							>
								<Mail size={22} />
							</div>
							<h1
								id="screen-title"
								class="max-w-[340px] text-[32px] leading-[1.06] font-bold tracking-[-0.055em] sm:text-[36px]"
							>
								Qual é o seu email?
							</h1>
						</div>

						<form class="grid gap-4" onsubmit={submitEmail}>
							<div
								class="flex h-[58px] items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--field-background)] px-4 transition-[border-color,box-shadow] focus-within:border-[var(--accent)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]"
							>
								<label class="sr-only" for="login-email">Email</label>
								<Mail size={19} class="shrink-0 text-[var(--text-muted)]" aria-hidden="true" />
								<input
									id="login-email"
									class="min-w-0 flex-1 bg-transparent text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]"
									type="email"
									name="email"
									bind:value={email}
									placeholder="Seu email"
									autocomplete="email"
									inputmode="email"
									autocapitalize="none"
									spellcheck="false"
									required
								/>
							</div>

							{#if noticeMessage}
								<p
									class="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm text-[var(--accent)]"
									role="status"
								>
									{noticeMessage}
								</p>
							{/if}

							{#if errorMessage}
								<p
									class="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm text-[var(--danger)]"
									role="alert"
								>
									{errorMessage}
								</p>
							{/if}

							<button
								class="mt-2 flex h-[58px] w-full items-center justify-between rounded-2xl bg-[var(--accent)] px-5 text-[15px] font-semibold text-[var(--accent-ink)] transition-[background-color,transform] duration-150 hover:bg-[var(--accent-hover)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.99]"
								type="submit"
							>
								<span>Continuar</span>
								<ArrowRight size={20} aria-hidden="true" />
							</button>
						</form>
					{:else if screen === 'password'}
						<div class="mb-8">
							<div
								class="mb-6 grid size-[52px] place-items-center rounded-[18px] border border-[var(--border)] bg-[var(--surface)] text-[var(--accent)]"
								aria-hidden="true"
							>
								<LockKeyhole size={22} />
							</div>
							<h1
								id="screen-title"
								class="max-w-[340px] text-[32px] leading-[1.06] font-bold tracking-[-0.055em] sm:text-[36px]"
							>
								Digite sua senha
							</h1>
							<div
								class="mt-5 flex min-h-11 items-center gap-2 rounded-xl border border-[var(--border)] bg-[var(--surface)] px-3 text-[13px] text-[var(--text-secondary)]"
							>
								<Mail size={16} class="shrink-0" aria-hidden="true" />
								<span class="truncate">{email}</span>
							</div>
						</div>

						<form class="grid gap-4" onsubmit={submitLogin}>
							<div
								class="flex h-[58px] items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--field-background)] px-4 transition-[border-color,box-shadow] focus-within:border-[var(--accent)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]"
							>
								<label class="sr-only" for="login-password">Senha</label>
								<LockKeyhole
									size={19}
									class="shrink-0 text-[var(--text-muted)]"
									aria-hidden="true"
								/>
								<input
									id="login-password"
									class="min-w-0 flex-1 bg-transparent text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]"
									type={showPassword ? 'text' : 'password'}
									name="password"
									bind:value={password}
									placeholder="Sua senha"
									autocomplete="current-password"
									minlength={8}
									maxlength={128}
									required
								/>
								<button
									class="grid size-9 shrink-0 place-items-center rounded-full text-[var(--text-muted)] transition-colors hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
									aria-label={showPassword ? 'Ocultar senha' : 'Mostrar senha'}
									aria-pressed={showPassword}
									type="button"
									onclick={() => (showPassword = !showPassword)}
								>
									{#if showPassword}
										<EyeOff size={18} aria-hidden="true" />
									{:else}
										<Eye size={18} aria-hidden="true" />
									{/if}
								</button>
							</div>

							{#if errorMessage}
								<p
									class="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm text-[var(--danger)]"
									role="alert"
								>
									{errorMessage}
								</p>
							{/if}

							<button
								class="mt-2 flex h-[58px] w-full items-center justify-between rounded-2xl bg-[var(--accent)] px-5 text-[15px] font-semibold text-[var(--accent-ink)] transition-[background-color,transform] duration-150 hover:bg-[var(--accent-hover)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.99] disabled:cursor-wait disabled:opacity-65"
								type="submit"
								disabled={submitting}
							>
								<span>{submitting ? 'Entrando...' : 'Entrar'}</span>
								{#if submitting}
									<LoaderCircle size={19} class="animate-spin" aria-hidden="true" />
								{:else}
									<ArrowRight size={20} aria-hidden="true" />
								{/if}
							</button>
						</form>
					{:else if screen === 'register'}
						<div class="mb-7">
							<div
								class="mb-5 grid size-[52px] place-items-center rounded-[18px] border border-[var(--border)] bg-[var(--surface)] text-[var(--accent)]"
								aria-hidden="true"
							>
								<UserRound size={22} />
							</div>
							<h1
								id="screen-title"
								class="max-w-[340px] text-[32px] leading-[1.06] font-bold tracking-[-0.055em] sm:text-[36px]"
							>
								Criar conta
							</h1>
						</div>

						<form class="grid gap-3" onsubmit={submitRegistration}>
							<div
								class="flex h-[56px] items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--field-background)] px-4 transition-[border-color,box-shadow] focus-within:border-[var(--accent)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]"
							>
								<label class="sr-only" for="register-name">Nome</label>
								<UserRound size={18} class="shrink-0 text-[var(--text-muted)]" aria-hidden="true" />
								<input
									id="register-name"
									class="min-w-0 flex-1 bg-transparent text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]"
									name="name"
									bind:value={name}
									placeholder="Seu nome"
									autocomplete="name"
									maxlength={120}
									required
								/>
							</div>

							<div
								class="flex h-[56px] items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--field-background)] px-4 transition-[border-color,box-shadow] focus-within:border-[var(--accent)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]"
							>
								<label class="sr-only" for="register-email">Email</label>
								<Mail size={18} class="shrink-0 text-[var(--text-muted)]" aria-hidden="true" />
								<input
									id="register-email"
									class="min-w-0 flex-1 bg-transparent text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]"
									type="email"
									name="email"
									bind:value={email}
									placeholder="Seu email"
									autocomplete="email"
									inputmode="email"
									autocapitalize="none"
									spellcheck="false"
									maxlength={254}
									required
								/>
							</div>

							<div
								class="flex h-[56px] items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--field-background)] px-4 transition-[border-color,box-shadow] focus-within:border-[var(--accent)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]"
							>
								<label class="sr-only" for="register-password">Senha</label>
								<LockKeyhole
									size={18}
									class="shrink-0 text-[var(--text-muted)]"
									aria-hidden="true"
								/>
								<input
									id="register-password"
									class="min-w-0 flex-1 bg-transparent text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]"
									type={showPassword ? 'text' : 'password'}
									name="password"
									bind:value={password}
									placeholder="Senha"
									autocomplete="new-password"
									minlength={8}
									maxlength={128}
									required
								/>
								<button
									class="grid size-9 shrink-0 place-items-center rounded-full text-[var(--text-muted)] transition-colors hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
									aria-label={showPassword ? 'Ocultar senha' : 'Mostrar senha'}
									aria-pressed={showPassword}
									type="button"
									onclick={() => (showPassword = !showPassword)}
								>
									{#if showPassword}
										<EyeOff size={18} aria-hidden="true" />
									{:else}
										<Eye size={18} aria-hidden="true" />
									{/if}
								</button>
							</div>

							<div
								class="flex h-[56px] items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--field-background)] px-4 transition-[border-color,box-shadow] focus-within:border-[var(--accent)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]"
							>
								<label class="sr-only" for="register-password-confirmation">Confirme a senha</label>
								<LockKeyhole
									size={18}
									class="shrink-0 text-[var(--text-muted)]"
									aria-hidden="true"
								/>
								<input
									id="register-password-confirmation"
									class="min-w-0 flex-1 bg-transparent text-[15px] text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]"
									type={showPasswordConfirmation ? 'text' : 'password'}
									name="password_confirmation"
									bind:value={passwordConfirmation}
									placeholder="Confirme sua senha"
									autocomplete="new-password"
									minlength={8}
									maxlength={128}
									required
								/>
								<button
									class="grid size-9 shrink-0 place-items-center rounded-full text-[var(--text-muted)] transition-colors hover:text-[var(--text-primary)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
									aria-label={showPasswordConfirmation
										? 'Ocultar confirmação da senha'
										: 'Mostrar confirmação da senha'}
									aria-pressed={showPasswordConfirmation}
									type="button"
									onclick={() => (showPasswordConfirmation = !showPasswordConfirmation)}
								>
									{#if showPasswordConfirmation}
										<EyeOff size={18} aria-hidden="true" />
									{:else}
										<Eye size={18} aria-hidden="true" />
									{/if}
								</button>
							</div>

							{#if errorMessage}
								<p
									class="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm text-[var(--danger)]"
									role="alert"
								>
									{errorMessage}
								</p>
							{/if}

							<button
								class="mt-2 flex h-[58px] w-full items-center justify-between rounded-2xl bg-[var(--accent)] px-5 text-[15px] font-semibold text-[var(--accent-ink)] transition-[background-color,transform] duration-150 hover:bg-[var(--accent-hover)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none active:scale-[0.99] disabled:cursor-wait disabled:opacity-65"
								type="submit"
								disabled={submitting}
							>
								<span>{submitting ? 'Criando...' : 'Criar conta'}</span>
								{#if submitting}
									<LoaderCircle size={19} class="animate-spin" aria-hidden="true" />
								{:else}
									<ArrowRight size={20} aria-hidden="true" />
								{/if}
							</button>
						</form>
					{/if}
				</section>
			{/key}
		</div>
	</div>
</main>
{/if}
