import { onMount } from 'svelte';
import { appCheckClient, isAbort, type CheckClient, type CheckEvent } from './api';
import { toLabel, suggestLabel, type CostFilter, type Status } from './platforms';

export interface Row {
	id: string;
	platform: string;
	platformName: string;
	name: string;
	url: string;
	status: Status | 'pending' | 'unsupported';
	detail?: string;
	paid: boolean;
}

export type FilterStatus = Status | 'unsupported';
// A filter is the set of statuses to keep. An empty set means no selection,
// which shows every row (the default), so `available + not checked` unions
// those two and hides the rest.
export type Filter = FilterStatus[];

// The default view is the verdicts; deferred/keyed sources are opt-in.
export const DEFAULT_FILTER: Filter = ['available', 'taken', 'unknown'];
export const ALL_STATUSES: Filter = [...DEFAULT_FILTER, 'unsupported'];
export const DEFAULT_COST: CostFilter = ['free', 'paid'];

// Order rows alphabetically by their display name, independent of status, so
// the list stays stable as results stream in. Ties break on the provider URL so
// the order is deterministic regardless of arrival order — domains sort by TLD
// (.ai .app .com .dev .io .org) and Hex sorts packages/ before users/.
export function byPlatformName(a: Row, b: Row): number {
	if (a.platformName < b.platformName) return -1;
	if (a.platformName > b.platformName) return 1;
	if (a.url < b.url) return -1;
	if (a.url > b.url) return 1;
	return 0;
}

function queryParam(): string {
	if (typeof location === 'undefined') return '';
	return new URLSearchParams(location.search).get('q') ?? '';
}

function matchesStatus(row: Row, filter: Filter): boolean {
	return (
		filter.length > 0 &&
		(filter.includes(row.status as FilterStatus) ||
			(row.status === 'pending' && filter.some((status) => status !== 'unsupported')))
	);
}

function matchesCost(row: Row, cost: CostFilter): boolean {
	return cost.includes(row.paid ? 'paid' : 'free');
}

function settlePending(rows: Row[]): Row[] {
	return rows.map((row) =>
		row.status === 'pending' ? { ...row, status: 'unknown' as const, detail: 'unreachable' } : row
	);
}

function syncUrl(query: string) {
	const url = new URL(location.href);
	if (url.searchParams.get('q') === query) return;
	url.searchParams.set('q', query);
	history.pushState(null, '', url);
}

// Lookup owns the form state, the abort/generation guards that keep a stale
// stream from winning, and the query's URL persistence.
export class Lookup {
	input = $state(queryParam());
	rows = $state<Row[]>([]);
	error = $state('');
	suggestion = $state<string | null>(null);
	loading = $state(false);
	// Seed from the URL so a shared ?q= link renders the results page straight
	// away, without a flash of the landing page.
	lastLabel = $state<string | null>(toLabel(queryParam()));
	filter = $state<Filter>([...DEFAULT_FILTER]);
	cost = $state<CostFilter>([...DEFAULT_COST]);

	#controller: AbortController | null = null;
	#generation = 0;
	#client: CheckClient;

	constructor(client: CheckClient = appCheckClient) {
		this.#client = client;
	}

