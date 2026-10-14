// Web-side catalogue of platform name namespaces that the checker knows about
// but cannot resolve — sourced from the Candidates section of docs/platforms.md.
// These are not verdicts: they are never counted as available/taken/unknown and
// are folded into the Paid or Free card (muted and non-interactive) so an
// unsupported platform is not mistaken for a missing one. `paid` follows the
// server's meaning — the platform charges money to claim the name — so the
// registry-style namespaces land in the Paid card and the free handles in Free.
//
// `url` is a display-only template (never linked); `{name}` is replaced with the
// checked label. Follow-up: derive this catalogue server-side from the provider
// registry plus a table so it cannot drift from the sources.
export type UnsupportedReason = 'needs key' | 'bot-blocked' | 'no public API' | 'not checked';

export interface UnsupportedSource {
	id: string;
	platformName: string;
	reason: UnsupportedReason;
	url: string;
	paid: boolean;
}

export const UNSUPPORTED: UnsupportedSource[] = [
	// needs key
	{
		id: 'unstoppable-domains',
		platformName: 'Unstoppable Domains',
		reason: 'needs key',
		url: 'https://unstoppabledomains.com/domains/{name}',
		paid: true
	},
	{
		id: 'product-hunt',
		platformName: 'Product Hunt',
		reason: 'needs key',
		url: 'https://www.producthunt.com/@{name}',
		paid: false
	},
	{
		id: 'tumblr',
		platformName: 'Tumblr',
		reason: 'needs key',
		url: 'https://{name}.tumblr.com',
		paid: false
	},
	{
		id: 'vk',
		platformName: 'VK',
		reason: 'needs key',
		url: 'https://vk.com/{name}',
		paid: false
	},
	{
		id: 'line',
		platformName: 'Line',
		reason: 'needs key',
		url: 'https://line.me/ti/p/~{name}',
		paid: false
	},
	{
		id: 'kakao',
		platformName: 'Kakao',
		reason: 'needs key',
		url: 'https://open.kakao.com/o/{name}',
		paid: false
	},
	{
		id: 'spotify-artist',
		platformName: 'Spotify (artist)',
		reason: 'needs key',
		url: 'https://open.spotify.com/artist/{name}',
		paid: false
	},
	{
		id: 'apple-music',
		platformName: 'Apple Music',
		reason: 'needs key',
		url: 'https://music.apple.com/us/artist/{name}',
		paid: false
	},
	{
		id: 'lastfm',
		platformName: 'Last.fm',
		reason: 'needs key',
		url: 'https://www.last.fm/user/{name}',
		paid: false
	},
	{
		id: 'trakt',
		platformName: 'Trakt',
		reason: 'needs key',
		url: 'https://trakt.tv/users/{name}',
		paid: false
	},
	{
		id: 'genius',
		platformName: 'Genius',
		reason: 'needs key',
		url: 'https://genius.com/{name}',
		paid: false
	},
	{
		id: 'strava',
		platformName: 'Strava',
		reason: 'needs key',
		url: 'https://www.strava.com/athletes/{name}',
		paid: false
	},

	// bot-blocked
	{
		id: 'medium',
		platformName: 'Medium',
		reason: 'bot-blocked',
		url: 'https://medium.com/@{name}',
		paid: false
	},
	{
		id: 'patreon',
		platformName: 'Patreon',
		reason: 'bot-blocked',
		url: 'https://www.patreon.com/{name}',
		paid: false
	},
	{
		id: 'kickstarter',
		platformName: 'Kickstarter',
		reason: 'bot-blocked',
		url: 'https://www.kickstarter.com/profile/{name}',
		paid: false
	},
	{
		id: 'vimeo',
		platformName: 'Vimeo',
		reason: 'bot-blocked',
		url: 'https://vimeo.com/{name}',
		paid: false
	},
	{
		id: 'lemon-squeezy',
		platformName: 'Lemon Squeezy',
		reason: 'bot-blocked',
		url: 'https://{name}.lemonsqueezy.com',
		paid: false
	},
	{
		id: 'codesandbox',
		platformName: 'CodeSandbox',
		reason: 'bot-blocked',
		url: 'https://codesandbox.io/u/{name}',
		paid: false
	},
	{
		id: 'npm-users',
		platformName: 'npm users',
		reason: 'bot-blocked',
		url: 'https://www.npmjs.com/~{name}',
		paid: false
	},
	{
		id: 'pypi-users',
		platformName: 'PyPI users',
		reason: 'bot-blocked',
		url: 'https://pypi.org/user/{name}/',
		paid: false
	},
	{
		id: 'swift-package-index',
		platformName: 'Swift Package Index',
		reason: 'bot-blocked',
		url: 'https://swiftpackageindex.com/{name}',
		paid: false
	},

	// no public API
	{
		id: 'discord',
		platformName: 'Discord',
		reason: 'no public API',
		url: 'https://discord.com/users/{name}',
		paid: false
	},
	{
		id: 'linkedin',
		platformName: 'LinkedIn',
		reason: 'no public API',
		url: 'https://www.linkedin.com/in/{name}',
		paid: false
	},
	{
		id: 'figma',
		platformName: 'Figma Community',
		reason: 'no public API',
		url: 'https://www.figma.com/@{name}',
		paid: false
	},
	{
		id: 'minecraft',
		platformName: 'Minecraft',
		reason: 'no public API',
		url: 'https://namemc.com/profile/{name}',
		paid: false
	},
	{
		id: 'roblox',
		platformName: 'Roblox',
		reason: 'no public API',
		url: 'https://www.roblox.com/user.aspx?username={name}',
		paid: false
	},
	{
		id: 'fandom',
		platformName: 'Fandom',
		reason: 'no public API',
		url: 'https://{name}.fandom.com',
		paid: false
	},

	// not checked
	{
		id: 'steam',
		platformName: 'Steam',
		reason: 'not checked',
		url: 'https://steamcommunity.com/id/{name}',
		paid: false
	},
	{
		id: 'epic-games',
		platformName: 'Epic Games',
		reason: 'not checked',
		url: 'https://www.epicgames.com/id/{name}',
		paid: false
	},
	{
		id: 'etsy',
		platformName: 'Etsy',
		reason: 'not checked',
		url: 'https://www.etsy.com/shop/{name}',
		paid: false
	},
	{
		id: 'amazon',
		platformName: 'Amazon',
		reason: 'not checked',
		url: 'https://www.amazon.com/{name}',
		paid: false
	},
	{
		id: 'ebay',
		platformName: 'eBay',
		reason: 'not checked',
		url: 'https://www.ebay.com/usr/{name}',
		paid: false
	},
	{
		id: 'imdb',
		platformName: 'IMDb',
		reason: 'not checked',
		url: 'https://www.imdb.com/user/{name}',
		paid: false
	},
	{
		id: 'bandcamp',
		platformName: 'Bandcamp',
		reason: 'not checked',
		url: 'https://{name}.bandcamp.com',
		paid: false
	},
	{
		id: 'itchio',
		platformName: 'Itch.io',
		reason: 'not checked',
		url: 'https://{name}.itch.io',
		paid: false
	},
	{
		id: 'codepen',
		platformName: 'CodePen',
		reason: 'not checked',
		url: 'https://codepen.io/{name}',
		paid: false
	},
	{
		id: 'stackblitz',
		platformName: 'StackBlitz',
		reason: 'not checked',
		url: 'https://stackblitz.com/@{name}',
		paid: false
	},
	{
		id: 'observable',
		platformName: 'Observable',
		reason: 'not checked',
		url: 'https://observablehq.com/@{name}',
		paid: false
	},
	{
		id: 'deviantart',
		platformName: 'DeviantArt',
		reason: 'not checked',
		url: 'https://www.deviantart.com/{name}',
		paid: false
	},
	{
		id: 'flickr',
		platformName: 'Flickr',
		reason: 'not checked',
		url: 'https://www.flickr.com/people/{name}',
		paid: false
	},
	{
		id: 'reddit',
		platformName: 'Reddit',
		reason: 'not checked',
		url: 'https://www.reddit.com/user/{name}',
		paid: false
	},
	{
		id: 'tiktok',
		platformName: 'TikTok',
		reason: 'not checked',
		url: 'https://www.tiktok.com/@{name}',
		paid: false
	},
	{
		id: 'threads',
		platformName: 'Threads',
		reason: 'not checked',
		url: 'https://www.threads.net/@{name}',
		paid: false
	},
	{
		id: 'facebook',
		platformName: 'Facebook',
		reason: 'not checked',
		url: 'https://www.facebook.com/{name}',
		paid: false
	},
	{
		id: 'snapchat',
		platformName: 'Snapchat',
		reason: 'not checked',
		url: 'https://www.snapchat.com/add/{name}',
		paid: false
	},
	{
		id: 'pinterest',
		platformName: 'Pinterest',
		reason: 'not checked',
		url: 'https://www.pinterest.com/{name}',
		paid: false
	}
];
