// Recent labels are the user's own searches, kept in localStorage so the
// landing page can offer them again. They are independent of the engine cache:
// that keeps verdicts for an hour, while a label stays worth reopening for much
// longer.

export const RECENT_LIMIT = 6;
const KEY = 'namecada:recent';

type Store = Pick<Storage, 'getItem' | 'setItem'>;

// storage returns localStorage, or null outside a browser (tests, SSR) and when
// the store is blocked (private mode, a locked-down WebView).
function storage(): Store | null {
	if (typeof window === 'undefined') return null;
	try {
		return window.localStorage;
	} catch {
		return null;
	}
}

export function loadRecent(target: Store | null = storage()): string[] {
	if (!target) return [];
	let raw: unknown;
	try {
		raw = JSON.parse(target.getItem(KEY) ?? '[]');
	} catch {
		return [];
	}
	if (!Array.isArray(raw)) return [];

	const seen = new Set<string>();
	const labels: string[] = [];
	for (const item of raw) {
		if (typeof item !== 'string') continue;
		const clean = item.trim();
		const key = clean.toLowerCase();
		if (!key || seen.has(key)) continue;
		seen.add(key);
		labels.push(clean);
		if (labels.length === RECENT_LIMIT) break;
	}
	return labels;
}

// rememberLabel puts a label first, drops any case-insensitive repeat and caps
// the list. A store that refuses the write still yields the new list in memory.
export function rememberLabel(label: string, target: Store | null = storage()): string[] {
	const clean = label.trim();
	const rest = loadRecent(target).filter((item) => item.toLowerCase() !== clean.toLowerCase());
	const next = clean ? [clean, ...rest].slice(0, RECENT_LIMIT) : rest;
	try {
		target?.setItem(KEY, JSON.stringify(next));
	} catch {
		// Keep the in-memory value; only persistence failed.
	}
	return next;
}
