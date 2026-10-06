import { lookupDomain, UnsupportedTldError } from './rdap';
import { getCached, isFresh, setCached } from './providerCache';

export type Status = 'available' | 'taken' | 'unknown';

export const PLATFORMS = {
	domains: 'Domains',
	npm: 'npm',
	pypi: 'PyPI',
	crates: 'crates.io',
	nuget: 'NuGet',
	hex: 'Hex',
	pub: 'pub.dev',
	homebrew: 'Homebrew',
	firefox: 'Firefox Add-ons',
	jsr: 'JSR',
	wordpress: 'WordPress.org',
	openvsx: 'Open VSX',
	github: 'GitHub',
	gitlab: 'GitLab',
	bitbucket: 'Bitbucket',
	codeberg: 'Codeberg',
	wikipedia: 'Wikipedia',
	devto: 'DEV',
	farcaster: 'Farcaster',
	minehut: 'Minehut',
	lichess: 'Lichess',
	chesscom: 'Chess.com',
	huggingface: 'Hugging Face',
	dailymotion: 'Dailymotion',
	mixcloud: 'Mixcloud',
	bluesky: 'Bluesky',
	mastodon: 'Mastodon',
	opencollective: 'Open Collective'
} as const;

export type PlatformId = keyof typeof PLATFORMS;

export interface Provider {
	id: string;
	platform: PlatformId;
	minLength?: number;
	minLengthSource?: string;
	name: (label: string) => string;
	url: (label: string) => string;
	check: (label: string, signal?: AbortSignal) => Promise<{ status: Status; detail?: string }>;
}

export interface ProviderResult {
	id: string;
	platform: PlatformId;
	name: string;
	url: string;
	status: Status;
	detail?: string;
}

export const TLDS = ['.com', '.org', '.dev', '.io', '.ai', '.app'] as const;

const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function toLabel(input: string): string | null {
	const value = input.trim().toLowerCase();
	return LABEL.test(value) ? value : null;
}

const TLD_MIN_LENGTH: Record<string, { min: number; source: string }> = {
	'.dev': {
		min: 2,
		source: 'Google Registry reserves 1-character .dev labels (no 1-char .dev is registered)'
	}
};

// RFC 6761 / ICANN Spec 5 special-use labels, reserved in every TLD. Some
// registries (e.g. Google's .dev/.app) reserve them without creating an RDAP
// record, so a 404 would otherwise read as available.
const RESERVED_LABELS = new Set(['example', 'test', 'invalid', 'localhost']);

const rdap = (tld: string): Provider => {
	const rule = TLD_MIN_LENGTH[tld];
	return {
		id: `domain${tld}`,
		platform: 'domains',
		minLength: rule?.min,
		minLengthSource: rule?.source,
		name: (label) => `${label}${tld}`,
		url: (label) => `https://${label}${tld}`,
		async check(label, signal) {
			if (RESERVED_LABELS.has(label)) return { status: 'taken' };
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
	};
};

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
	platform: 'github',
	minLength: 2,
	minLengthSource:
		'GitHub 1-character usernames are effectively unavailable (32/36 are claimed; the rest reserved)',
	name: (label) => label,
	url: (label) => `https://github.com/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://api.github.com/users/${label}`, signal) };
	}
};

export const gitlab: Provider = {
	id: 'gitlab',
	platform: 'gitlab',
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
	platform: 'npm',
	name: (label) => label,
	url: (label) => `https://www.npmjs.com/package/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://registry.npmjs.org/${label}`, signal) };
	}
};

export const pypi: Provider = {
	id: 'pypi',
	platform: 'pypi',
	name: (label) => label,
	url: (label) => `https://pypi.org/project/${label}/`,
	async check(label, signal) {
		return { status: await checkHttp(`https://pypi.org/pypi/${label}/json`, signal) };
	}
};

export const crates: Provider = {
	id: 'crates',
	platform: 'crates',
	name: (label) => label,
	url: (label) => `https://crates.io/crates/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://crates.io/api/v1/crates/${label}`, signal) };
	}
};

export const nuget: Provider = {
	id: 'nuget',
	platform: 'nuget',
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
	platform: 'hex',
	minLength: 2,
	minLengthSource: 'hexpm package name validation: min 2 chars (repository/package.ex)',
	name: (label) => label,
	url: (label) => `https://hex.pm/packages/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://hex.pm/api/packages/${label}`, signal) };
	}
};

export const hexUser: Provider = {
	id: 'hex-user',
	platform: 'hex',
	minLength: 3,
	minLengthSource: 'hexpm username validation: min 3 chars (accounts/user.ex)',
	name: (label) => label,
	url: (label) => `https://hex.pm/users/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://hex.pm/api/users/${label}`, signal) };
	}
};

export const pub: Provider = {
	id: 'pub',
	platform: 'pub',
	name: (label) => label,
	url: (label) => `https://pub.dev/packages/${label}`,
	async check(label, signal) {
		return { status: await checkHttp(`https://pub.dev/api/packages/${label}`, signal) };
	}
};

