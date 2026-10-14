export type Status = 'available' | 'taken' | 'unknown';

export interface ProviderResult {
	id: string;
	platform: string;
	platformName: string;
	name: string;
	url: string;
	status: Status;
	detail?: string;
	paid: boolean;
}

const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function toLabel(input: string): string | null {
	const value = input.trim().toLowerCase();
	return LABEL.test(value) ? value : null;
}
