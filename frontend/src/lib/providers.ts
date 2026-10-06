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

export const TLDS = ['.com', '.org', '.dev', '.io', '.ai', '.app'] as const;

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
							: 'unknown'
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
	availableCodes: number[] = [404],
	takenCodes: number[] = [200]
): Promise<Status> {
	const res = await fetch(url, {
		signal,
		referrerPolicy: 'no-referrer',
		credentials: 'omit',
		headers: { Accept: 'application/json' }
	});
	res.body?.cancel().catch(() => {});
	if (takenCodes.includes(res.status)) return 'taken';
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

export const gitlab: Provider = {
	id: 'gitlab',
	label: 'GitLab',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://gitlab.com/${label}`,
	async check(label, signal) {
		const res = await fetch(
			`https://gitlab.com/api/v4/users?username=${encodeURIComponent(label)}`,
			{
				signal,
				referrerPolicy: 'no-referrer',
				credentials: 'omit',
				headers: { Accept: 'application/json' }
			}
		);
		if (res.status !== 200) {
			res.body?.cancel().catch(() => {});
			return { status: 'unknown' };
		}
		const data = await res.json().catch(() => null);
		if (!Array.isArray(data)) return { status: 'unknown' };
		if (data.length > 0) return { status: 'taken' };
		const group = await checkHttp(
			`https://gitlab.com/api/v4/groups/${encodeURIComponent(label)}`,
			signal
		);
		return { status: group };
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

export const hex: Provider = {
	id: 'hex',
	label: 'Hex',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://hex.pm/packages/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://hex.pm/api/packages/${label}`, signal) };
	}
};

export const pub: Provider = {
	id: 'pub',
	label: 'pub.dev',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://pub.dev/packages/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://pub.dev/api/packages/${label}`, signal) };
	}
};

export const homebrew: Provider = {
	id: 'homebrew',
	label: 'Homebrew',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://formulae.brew.sh/formula/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(`https://formulae.brew.sh/api/formula/${label}.json`, signal)
		};
	}
};

export const firefox: Provider = {
	id: 'firefox',
	label: 'Firefox Add-ons',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://addons.mozilla.org/firefox/addon/${label}/`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://addons.mozilla.org/api/v5/addons/addon/${label}/`,
				signal,
				[404],
				[200, 401]
			)
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

export const jsr: Provider = {
	id: 'jsr',
	label: 'JSR',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://jsr.io/@${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://jsr.io/api/scopes/${label}`, signal) };
	}
};

export const wordpress: Provider = {
	id: 'wordpress',
	label: 'WordPress.org',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://wordpress.org/plugins/${label}/`,
	async check(label, signal) {
		const res = await fetch(
			`https://api.wordpress.org/plugins/info/1.2/?action=plugin_information&request%5Bslug%5D=${label}`,
			{
				signal,
				referrerPolicy: 'no-referrer',
				credentials: 'omit',
				headers: { Accept: 'application/json' }
			}
		);
		if (res.status === 200) {
			res.body?.cancel().catch(() => {});
			return { status: 'taken' };
		}
		if (res.status === 404) {
			const data = await res.json().catch(() => null);
			const closed = (data as { closed?: unknown })?.closed === true;
			return { status: closed ? 'taken' : 'available' };
		}
		res.body?.cancel().catch(() => {});
		return { status: 'unknown' };
	}
};

export const bitbucket: Provider = {
	id: 'bitbucket',
	label: 'Bitbucket',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://bitbucket.org/${label}/`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://api.bitbucket.org/2.0/workspaces/${encodeURIComponent(label)}`,
				signal,
				[404],
				[200, 403]
			)
		};
	}
};

export const codeberg: Provider = {
	id: 'codeberg',
	label: 'Codeberg',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://codeberg.org/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://codeberg.org/api/v1/users/${encodeURIComponent(label)}`,
				signal
			)
		};
	}
};

export const wikipedia: Provider = {
	id: 'wikipedia',
	label: 'Wikipedia',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://en.wikipedia.org/wiki/User:${label}`,
	async check(label, signal) {
		const res = await fetch(
			`https://en.wikipedia.org/w/api.php?action=query&list=users&ususers=${encodeURIComponent(label)}&format=json&origin=*`,
			{
				signal,
				referrerPolicy: 'no-referrer',
				credentials: 'omit',
				headers: { Accept: 'application/json' }
			}
		);
		if (res.status !== 200) {
			res.body?.cancel().catch(() => {});
			return { status: 'unknown' };
		}
		const data = await res.json().catch(() => null);
		const users = (data as { query?: { users?: unknown[] } })?.query?.users;
		if (!Array.isArray(users) || users.length === 0) return { status: 'unknown' };
		const user = users[0] as { missing?: unknown };
		return { status: 'missing' in user ? 'available' : 'taken' };
	}
};

