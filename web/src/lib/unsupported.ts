// Web-side catalogue of platform name namespaces that the checker knows about
// but cannot resolve — sourced from the Candidates section of docs/platforms.md.
// These are not verdicts: they are never counted as available/taken/unknown and
// are mixed into the Paid or Free card like any other row, minus the verdict, so
// an unsupported platform is not mistaken for a missing one. `paid` follows the
// server's meaning — the platform charges money to claim the name — so the
// registry-style namespaces land in the Paid card and the free handles in Free.
//
// `url` is a template; `{name}` is replaced with the checked label and the row
// links to the resulting page. Follow-up: derive this catalogue server-side from
// the provider registry plus a table so it cannot drift from the sources.
export interface UnsupportedSource {
	id: string;
	platformName: string;
	url: string;
	paid: boolean;
}

export const UNSUPPORTED: UnsupportedSource[] = [
	// needs key
	{
		id: 'unstoppable-domains',
		platformName: 'Unstoppable Domains',
		url: 'https://unstoppabledomains.com/domains/{name}',
		paid: true
	},
	{
		id: 'product-hunt',
		platformName: 'Product Hunt',
		url: 'https://www.producthunt.com/@{name}',
		paid: false
	},
	{
		id: 'tumblr',
		platformName: 'Tumblr',
		url: 'https://{name}.tumblr.com',
		paid: false
	},
	{
		id: 'vk',
		platformName: 'VK',
		url: 'https://vk.com/{name}',
		paid: false
	},
	{
		id: 'line',
		platformName: 'Line',
		url: 'https://line.me/ti/p/~{name}',
		paid: false
	},
	{
		id: 'kakao',
		platformName: 'Kakao',
		url: 'https://open.kakao.com/o/{name}',
		paid: false
	},
	{
		id: 'spotify-artist',
		platformName: 'Spotify (artist)',
		url: 'https://open.spotify.com/artist/{name}',
		paid: false
	},
	{
		id: 'apple-music',
		platformName: 'Apple Music',
		url: 'https://music.apple.com/us/artist/{name}',
		paid: false
	},
	{
		id: 'lastfm',
		platformName: 'Last.fm',
		url: 'https://www.last.fm/user/{name}',
		paid: false
	},
	{
		id: 'trakt',
		platformName: 'Trakt',
		url: 'https://trakt.tv/users/{name}',
		paid: false
	},
	{
		id: 'genius',
		platformName: 'Genius',
		url: 'https://genius.com/{name}',
		paid: false
	},
	{
		id: 'strava',
		platformName: 'Strava',
		url: 'https://www.strava.com/athletes/{name}',
		paid: false
	},

	// bot-blocked
	{
		id: 'medium',
		platformName: 'Medium',
		url: 'https://medium.com/@{name}',
		paid: false
	},
	{
		id: 'patreon',
		platformName: 'Patreon',
		url: 'https://www.patreon.com/{name}',
		paid: false
	},
	{
		id: 'kickstarter',
		platformName: 'Kickstarter',
		url: 'https://www.kickstarter.com/profile/{name}',
		paid: false
	},
	{
		id: 'vimeo',
		platformName: 'Vimeo',
		url: 'https://vimeo.com/{name}',
		paid: false
	},
	{
		id: 'lemon-squeezy',
		platformName: 'Lemon Squeezy',
		url: 'https://{name}.lemonsqueezy.com',
		paid: false
	},
	{
		id: 'codesandbox',
		platformName: 'CodeSandbox',
		url: 'https://codesandbox.io/u/{name}',
		paid: false
	},
	{
		id: 'npm-users',
		platformName: 'npm users',
		url: 'https://www.npmjs.com/~{name}',
		paid: false
	},
	{
		id: 'pypi-users',
		platformName: 'PyPI users',
		url: 'https://pypi.org/user/{name}/',
		paid: false
	},
	{
		id: 'swift-package-index',
		platformName: 'Swift Package Index',
		url: 'https://swiftpackageindex.com/{name}',
		paid: false
	},

	// no public API
	{
		id: 'discord',
		platformName: 'Discord',
		url: 'https://discord.com/users/{name}',
		paid: false
	},
	{
		id: 'linkedin',
		platformName: 'LinkedIn',
		url: 'https://www.linkedin.com/in/{name}',
		paid: false
	},
	{
		id: 'figma',
		platformName: 'Figma Community',
		url: 'https://www.figma.com/@{name}',
		paid: false
	},
	{
		id: 'minecraft',
		platformName: 'Minecraft',
		url: 'https://namemc.com/profile/{name}',
		paid: false
	},
	{
		id: 'roblox',
		platformName: 'Roblox',
		url: 'https://www.roblox.com/user.aspx?username={name}',
		paid: false
	},
	{
		id: 'fandom',
		platformName: 'Fandom',
		url: 'https://{name}.fandom.com',
		paid: false
	},

	// not checked
	{
		id: 'steam',
		platformName: 'Steam',
		url: 'https://steamcommunity.com/id/{name}',
		paid: false
	},
	{
		id: 'epic-games',
		platformName: 'Epic Games',
		url: 'https://www.epicgames.com/id/{name}',
		paid: false
	},
	{
		id: 'etsy',
		platformName: 'Etsy',
		url: 'https://www.etsy.com/shop/{name}',
		paid: false
	},
	{
		id: 'amazon',
		platformName: 'Amazon',
		url: 'https://www.amazon.com/{name}',
		paid: false
	},
	{
		id: 'ebay',
		platformName: 'eBay',
		url: 'https://www.ebay.com/usr/{name}',
		paid: false
	},
	{
		id: 'imdb',
		platformName: 'IMDb',
		url: 'https://www.imdb.com/user/{name}',
		paid: false
	},
	{
		id: 'bandcamp',
		platformName: 'Bandcamp',
		url: 'https://{name}.bandcamp.com',
		paid: false
	},
	{
		id: 'itchio',
		platformName: 'Itch.io',
		url: 'https://{name}.itch.io',
		paid: false
	},
	{
		id: 'codepen',
		platformName: 'CodePen',
		url: 'https://codepen.io/{name}',
		paid: false
	},
	{
		id: 'stackblitz',
		platformName: 'StackBlitz',
		url: 'https://stackblitz.com/@{name}',
		paid: false
	},
	{
		id: 'observable',
		platformName: 'Observable',
		url: 'https://observablehq.com/@{name}',
		paid: false
	},
	{
		id: 'deviantart',
		platformName: 'DeviantArt',
		url: 'https://www.deviantart.com/{name}',
		paid: false
	},
	{
		id: 'flickr',
		platformName: 'Flickr',
		url: 'https://www.flickr.com/people/{name}',
		paid: false
	},
	{
		id: 'reddit',
		platformName: 'Reddit',
		url: 'https://www.reddit.com/user/{name}',
		paid: false
	},
	{
		id: 'tiktok',
		platformName: 'TikTok',
		url: 'https://www.tiktok.com/@{name}',
		paid: false
	},
	{
		id: 'threads',
		platformName: 'Threads',
		url: 'https://www.threads.net/@{name}',
		paid: false
	},
	{
		id: 'facebook',
		platformName: 'Facebook',
		url: 'https://www.facebook.com/{name}',
		paid: false
	},
	{
		id: 'snapchat',
		platformName: 'Snapchat',
		url: 'https://www.snapchat.com/add/{name}',
		paid: false
	},
	{
		id: 'pinterest',
		platformName: 'Pinterest',
		url: 'https://www.pinterest.com/{name}',
		paid: false
	}
];
