import { onMount } from 'svelte';
import { isAbort, streamFromApi } from './api';
import { toLabel, type Status } from './platforms';

const QUERY_KEY = 'domain-browser:query';

export interface Row {
	id: string;
	platform: string;
	platformName: string;
	name: string;
	url: string;
	status: Status | 'pending';
	detail?: string;
	paid: boolean;
}

export type Filter = 'all' | Status;

// Order rows alphabetically by their display name, independent of status, so
// the list stays stable as results stream in. Array.prototype.sort is stable,
// so rows sharing a platformName keep their arrival order.
export function byPlatformName(a: Row, b: Row): number {
	if (a.platformName < b.platformName) return -1;
	if (a.platformName > b.platformName) return 1;
	return 0;
}

// Order rows paid-first, then alphabetically by display name within each group,
// so the report can split into a Paid section above a Free one.
export function byPaidThenPlatformName(a: Row, b: Row): number {
	if (a.paid !== b.paid) return a.paid ? -1 : 1;
	return byPlatformName(a, b);
}

function queryParam(): string {
	if (typeof location === 'undefined') return '';
	return new URLSearchParams(location.search).get('q') ?? '';
}

function loadQuery(): string {
	const fromUrl = queryParam();
	if (fromUrl) return fromUrl;
	if (typeof localStorage === 'undefined') return '';
	try {
		return localStorage.getItem(QUERY_KEY) ?? '';
	} catch {
		return '';
	}
}

function syncUrl(query: string) {
	const url = new URL(location.href);
	if (url.searchParams.get('q') === query) return;
	url.searchParams.set('q', query);
	history.pushState(null, '', url);
}

// Lookup owns the form state, the abort/generation guards that keep a stale
// stream from winning, and the query's URL plus localStorage persistence.
export class Lookup {
	input = $state(loadQuery());
	rows = $state<Row[]>([]);
	error = $state('');
	loading = $state(false);
	lastLabel = $state<string | null>(null);
	filter = $state<Filter>('all');
	copied = $state(false);

	#controller: AbortController | null = null;
	#generation = 0;

	check = ({ sync = true }: { sync?: boolean } = {}): void => {
		const label = toLabel(this.input);
		const gen = ++this.#generation;

		this.#controller?.abort();
		this.#controller = null;
		this.error = '';
		this.filter = 'all';

		if (!label) {
			this.rows = [];
			this.lastLabel = null;
			this.loading = false;
			const raw = this.input.trim();
			this.error = !raw
				? 'Enter a name to check.'
				: raw.includes('.')
					? 'Enter the name without the dot — e.g. “example”, not “example.com”.'
					: 'Names can use letters, numbers and hyphens, e.g. acema.';
			return;
		}

		try {
			localStorage.setItem(QUERY_KEY, this.input.trim());
		} catch {
			// storage unavailable: persistence is best-effort
		}
		if (sync) syncUrl(this.input.trim());

		const ac = new AbortController();
		this.#controller = ac;
		this.lastLabel = label;
		this.loading = true;
		this.rows = [];

		streamFromApi(
			label,
			(result) => {
				if (gen !== this.#generation || ac.signal.aborted) return;
				const next: Row = {
					id: result.id,
					platform: result.platform,
					platformName: result.platformName,
					name: result.name,
					url: result.url,
					status: result.status,
					detail: result.detail,
					paid: result.paid
				};
				this.rows = this.rows.some((row) => row.id === result.id)
					? this.rows.map((row) => (row.id === result.id ? next : row))
					: [...this.rows, next];
			},
			ac.signal
		)
			.catch((e) => {
				if (isAbort(e)) return;
				if (ac.signal.aborted || gen !== this.#generation) return;
				// The backend is the sole source, so a failed request leaves no
				// rows to mark unreachable; surface it instead.
				if (!this.rows.length) this.error = 'Could not reach the backend. Is `serve` running?';
				this.rows = this.rows.map((row) =>
					row.status === 'pending' ? { ...row, status: 'unknown', detail: 'unreachable' } : row
				);
			})
			.finally(() => {
				if (gen === this.#generation) this.loading = false;
			});
	};

	copyResults = async (): Promise<void> => {
		const lines = this.rows
			.toSorted(byPlatformName)
			.map((row) => `${row.platformName}\t${row.name}\t${row.status}`);
		const text = `${this.lastLabel ?? ''}\n${lines.join('\n')}`;
		try {
			await navigator.clipboard.writeText(text);
			this.copied = true;
			setTimeout(() => (this.copied = false), 1500);
		} catch {
			// clipboard unavailable
		}
	};

	searching = $derived(this.loading || this.rows.some((row) => row.status === 'pending'));
	checked = $derived(this.rows.filter((row) => row.status !== 'pending').length);

	stale = $derived(
		!this.searching &&
			this.rows.length > 0 &&
			this.lastLabel !== null &&
			toLabel(this.input) !== this.lastLabel
	);

	visibleRows = $derived(
		(this.filter === 'all'
			? this.rows
			: this.rows.filter((row) => row.status === this.filter)
		).toSorted(byPaidThenPlatformName)
	);

	paidRows = $derived(this.visibleRows.filter((row) => row.paid));
	freeRows = $derived(this.visibleRows.filter((row) => !row.paid));

	counts = $derived.by(() => {
		const totals = { available: 0, taken: 0, unknown: 0 };
		for (const row of this.rows) {
			if (row.status !== 'pending') totals[row.status]++;
		}
		return totals;
	});

	announcement = $derived(
		this.searching
			? ''
			: `${this.counts.available} available, ${this.counts.taken} unavailable, ${this.counts.unknown} unknown`
	);
}

export function useLookup(): Lookup {
	const lookup = new Lookup();

	onMount(() => {
		if (toLabel(lookup.input)) lookup.check({ sync: false });

		const onPop = () => {
			lookup.input = queryParam();
			lookup.check({ sync: false });
		};
		addEventListener('popstate', onPop);
		return () => removeEventListener('popstate', onPop);
	});

	return lookup;
}
