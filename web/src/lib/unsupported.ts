// Web-side catalogue of platform name namespaces that the checker knows about
// but cannot resolve — sourced from the Candidates section of docs/platforms.md.
// These are not verdicts: they are never counted as available/taken/unknown and
// are shown as a disabled group so an unsupported platform is not mistaken for
// a missing one.
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
}

export const UNSUPPORTED: UnsupportedSource[] = [
	// needs key
	{
		id: 'unstoppable-domains',
		platformName: 'Unstoppable Domains',
		reason: 'needs key',
		url: 'https://unstoppabledomains.com/domains/{name}'
	},
	{
		id: 'product-hunt',
		platformName: 'Product Hunt',
		reason: 'needs key',
		url: 'https://www.producthunt.com/@{name}'
	},
	{ id: 'tumblr', platformName: 'Tumblr', reason: 'needs key', url: 'https://{name}.tumblr.com' },
	{ id: 'vk', platformName: 'VK', reason: 'needs key', url: 'https://vk.com/{name}' },
	{ id: 'line', platformName: 'Line', reason: 'needs key', url: 'https://line.me/ti/p/~{name}' },
	{
		id: 'kakao',
		platformName: 'Kakao',
		reason: 'needs key',
		url: 'https://open.kakao.com/o/{name}'
	},
	{
		id: 'spotify-artist',
		platformName: 'Spotify (artist)',
		reason: 'needs key',
		url: 'https://open.spotify.com/artist/{name}'
	},
	{
		id: 'apple-music',
		platformName: 'Apple Music',
		reason: 'needs key',
		url: 'https://music.apple.com/us/artist/{name}'
	},
	{
		id: 'lastfm',
		platformName: 'Last.fm',
		reason: 'needs key',
		url: 'https://www.last.fm/user/{name}'
	},
	{ id: 'trakt', platformName: 'Trakt', reason: 'needs key', url: 'https://trakt.tv/users/{name}' },
	{ id: 'genius', platformName: 'Genius', reason: 'needs key', url: 'https://genius.com/{name}' },
	{
		id: 'strava',
		platformName: 'Strava',
		reason: 'needs key',
		url: 'https://www.strava.com/athletes/{name}'
	},

	// bot-blocked
	{
		id: 'medium',
		platformName: 'Medium',
		reason: 'bot-blocked',
		url: 'https://medium.com/@{name}'
	},
	{
		id: 'patreon',
		platformName: 'Patreon',
		reason: 'bot-blocked',
		url: 'https://www.patreon.com/{name}'
	},
	{
		id: 'kickstarter',
		platformName: 'Kickstarter',
		reason: 'bot-blocked',
		url: 'https://www.kickstarter.com/profile/{name}'
	},
	{ id: 'vimeo', platformName: 'Vimeo', reason: 'bot-blocked', url: 'https://vimeo.com/{name}' },
	{
		id: 'lemon-squeezy',
		platformName: 'Lemon Squeezy',
		reason: 'bot-blocked',
		url: 'https://{name}.lemonsqueezy.com'
	},
	{
		id: 'codesandbox',
		platformName: 'CodeSandbox',
		reason: 'bot-blocked',
		url: 'https://codesandbox.io/u/{name}'
	},
	{
		id: 'npm-users',
		platformName: 'npm users',
		reason: 'bot-blocked',
		url: 'https://www.npmjs.com/~{name}'
	},
	{
		id: 'pypi-users',
		platformName: 'PyPI users',
		reason: 'bot-blocked',
		url: 'https://pypi.org/user/{name}/'
	},
	{
		id: 'swift-package-index',
		platformName: 'Swift Package Index',
		reason: 'bot-blocked',
		url: 'https://swiftpackageindex.com/{name}'
	},

	// no public API
	{
		id: 'discord',
		platformName: 'Discord',
		reason: 'no public API',
		url: 'https://discord.com/users/{name}'
	},
	{
		id: 'linkedin',
		platformName: 'LinkedIn',
		reason: 'no public API',
		url: 'https://www.linkedin.com/in/{name}'
	},
	{
		id: 'figma',
		platformName: 'Figma Community',
		reason: 'no public API',
		url: 'https://www.figma.com/@{name}'
	},
	{
		id: 'minecraft',
		platformName: 'Minecraft',
		reason: 'no public API',
		url: 'https://namemc.com/profile/{name}'
	},
	{
		id: 'roblox',
		platformName: 'Roblox',
		reason: 'no public API',
		url: 'https://www.roblox.com/user.aspx?username={name}'
	},
	{
		id: 'fandom',
		platformName: 'Fandom',
		reason: 'no public API',
		url: 'https://{name}.fandom.com'
	},

	// not checked
	{
		id: 'steam',
		platformName: 'Steam',
		reason: 'not checked',
		url: 'https://steamcommunity.com/id/{name}'
	},
	{
		id: 'epic-games',
		platformName: 'Epic Games',
		reason: 'not checked',
		url: 'https://www.epicgames.com/id/{name}'
	},
	{
		id: 'etsy',
		platformName: 'Etsy',
		reason: 'not checked',
		url: 'https://www.etsy.com/shop/{name}'
	},
	{
		id: 'amazon',
		platformName: 'Amazon',
		reason: 'not checked',
		url: 'https://www.amazon.com/{name}'
	},
	{
		id: 'ebay',
		platformName: 'eBay',
		reason: 'not checked',
		url: 'https://www.ebay.com/usr/{name}'
	},
	{
		id: 'imdb',
		platformName: 'IMDb',
		reason: 'not checked',
		url: 'https://www.imdb.com/user/{name}'
	},
	{
		id: 'bandcamp',
		platformName: 'Bandcamp',
		reason: 'not checked',
		url: 'https://{name}.bandcamp.com'
	},
	{ id: 'itchio', platformName: 'Itch.io', reason: 'not checked', url: 'https://{name}.itch.io' },
	{
		id: 'codepen',
		platformName: 'CodePen',
		reason: 'not checked',
		url: 'https://codepen.io/{name}'
	},
	{
		id: 'stackblitz',
		platformName: 'StackBlitz',
		reason: 'not checked',
		url: 'https://stackblitz.com/@{name}'
	},
	{
		id: 'observable',
		platformName: 'Observable',
		reason: 'not checked',
		url: 'https://observablehq.com/@{name}'
	},
	{
		id: 'deviantart',
		platformName: 'DeviantArt',
		reason: 'not checked',
		url: 'https://www.deviantart.com/{name}'
	},
	{
		id: 'flickr',
		platformName: 'Flickr',
		reason: 'not checked',
		url: 'https://www.flickr.com/people/{name}'
	},
	{
		id: 'reddit',
		platformName: 'Reddit',
		reason: 'not checked',
		url: 'https://www.reddit.com/user/{name}'
	},
	{
		id: 'tiktok',
		platformName: 'TikTok',
		reason: 'not checked',
		url: 'https://www.tiktok.com/@{name}'
	},
	{
		id: 'threads',
		platformName: 'Threads',
		reason: 'not checked',
		url: 'https://www.threads.net/@{name}'
	},
	{
		id: 'facebook',
		platformName: 'Facebook',
		reason: 'not checked',
		url: 'https://www.facebook.com/{name}'
	},
	{
		id: 'snapchat',
		platformName: 'Snapchat',
		reason: 'not checked',
		url: 'https://www.snapchat.com/add/{name}'
	},
	{
		id: 'pinterest',
		platformName: 'Pinterest',
		reason: 'not checked',
		url: 'https://www.pinterest.com/{name}'
	}
];
