export type OnboardingScreen = 'welcome' | 'email' | 'password' | 'register' | 'success';
export type ChoiceScreen = Extract<OnboardingScreen, 'email' | 'register'>;

export function getChoiceScreen(hasAccount: boolean): ChoiceScreen {
	return hasAccount ? 'email' : 'register';
}

export function passwordsMatch(password: string, confirmation: string): boolean {
	return password === confirmation;
}
