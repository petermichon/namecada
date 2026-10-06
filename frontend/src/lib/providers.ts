import { lookupDomain, parseDomain, UnsupportedTldError } from './rdap';
import { getCached, isFresh, setCached } from './providerCache';

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

async function checkHttp(
	url: string,
	signal?: AbortSignal,
	availableCodes: number[] = [404]
): Promise<Status> {
	const res = await fetch(url, {
		signal,
		referrerPolicy: 'no-referrer',
		credentials: 'omit',
		headers: { Accept: 'application/json' }
	});
	res.body?.cancel().catch(() => {});
	if (res.status === 200) return 'taken';
	if (availableCodes.includes(res.status)) return 'available';
	return 'unknown';
}

export const github: Provider = {
	id: 'github',
	label: 'GitHub',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://github.com/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://api.github.com/users/${label}`, signal) };
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

export const pypi: Provider = {
	id: 'pypi',
	label: 'PyPI',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://pypi.org/project/${label}/`,
	async check(label, signal) {
		return { status: await checkHttp(`https://pypi.org/pypi/${label}/json`, signal) };
	}
};

export const crates: Provider = {
	id: 'crates',
	label: 'crates.io',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://crates.io/crates/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://crates.io/api/v1/crates/${label}`, signal) };
	}
};

export const nuget: Provider = {
	id: 'nuget',
	label: 'NuGet',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://www.nuget.org/packages/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(`https://api.nuget.org/v3-flatcontainer/${label}/index.json`, signal)
		};
	}
};

export const bluesky: Provider = {
	id: 'bluesky',
	label: 'Bluesky',
	kind: 'social',
	name: (label) => `${label}.bsky.social`,
	url: (label) => `https://bsky.app/profile/${label}.bsky.social`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=${label}.bsky.social`,
				signal,
				[400, 404]
			)
		};
	}
};

export const mastodon: Provider = {
	id: 'mastodon',
	label: 'Mastodon (mastodon.social)',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://mastodon.social/@${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://mastodon.social/api/v1/accounts/lookup?acct=${encodeURIComponent(label)}`,
				signal
			)
		};
	}
};

export const providers: Provider[] = [
	...TLDS.map(rdap),
	npm,
	pypi,
	crates,
	nuget,
	github,
	bluesky,
	mastodon
];

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
	const cached = getCached(provider.id, label);
	if (cached && isFresh(cached)) {
		return { ...base, status: cached.status, detail: cached.detail };
	}
	try {
		const { status, detail } = await provider.check(label, signal);
		if (status !== 'unknown') setCached(provider.id, label, status, detail);
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
