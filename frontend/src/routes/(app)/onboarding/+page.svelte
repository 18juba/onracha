<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { prefersReducedMotion } from 'svelte/motion';
	import { fly } from 'svelte/transition';
	import {
		ArrowLeft,
		ArrowRight,
		Building2,
		CalendarDays,
		Check,
		MapPin,
		UsersRound
	} from '@lucide/svelte';
	import Button from '../../../lib/components/ui/Button.svelte';

	type OnboardingInterest = 'arena' | 'games';
	type TutorialIcon = 'arena' | 'map' | 'calendar' | 'players';
	type TutorialStep = {
		title: string;
		body: string;
		tip: string;
		icon: TutorialIcon;
	};
	type TutorialGuide = {
		label: string;
		finishLabel: string;
		steps: TutorialStep[];
	};

	const interests: { id: OnboardingInterest; title: string; description: string }[] = [
		{
			id: 'arena',
			title: 'Gerenciar uma arena',
			description: 'Cadastrar meu espaço e liberar horários para jogos.'
		},
		{
			id: 'games',
			title: 'Jogar ou organizar jogos',
			description: 'Encontrar uma arena, marcar um racha ou entrar em um jogo.'
		}
	];

	const tutorials: Record<OnboardingInterest, TutorialGuide> = {
		arena: {
			label: 'Gestão da arena',
			finishLabel: 'Abrir minhas arenas',
			steps: [
				{
					title: 'Cadastre sua arena',
					body: 'Informe o nome, a modalidade e o endereço para apresentar seu espaço aos jogadores.',
					tip: '',
					icon: 'arena'
				},
				{
					title: 'Marque onde ela fica',
					body: 'Toque no mapa para indicar a localização exata da arena e facilitar que os jogadores encontrem o lugar.',
					tip: 'Você pode editar os dados quando precisar.',
					icon: 'map'
				},
				{
					title: 'Libere horários para jogos',
					body: 'Na agenda, escolha datas e horários disponíveis. Os jogadores poderão agendar partidas nesses horários.',
					tip: 'Se um horário deixar de estar disponível, você pode encerrá-lo na agenda.',
					icon: 'calendar'
				}
			]
		},
		games: {
			label: 'Jogadores e organizadores',
			finishLabel: 'Encontrar jogos',
			steps: [
				{
					title: 'Encontre uma arena',
					body: 'Explore o mapa na aba Buscar. Toque em uma arena para conferir o endereço, os horários livres e os jogos abertos.',
					tip: 'Permitir o acesso à localização ajuda a começar pela sua região.',
					icon: 'map'
				},
				{
					title: 'Organize ou entre em um jogo',
					body: 'Use um horário livre para organizar uma partida ou entre em um jogo que já está aberto para participantes.',
					tip: 'Os detalhes da arena mostram as opções disponíveis.',
					icon: 'players'
				},
				{
					title: 'Acompanhe tudo na sala',
					body: 'Confira o local, o horário e os jogadores confirmados. Seus jogos ficam reunidos na aba Jogos.',
					tip: 'Abra a sala da partida para rever as informações quando quiser.',
					icon: 'calendar'
				}
			]
		}
	};

	let selectedInterest = $state<OnboardingInterest | null>(null);
	let stepIndex = $state(0);
	let direction = $state<1 | -1>(1);
	let currentGuide = $derived(selectedInterest ? tutorials[selectedInterest] : null);
	let currentStep = $derived(currentGuide?.steps[stepIndex] ?? null);
	let progress = $derived(currentGuide ? ((stepIndex + 1) / currentGuide.steps.length) * 100 : 0);
	let returnToProfile = $derived(page.state.onboardingReturnToProfile === true);
	let isLastStep = $derived(currentGuide !== null && stepIndex === currentGuide.steps.length - 1);

	function chooseInterest(interest: OnboardingInterest) {
		selectedInterest = interest;
		stepIndex = 0;
		direction = 1;
	}

	function leaveChoice() {
		const destination = returnToProfile ? resolve('/(app)/perfil') : resolve('/(app)/buscar');
		void goto(destination);
	}

	function goBack() {
		if (!selectedInterest) {
			leaveChoice();
			return;
		}

		if (stepIndex === 0) {
			selectedInterest = null;
			stepIndex = 0;
			direction = -1;
			return;
		}

		direction = -1;
		stepIndex -= 1;
	}

	function goToDestination() {
		if (!selectedInterest) return;

		const destination =
			selectedInterest === 'arena' ? resolve('/(app)/arenas') : resolve('/(app)/buscar');
		void goto(destination);
	}

	function nextStep() {
		if (!currentGuide) return;

		if (isLastStep) {
			goToDestination();
			return;
		}

		direction = 1;
		stepIndex += 1;
	}
