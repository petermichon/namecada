import { lookupDomain, parseDomain, UnsupportedTldError } from './rdap';

export type Status = 'available' | 'taken' | 'unknown';
export type ProviderKind = 'domain' | 'dev' | 'social' | 'org' | 'brand';

export interface Provider {
	id: string;
	label: string;
	kind: ProviderKind;
	name: (label: string) => string;
	url: (label: string) => string;
	check: (label: string, signal?: AbortSignal) => Promise<{ status: Status; detail?: string }>;
}

export interface ProviderResult {
	id: string;
	label: string;
	kind: ProviderKind;
	name: string;
	url: string;
	status: Status;
	detail?: string;
}

export const TLDS = ['.com', '.org', '.net', '.dev'] as const;

const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function toLabel(input: string): string | null {
	const value = input
		.trim()
		.toLowerCase()
		.replace(/^\.+|\.+$/g, '');
	if (!value) return null;
	if (value.includes('.')) {
		const parsed = parseDomain(value);
		if (!parsed) return null;
		return parsed.domain.slice(0, parsed.domain.indexOf('.'));
	}
	return LABEL.test(value) ? value : null;
}

function statusFrom(code: number): Status {
	if (code === 200) return 'taken';
	if (code === 404) return 'available';
	return 'unknown';
}

const rdap = (tld: string): Provider => ({
	id: `domain${tld}`,
	label: `Domain ${tld}`,
	kind: 'domain',
	name: (label) => `${label}${tld}`,
	url: (label) => `https://${label}${tld}`,
	async check(label, signal) {
		try {
			const res = await lookupDomain(`${label}${tld}`, signal);
			return {
				status:
					res.status === 'registered'
						? 'taken'
						: res.status === 'unregistered'
							? 'available'
							: 'unknown',
				detail: res.registeredOn ? `registered ${res.registeredOn.slice(0, 10)}` : undefined
			};
		} catch (e) {
			if (e instanceof DOMException && e.name === 'AbortError') throw e;
			if (e instanceof UnsupportedTldError)
				return { status: 'unknown', detail: 'no RDAP endpoint' };
			throw e;
		}
	}
});

async function checkHttp(url: string, signal?: AbortSignal): Promise<Status> {
	const res = await fetch(url, {
		signal,
		referrerPolicy: 'no-referrer',
		credentials: 'omit',
		headers: { Accept: 'application/json' }
	});
	res.body?.cancel().catch(() => {});
	return statusFrom(res.status);
}

export const github: Provider = {
	id: 'github',
	label: 'GitHub',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://github.com/${label}`,
	async check(label, signal) {
		const user = await checkHttp(`https://api.github.com/users/${label}`, signal);
		if (user !== 'available') return { status: user };
		const org = await checkHttp(`https://api.github.com/orgs/${label}`, signal);
		return { status: org };
	}
};

export const npm: Provider = {
	id: 'npm',
	label: 'npm',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://www.npmjs.com/package/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://registry.npmjs.org/${label}`, signal) };
	}
};

export const providers: Provider[] = [...TLDS.map(rdap), github, npm];

async function runProvider(
	provider: Provider,
	label: string,
	signal?: AbortSignal
): Promise<ProviderResult> {
	const base = {
		id: provider.id,
		label: provider.label,
		kind: provider.kind,
		name: provider.name(label),
		url: provider.url(label)
	};
	try {
		const { status, detail } = await provider.check(label, signal);
		return { ...base, status, detail };
	} catch (e) {
		if (e instanceof DOMException && e.name === 'AbortError') throw e;
		return { ...base, status: 'unknown' as Status, detail: 'unreachable' };
	}
}

export async function streamAll(
	label: string,
	onResult: (result: ProviderResult) => void,
	signal?: AbortSignal
): Promise<void> {
	await Promise.all(
		providers.map(async (provider) => {
			onResult(await runProvider(provider, label, signal));
		})
	);
}

export async function checkAll(label: string, signal?: AbortSignal): Promise<ProviderResult[]> {
	const byId = new Map<string, ProviderResult>();
	await streamAll(label, (result) => byId.set(result.id, result), signal);
	return providers
		.map((provider) => byId.get(provider.id))
		.filter((result): result is ProviderResult => result !== undefined);
}
