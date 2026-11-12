import { onMount } from 'svelte';
import { appCheckClient, isAbort, type CheckClient, type CheckEvent } from './api';
import { toLabel, type CostFilter } from './platforms';
import {
	announce,
	costTally,
	DEFAULT_COST,
	DEFAULT_FILTER,
	ALL_STATUSES,
	matchesCost,
	settlePending,
	tally,
	validate,
	visible,
	type Filter,
	type Row
} from './rows';

function queryParam(): string {
	if (typeof location === 'undefined') return '';
	return new URLSearchParams(location.search).get('q') ?? '';
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

	// reset abandons any in-flight scan and clears the per-scan feedback.
	#reset(): void {
		this.#controller?.abort();
		this.#controller = null;
		this.error = '';
		this.suggestion = null;
		this.filter = [...DEFAULT_FILTER];
		this.cost = [...DEFAULT_COST];
	}

	// holdRows keeps the previous row slots mounted as pending until the
	// manifest arrives, so the document keeps its height and the browser cannot
	// clamp the scroll position to the top. Deferred rows stay deferred and are
	// hidden by the default filter.
	#holdRows(): void {
		if (!this.rows.length) return;
		this.rows = this.rows.map((row) =>
			row.status === 'unsupported'
				? row
				: { ...row, name: '', url: '', status: 'pending' as const, detail: undefined }
		);
	}

	check = ({ sync = true }: { sync?: boolean } = {}): void => {
		const { label, error, suggestion } = validate(this.input);
		const generation = ++this.#generation;
		this.#reset();

		if (!label) {
			this.rows = [];
			this.lastLabel = null;
			this.loading = false;
			this.error = error;
			this.suggestion = suggestion;
			return;
		}

		if (sync) syncUrl(this.input.trim());

		this.#controller = new AbortController();
		this.lastLabel = label;
		this.loading = true;
		this.#holdRows();
		this.#stream(label, generation);
	};

	// stream runs one scan. The generation guard stops a stale stream from
	// writing rows once a newer check has started.
	#stream(label: string, generation: number): void {
		const ac = this.#controller;
		if (!ac) return;
		let manifestReceived = false;

		this.#client
			.stream(
				this.input.trim(),
				(event) => {
					if (generation !== this.#generation || ac.signal.aborted) return;
					if (event.type === 'manifest') manifestReceived = true;
					this.#apply(event, label);
				},
				ac.signal
			)
			.then(() => {
				if (ac.signal.aborted || generation !== this.#generation) return;
				// A clean EOF before every result still must not leave rows pending.
				this.rows = settlePending(this.rows);
			})
			.catch((e) => {
				if (isAbort(e)) return;
				if (ac.signal.aborted || generation !== this.#generation) return;
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
				if (generation === this.#generation) this.loading = false;
			});
	}

	#apply(event: CheckEvent, label: string): void {
		if (event.type === 'manifest') {
			this.rows = event.sources.map((source) => ({
				id: source.id,
				platform: source.id,
				platformName: source.platformName,
				name: event.label || label,
				url: source.url,
				status: source.checked ? ('pending' as const) : ('unsupported' as const),
				paid: source.paid
			}));
			return;
		}
		if (event.type === 'done') {
			this.rows = settlePending(this.rows);
			return;
		}
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
		this.#reset();
		this.#generation += 1;
		this.input = '';
		this.rows = [];
		this.loading = false;
		this.lastLabel = null;
		if (typeof location !== 'undefined' && location.search) {
			history.pushState(null, '', location.pathname);
		}
	};

	searching = $derived(this.loading || this.rows.some((row) => row.status === 'pending'));

	// Selection is what is shown. Pending rows stay visible while a verdict
	// status is selected so the list does not empty and refill mid-scan.
	visibleRows = $derived(visible(this.rows, this.filter, this.cost));

	counts = $derived(tally(this.rows.filter((row) => matchesCost(row, this.cost))));

	statusTotal = $derived(
		this.counts.available + this.counts.taken + this.counts.unknown + this.counts.unsupported
	);

	visibleCounts = $derived(tally(this.visibleRows));

	// Sources that were actually checked, on screen and overall.
	visibleChecked = $derived(
		this.visibleCounts.available + this.visibleCounts.taken + this.visibleCounts.unknown
	);

	checkedCount = $derived(this.rows.filter((row) => row.status !== 'unsupported').length);

	// The cost chips honour the status filter, so a chip previews how many rows
	// selecting it would show.
	costCounts = $derived(costTally(this.rows, this.filter));

	announcement = $derived(announce(this.visibleRows, this.filter, this.searching));
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