export const homebrew: Provider = {
	id: 'homebrew',
	platform: 'homebrew',
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
	platform: 'firefox',
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
	platform: 'bluesky',
	minLength: 3,
	minLengthSource: 'bsky.social help: "Handle Too Short – Must be at least 3 characters"',
	name: (label) => `${label}.bsky.social`,
	url: (label) => `https://bsky.app/profile/${label}.bsky.social`,
	async check(label, signal) {
		const res = await fetch(
			`https://bsky.social/xrpc/com.atproto.temp.checkHandleAvailability?handle=${encodeURIComponent(label)}.bsky.social`,
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
		const type = (data as { result?: { $type?: unknown } } | null)?.result?.$type;
		if (typeof type !== 'string') return { status: 'unknown' };
		// resolveHandle can't see reserved/deactivated/over-length handles, so
		// ask the signup check instead.
		return { status: type.endsWith('#resultAvailable') ? 'available' : 'taken' };
	}
};

export const mastodon: Provider = {
	id: 'mastodon',
	platform: 'mastodon',
	name: (label) => label,
	url: (label) => `https://mastodon.social/@${label}`,
	async check(label, signal) {
		const res = await fetch(`https://mastodon.social/@${encodeURIComponent(label)}`, {
			signal,
			referrerPolicy: 'no-referrer',
			credentials: 'omit',
			headers: { Accept: 'text/html' }
		});
		res.body?.cancel().catch(() => {});
		// The account lookup API flattens suspended accounts to 404, so probe the
		// profile: 410 Gone means suspended, and the username is still taken.
		if (res.status === 200 || res.status === 410) return { status: 'taken' };
		if (res.status === 404) return { status: 'available' };
		return { status: 'unknown' };
	}
};

export const jsr: Provider = {
	id: 'jsr',
	platform: 'jsr',
	name: (label) => label,
	url: (label) => `https://jsr.io/@${label}`,
	async check(label, signal) {
		return {
			status: await checkHttp(`https://jsr.io/api/scopes/${label}`, signal, [404], [200, 400])
		};
	}
};

export const wordpress: Provider = {
	id: 'wordpress',
	platform: 'wordpress',
	minLength: 5,
	minLengthSource:
		'WordPress.org new plugin names require ≥5 alphanumeric characters (Plugin Check naming rules)',
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
	platform: 'bitbucket',
	minLength: 4,
	minLengthSource:
		'Bitbucket rejects workspace IDs under 4 characters at creation (1–3 rejected, 4 accepted)',
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
	platform: 'codeberg',
	name: (label) => label,
	url: (label) => `https://codeberg.org/${label}`,
	async check(label, signal) {
		const res = await fetch(`https://codeberg.org/api/v1/users/${encodeURIComponent(label)}`, {
			signal,
			referrerPolicy: 'no-referrer',
			credentials: 'omit',
			headers: { Accept: 'application/json' }
		});
		if (res.status === 200) {
			res.body?.cancel().catch(() => {});
			return { status: 'taken' };
		}
		if (res.status === 404) {
			const data = await res.json().catch(() => null);
			const message = (data as { message?: unknown } | null)?.message;
			if (typeof message !== 'string') return { status: 'unknown' };
			// Free names report the user "does not exist"; any other 404 body
			// (e.g. a limited-visibility account) means the name is occupied.
			return { status: message.includes('does not exist') ? 'available' : 'taken' };
		}
		res.body?.cancel().catch(() => {});
		return { status: 'unknown' };
	}
};

export const wikipedia: Provider = {
	id: 'wikipedia',
	platform: 'wikipedia',
	name: (label) => label,
	url: (label) => `https://en.wikipedia.org/wiki/Special:CentralAuth/${label}`,
	async check(label, signal) {
		const res = await fetch(
			`https://meta.wikimedia.org/w/api.php?action=query&meta=globaluserinfo&guiuser=${encodeURIComponent(label)}&format=json&origin=*`,
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
		const info = (data as { query?: { globaluserinfo?: { missing?: unknown } } })?.query
			?.globaluserinfo;
		if (!info) return { status: 'unknown' };
		return { status: 'missing' in info ? 'available' : 'taken' };
	}
};

export const devto: Provider = {
	id: 'devto',
	platform: 'devto',
	minLength: 2,
	minLengthSource: 'Forem/DEV usernames are at least 2 characters',
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
	platform: 'farcaster',
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

export const opencollective: Provider = {
	id: 'opencollective',
	platform: 'opencollective',
	name: (label) => label,
	url: (label) => `https://opencollective.com/${label}`,
	async check(label, signal) {
		const url = `https://opencollective.com/${encodeURIComponent(label)}.json`;
		const status = await checkHttp(url, signal);
		if (status !== 'unknown') return { status };
		// Their CDN can cache a transient 5xx per-Origin for up to an hour.
		// Retry once with a cache-busting param to bypass the poisoned entry.
		return { status: await checkHttp(`${url}?_=${Date.now()}`, signal) };
	}
};

export const minehut: Provider = {
	id: 'minehut',
	platform: 'minehut',
	minLength: 4,
	minLengthSource: 'Minehut server names require at least 4 characters (legacy names shorter)',
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
	platform: 'lichess',
	minLength: 2,
	minLengthSource: 'Lichess usernames are at least 2 characters',
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
	platform: 'chesscom',
	minLength: 3,
	minLengthSource: 'Chess.com usernames are 3–25 characters',
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
	platform: 'huggingface',
	minLength: 2,
	minLengthSource: 'Hugging Face: 2-char users exist, none 1-char (min 2 observed)',
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
	platform: 'dailymotion',
	minLength: 3,
	minLengthSource: 'Dailymotion: no 2-char users exist, 3-char do (min 3 observed)',
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
	platform: 'mixcloud',
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
	platform: 'openvsx',
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
	hexUser,
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
		platform: provider.platform,
		name: provider.name(label),
		url: provider.url(label)
	};
	if (provider.minLength && label.length < provider.minLength) {
		return { ...base, status: 'taken' as Status };
	}
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
