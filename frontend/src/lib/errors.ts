import { ApiError } from './api';

export function getUserFacingErrorMessage(error: unknown, fallbackMessage: string): string {
	return error instanceof ApiError ? error.message : fallbackMessage;
}