export const devto: Provider = {
	id: 'devto',
	label: 'DEV',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://dev.to/${label}`,
	async check(label, signal) {
		const user = await checkHttp(
			`https://dev.to/api/users/by_username?url=${encodeURIComponent(label)}`,
			signal
		);
		if (user !== 'available') return { status: user };
		const org = await checkHttp(
			`https://dev.to/api/organizations/${encodeURIComponent(label)}`,
			signal
		);
		return { status: org };
	}
};

export const farcaster: Provider = {
	id: 'farcaster',
	label: 'Farcaster',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://farcaster.xyz/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://fnames.farcaster.xyz/transfers/current?name=${encodeURIComponent(label)}`,
				signal
			)
		};
	}
};

export const hackernews: Provider = {
	id: 'hackernews',
	label: 'Hacker News',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://news.ycombinator.com/user?id=${label}`,
	async check(label, signal) {
		const res = await fetch(
			`https://hacker-news.firebaseio.com/v0/user/${encodeURIComponent(label)}.json`,
			{
				signal,
				referrerPolicy: 'no-referrer',
				credentials: 'omit',
				headers: { Accept: 'application/json' }
			}
		);
		if (res.status !== 200) {
			res.body?.cancel().catch(() => {});
			return { status: 'unknown' };
		}
		const data = await res.json().catch(() => undefined);
		if (data === null) return { status: 'available' };
		if (data) return { status: 'taken' };
		return { status: 'unknown' };
	}
};

export const opencollective: Provider = {
	id: 'opencollective',
	label: 'Open Collective',
	kind: 'org',
	name: (label) => label,
	url: (label) => `https://opencollective.com/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://opencollective.com/${encodeURIComponent(label)}.json`,
				signal
			)
		};
	}
};

export const minehut: Provider = {
	id: 'minehut',
	label: 'Minehut',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://www.minehut.com/servers/${label}`,
	async check(label, signal) {
		const res = await fetch(
			`https://api.minehut.com/server/${encodeURIComponent(label)}?byName=true`,
			{
				signal,
				referrerPolicy: 'no-referrer',
				credentials: 'omit',
				headers: { Accept: 'application/json' }
			}
		);
		if (res.status !== 200) {
			res.body?.cancel().catch(() => {});
			return { status: 'unknown' };
		}
		const data = await res.json().catch(() => undefined);
		const server = (data as { server?: unknown } | undefined)?.server;
		if (server === null) return { status: 'available' };
		if (server) return { status: 'taken' };
		return { status: 'unknown' };
	}
};

export const lichess: Provider = {
	id: 'lichess',
	label: 'Lichess',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://lichess.org/@/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(`https://lichess.org/api/user/${encodeURIComponent(label)}`, signal)
		};
	}
};

export const chesscom: Provider = {
	id: 'chesscom',
	label: 'Chess.com',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://www.chess.com/member/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://api.chess.com/pub/player/${encodeURIComponent(label)}`,
				signal
			)
		};
	}
};

export const huggingface: Provider = {
	id: 'huggingface',
	label: 'Hugging Face',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://huggingface.co/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://huggingface.co/api/users/${encodeURIComponent(label)}/overview`,
				signal
			)
		};
	}
};

export const dailymotion: Provider = {
	id: 'dailymotion',
	label: 'Dailymotion',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://www.dailymotion.com/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(
				`https://api.dailymotion.com/user/${encodeURIComponent(label)}`,
				signal
			)
		};
	}
};

export const mixcloud: Provider = {
	id: 'mixcloud',
	label: 'Mixcloud',
	kind: 'social',
	name: (label) => label,
	url: (label) => `https://www.mixcloud.com/${label}/`,
	async check(label, signal) {
		return {
			status: await checkHttp(`https://api.mixcloud.com/${encodeURIComponent(label)}/`, signal)
		};
	}
};

export const openvsx: Provider = {
	id: 'openvsx',
	label: 'Open VSX',
	kind: 'dev',
	name: (label) => label,
	url: (label) => `https://open-vsx.org/namespace/${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(`https://open-vsx.org/api/${encodeURIComponent(label)}`, signal)
		};
	}
};

export const providers: Provider[] = [
	...TLDS.map(rdap),
	npm,
	pypi,
	crates,
	nuget,
	hex,
	pub,
	homebrew,
	firefox,
	jsr,
	wordpress,
	openvsx,
	github,
	gitlab,
	bitbucket,
	codeberg,
	wikipedia,
	devto,
	farcaster,
	hackernews,
	minehut,
	lichess,
	chesscom,
	huggingface,
	dailymotion,
	mixcloud,
	bluesky,
	mastodon,
	opencollective
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
