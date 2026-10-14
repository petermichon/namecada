export type Status = 'available' | 'taken' | 'unknown';

// PLATFORMS maps a platform id to its display name. The backend decides which
// providers run; this only labels the results it streams back.
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
	opencollective: 'Open Collective',
	lemmy: 'Lemmy',
	peertube: 'PeerTube'
} as const;

export type PlatformId = keyof typeof PLATFORMS;

export interface ProviderResult {
	id: string;
	platform: PlatformId;
	name: string;
	url: string;
	status: Status;
	detail?: string;
}

const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function toLabel(input: string): string | null {
	const value = input.trim().toLowerCase();
	return LABEL.test(value) ? value : null;
}