</script>

<svelte:head>
	<title>ONRACHA - Primeiros passos</title>
	<meta
		name="description"
		content="Veja como encontrar arenas, organizar jogos ou cuidar da sua arena no ON Racha."
	/>
</svelte:head>

{#if selectedInterest === null}
	<section
		class="relative z-10 flex min-h-0 flex-1 flex-col overflow-y-auto bg-[var(--app-background)]"
		aria-labelledby="onboarding-choice-title"
		in:fly={{
			y: prefersReducedMotion.current ? 0 : 12,
			duration: prefersReducedMotion.current ? 0 : 200
		}}
		out:fly={{
			y: prefersReducedMotion.current ? 0 : -10,
			duration: prefersReducedMotion.current ? 0 : 140
		}}
	>
		<div class="mx-auto flex min-h-full w-full max-w-sm flex-col px-5 py-5">
			<header class="flex justify-end">
				<Button variant="ghost" size="sm" onclick={leaveChoice}>
					Agora não
					<ArrowRight size={16} aria-hidden="true" />
				</Button>
			</header>

			<div class="mx-auto my-auto py-6">
				<h1 id="onboarding-choice-title" class="text-xl font-semibold tracking-tight">
					Como você quer usar o ON Racha?
				</h1>
			</div>

			<div class="grid gap-3">
				{#each interests as interest (interest.id)}
					<Button
						variant="outline"
						class="min-h-24 w-full justify-start gap-3 rounded-2xl bg-[var(--surface)] p-4 text-left whitespace-normal hover:border-[var(--accent)]"
						onclick={() => chooseInterest(interest.id)}
					>
						<span
							class="grid size-11 shrink-0 place-items-center rounded-xl bg-[var(--surface-raised)] text-[var(--accent)]"
						>
							{#if interest.id === 'arena'}
								<Building2 size={21} aria-hidden="true" />
							{:else}
								<UsersRound size={21} aria-hidden="true" />
							{/if}
						</span>
						<span class="min-w-0 flex-1">
							<span class="block font-semibold text-[var(--text-primary)]">{interest.title}</span>
							<span class="mt-1 block text-xs leading-5 text-[var(--text-secondary)]">
								{interest.description}
							</span>
						</span>
						<ArrowRight size={17} class="shrink-0 text-[var(--text-muted)]" aria-hidden="true" />
					</Button>
				{/each}
			</div>

			<p class="mt-5 text-center text-xs text-[var(--text-muted)]">
				Você pode mudar de ideia depois, sem problemas
			</p>
		</div>
	</section>
{:else if currentGuide && currentStep}
	<section class="relative z-10 flex min-h-0 flex-1 flex-col bg-[var(--app-background)]">
		<header class="flex shrink-0 items-center gap-2 px-4 py-3">
			<Button variant="ghost" size="icon" aria-label="Voltar uma etapa" onclick={goBack}>
				<ArrowLeft size={19} aria-hidden="true" />
			</Button>
			<div class="min-w-0 flex-1">
				<p class="truncate text-xs text-[var(--text-secondary)]">{currentGuide.label}</p>
				<p class="text-sm font-semibold" aria-live="polite">
					Etapa {stepIndex + 1} de {currentGuide.steps.length}
				</p>
			</div>
			<Button variant="ghost" size="sm" onclick={goToDestination}>Pular tutorial</Button>
		</header>

		<div
			class="mx-5 h-1.5 shrink-0 overflow-hidden rounded-full bg-[var(--surface-raised)]"
			role="progressbar"
			aria-label="Progresso do tutorial"
			aria-valuemin={1}
			aria-valuemax={currentGuide.steps.length}
			aria-valuenow={stepIndex + 1}
		>
			<div
				class="h-full rounded-full bg-[var(--accent)] transition-[width] duration-300 motion-reduce:transition-none"
				style:width={progress + '%'}
			></div>
		</div>

		<div class="min-h-0 flex-1 overflow-y-auto">
			{#key stepIndex}
				<article
					class="mx-auto flex min-h-full w-full max-w-sm flex-col justify-center px-5 py-7"
					aria-labelledby="onboarding-step-title"
					aria-live="polite"
					aria-atomic="true"
					in:fly={{
						x: prefersReducedMotion.current ? 0 : direction * 24,
						duration: prefersReducedMotion.current ? 0 : 220
					}}
					out:fly={{
						x: prefersReducedMotion.current ? 0 : direction * -18,
						duration: prefersReducedMotion.current ? 0 : 140
					}}
				>
					<div
						class="relative mx-auto mb-8 grid size-36 place-items-center rounded-[32px] bg-[var(--surface-raised)] text-[var(--accent)]"
						aria-hidden="true"
					>
						<div
							class="absolute -right-2 -bottom-2 grid size-11 place-items-center rounded-2xl border border-[var(--border)] bg-[var(--surface)] text-[var(--success)] shadow-sm"
						>
							<Check size={19} strokeWidth={2.2} />
						</div>
						{#if currentStep.icon === 'arena'}
							<Building2 size={54} strokeWidth={1.5} />
						{:else if currentStep.icon === 'map'}
							<MapPin size={54} strokeWidth={1.5} />
						{:else if currentStep.icon === 'calendar'}
							<CalendarDays size={54} strokeWidth={1.5} />
						{:else}
							<UsersRound size={54} strokeWidth={1.5} />
						{/if}
					</div>

					<div class="text-center">
						<h1 id="onboarding-step-title" class="text-2xl font-semibold tracking-tight">
							{currentStep.title}
						</h1>
						<p class="mt-3 text-sm leading-6 text-[var(--text-secondary)]">
							{currentStep.body}
						</p>
					</div>

					{#if currentStep.tip}
						<p
							class="mt-5 rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-center text-xs leading-5 text-[var(--text-secondary)]"
						>
							{currentStep.tip}
						</p>
					{/if}

					<div class="mt-5 flex justify-center gap-1.5" aria-hidden="true">
						{#each currentGuide.steps as step (step.title)}
							<span
								class={step.title === currentStep.title
									? 'h-1.5 w-6 rounded-full bg-[var(--accent)] transition-all motion-reduce:transition-none'
									: 'size-1.5 rounded-full bg-[var(--border)]'}
							></span>
						{/each}
					</div>
				</article>
			{/key}
		</div>

		<footer class="shrink-0 border-t border-[var(--border)] px-5 pt-3 pb-4">
			<div class="mx-auto grid w-full max-w-sm grid-cols-[auto_1fr] gap-3">
				<Button variant="outline" onclick={goBack}>Voltar</Button>
				<Button variant="primary" class="min-w-0" onclick={nextStep}>
					<span class="truncate">{isLastStep ? currentGuide.finishLabel : 'Próximo passo'}</span>
					{#if isLastStep}
						<Check size={17} aria-hidden="true" />
					{:else}
						<ArrowRight size={17} aria-hidden="true" />
					{/if}
				</Button>
			</div>
		</footer>
	</section>
{/if}