	check = ({ sync = true }: { sync?: boolean } = {}): void => {
		const label = toLabel(this.input);
		const gen = ++this.#generation;

		this.#controller?.abort();
		this.#controller = null;
		this.error = '';
		this.suggestion = null;
		this.filter = [...DEFAULT_FILTER];
		this.cost = [...DEFAULT_COST];

		if (!label) {
			this.rows = [];
			this.lastLabel = null;
			this.loading = false;
			const raw = this.input.trim();
			this.error = !raw
				? 'Enter a name to check.'
				: raw.includes('.')
					? 'Enter the name without the dot (e.g. “example”, not “example.com”).'
					: 'Names can use letters, numbers and hyphens, e.g. acema.';
			this.suggestion = suggestLabel(raw);
			return;
		}

		if (sync) syncUrl(this.input.trim());

		const ac = new AbortController();
		this.#controller = ac;
		this.lastLabel = label;
		this.loading = true;
		// The manifest arrives only after a round trip. Keep the previous row
		// slots mounted as pending in the meantime so the document keeps its
		// height and the browser cannot clamp the scroll position to the top.
		// Deferred rows stay deferred and are hidden by the default filter.
		if (this.rows.length) {
			this.rows = this.rows.map((row) =>
				row.status === 'unsupported'
					? row
					: {
							...row,
							name: '',
							url: '',
							status: 'pending' as const,
							detail: undefined
						}
			);
		}

		let manifestReceived = false;

		this.#client
			.stream(
				this.input.trim(),
				(event: CheckEvent) => {
					if (gen !== this.#generation || ac.signal.aborted) return;
					if (event.type === 'manifest') {
						manifestReceived = true;
						this.rows = event.sources.map((source) => ({
							id: source.id,
							platform: source.id,
							platformName: source.platformName,
							name: event.label || label,
							url: source.url,
							status: source.checked ? ('pending' as const) : ('unsupported' as const),
							paid: source.paid
						}));
					} else if (event.type === 'done') {
						this.rows = settlePending(this.rows);
					} else if (event.type === 'result') {
						const next: Row = {
							id: event.id,
							platform: event.platform,
							platformName: event.platformName,
							name: event.name,
							url: event.url,
							status: event.status,
							detail: event.detail,
							paid: event.paid
						};
						this.rows = this.rows.some((row) => row.id === event.id)
							? this.rows.map((row) => (row.id === event.id ? next : row))
							: [...this.rows, next];
					}
				},
				ac.signal
			)
			.then(() => {
				if (ac.signal.aborted || gen !== this.#generation) return;
				// A clean EOF before every result still must not leave rows pending.
				this.rows = settlePending(this.rows);
			})
			.catch((e) => {
				if (isAbort(e)) return;
				if (ac.signal.aborted || gen !== this.#generation) return;
				if (!manifestReceived) {
					// The retained slots belong to the previous scan; do not show
					// them as unknown results for this one.
					this.rows = [];
					this.error = 'Could not reach the backend. Is `serve` running?';
					return;
				}
				// A scan that fails after the manifest still knows every source;
				// mark the ones that never resolved so they are not stuck pending.
				this.rows = settlePending(this.rows);
			})
			.finally(() => {
				if (gen === this.#generation) this.loading = false;
			});
	};

	// clearError drops validation feedback as soon as the user starts editing;
	// the next submit re-validates.
	clearError = (): void => {
		this.error = '';
		this.suggestion = null;
	};

	// acceptSuggestion applies the proposed bare label instead of silently
	// rewriting the user's value.
	acceptSuggestion = ({ sync = true }: { sync?: boolean } = {}): void => {
		if (this.suggestion === null) return;
		this.input = this.suggestion;
		this.suggestion = null;
		this.check({ sync });
	};

	// showAll clears every facet and restores the default verdict filter with
	// deferred sources included.
	showAll = (): void => {
		this.filter = [...ALL_STATUSES];
		this.cost = [...DEFAULT_COST];
	};

	// goHome returns to the landing state without a reload: it abandons any
	// in-flight scan (so a late result cannot repopulate the page) and drops
	// the results, while pushing "/" so Back still returns to the report.
	goHome = (): void => {
		this.#controller?.abort();
		this.#controller = null;
		this.#generation += 1;
		this.input = '';
		this.rows = [];
		this.error = '';
		this.suggestion = null;
		this.loading = false;
		this.filter = [...DEFAULT_FILTER];
		this.cost = [...DEFAULT_COST];
		this.lastLabel = null;
		if (typeof location !== 'undefined' && location.search) {
			history.pushState(null, '', location.pathname);
		}
	};

	searching = $derived(this.loading || this.rows.some((row) => row.status === 'pending'));

	// Selection is what is shown. Pending rows stay visible while a verdict
	// status is selected so the list does not empty and refill mid-scan.
	visibleRows = $derived(
		(this.filter.length === 0
			? []
			: this.rows.filter((row) => matchesStatus(row, this.filter) && matchesCost(row, this.cost))
		).toSorted(byPlatformName)
	);

	counts = $derived.by(() => {
		const totals = { available: 0, taken: 0, unknown: 0, unsupported: 0 };
		for (const row of this.rows) {
			if (!matchesCost(row, this.cost)) continue;
			if (row.status === 'available' || row.status === 'taken' || row.status === 'unknown') {
				totals[row.status]++;
			} else if (row.status === 'unsupported') {
				totals.unsupported++;
			}
		}
		return totals;
	});

	statusTotal = $derived(
		this.counts.available + this.counts.taken + this.counts.unknown + this.counts.unsupported
	);

	visibleCounts = $derived.by(() => {
		const totals = { available: 0, taken: 0, unknown: 0, unsupported: 0, pending: 0 };
		for (const row of this.visibleRows) {
			if (row.status === 'pending') {
				totals.pending++;
			} else {
				totals[row.status]++;
			}
		}
		return totals;
	});

	// The cost chips honour the status filter, so a chip previews how many rows
	// selecting it would show.
	costCounts = $derived.by(() => {
		let free = 0;
		let paid = 0;
		for (const row of this.rows) {
			if (!matchesStatus(row, this.filter)) continue;
			if (row.paid) paid++;
			else free++;
		}
		return { free, paid };
	});

	announcement = $derived.by(() => {
		if (this.searching) return '';
		if (this.filter.length === 0) return 'No statuses selected.';
		const summary = `${this.visibleCounts.available} available, ${this.visibleCounts.taken} unavailable, ${this.visibleCounts.unknown} unknown`;
		const notChecked = this.visibleCounts.unsupported
			? `, ${this.visibleCounts.unsupported} not checked`
			: '';
		return `${this.visibleRows.length} shown. ${summary}${notChecked}`;
	});
}

export function useLookup(client: CheckClient = appCheckClient): Lookup {
	const lookup = new Lookup(client);

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
