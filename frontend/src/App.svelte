<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import { onMount } from 'svelte';
	import { getBootstrap } from '$lib/bootstrap';
	import { providers, streamAll, toLabel, type ProviderKind, type Status } from '$lib/providers';
	import { Button } from '$lib/components/ui/button/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';

	getBootstrap().catch(() => {});

	const QUERY_KEY = 'domain-browser:query';

	interface Row {
		id: string;
		label: string;
		kind: ProviderKind;
		name: string;
		url: string;
		status: Status | 'pending';
		detail?: string;
	}

	const GROUP_LABELS: Record<ProviderKind, string> = {
		domain: 'Domains',
		dev: 'Packages',
		social: 'Social',
		org: 'Organizations',
		brand: 'Brands'
	};

	const GROUP_ORDER: ProviderKind[] = ['domain', 'dev', 'social', 'org', 'brand'];

	const STATUS: Record<
		Row['status'],
		{ text: string; cls: string; Icon: typeof CheckCircle2Icon; spin?: boolean }
	> = {
		available: {
			text: 'available',
			cls: 'text-success',
			Icon: CheckCircle2Icon
		},
		taken: { text: 'taken', cls: 'text-muted-foreground', Icon: XCircleIcon },
		unknown: { text: 'unknown', cls: 'text-muted-foreground', Icon: HelpCircleIcon },
		pending: { text: 'checking', cls: 'text-muted-foreground', Icon: LoaderCircleIcon, spin: true }
	};

	function loadQuery(): string {
		if (typeof localStorage === 'undefined') return '';
		try {
			return localStorage.getItem(QUERY_KEY) ?? '';
		} catch {
			return '';
		}
	}

	let input = $state(loadQuery());
	let rows = $state<Row[]>([]);
	let error = $state('');
	let lastLabel = $state<string | null>(null);
	let controller: AbortController | null = null;
	let generation = 0;

	function check() {
		const label = toLabel(input);
		const gen = ++generation;

		controller?.abort();
		controller = null;
		error = '';

		if (!label) {
			rows = [];
			lastLabel = null;
			error = input.trim()
				? 'Names can use letters, numbers and hyphens, e.g. acema.'
				: 'Enter a name to check.';
			return;
		}

		try {
			localStorage.setItem(QUERY_KEY, input.trim());
		} catch {
			// storage unavailable: persistence is best-effort
		}

		const ac = new AbortController();
		controller = ac;
		lastLabel = label;
		rows = providers.map((provider) => ({
			id: provider.id,
			label: provider.label,
			kind: provider.kind,
			name: provider.name(label),
			url: provider.url(label),
			status: 'pending'
		}));

		streamAll(
			label,
			(result) => {
				if (gen !== generation || ac.signal.aborted) return;
				rows = rows.map((row) =>
					row.id === result.id ? { ...row, status: result.status, detail: result.detail } : row
				);
			},
			ac.signal
		).catch((e) => {
			if (e instanceof DOMException && e.name === 'AbortError') return;
			if (ac.signal.aborted || gen !== generation) return;
			rows = rows.map((row) =>
				row.status === 'pending' ? { ...row, status: 'unknown', detail: 'unreachable' } : row
			);
		});
	}

	onMount(() => {
		if (toLabel(input)) check();
	});

	const searching = $derived(rows.some((row) => row.status === 'pending'));
	const checked = $derived(rows.filter((row) => row.status !== 'pending').length);

	const stale = $derived(
		!searching && rows.length > 0 && lastLabel !== null && toLabel(input) !== lastLabel
	);

	const groups = $derived.by(() =>
		GROUP_ORDER.map((kind) => ({
			kind,
			label: GROUP_LABELS[kind],
			rows: rows.filter((row) => row.kind === kind)
		})).filter((group) => group.rows.length > 0)
	);

	const counts = $derived.by(() => {
		const totals = { available: 0, taken: 0, unknown: 0 };
		for (const row of rows) {
			if (row.status !== 'pending') totals[row.status]++;
		}
		return totals;
	});
</script>

<ModeWatcher />

<div class="min-h-svh bg-background text-foreground">
	<div class="mx-auto max-w-xl px-4 py-16">
		<header class="mb-6 flex items-start justify-between gap-4">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">Domain Browser</h1>
				<p class="mt-1 text-sm text-muted-foreground">
					Check a name across domains, packages and GitHub handles.
				</p>
			</div>
			<Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
				<SunIcon class="size-4 dark:hidden" />
				<MoonIcon class="hidden size-4 dark:block" />
			</Button>
		</header>

		<LookupForm bind:value={input} loading={searching} onsubmit={check} />

		{#if stale}
			<p class="mt-3 text-xs text-muted-foreground">Press Enter to check the updated name.</p>
		{/if}

		<div aria-live="polite" aria-atomic="true">
			{#if error}
				<p class="mt-4 text-sm text-destructive">{error}</p>
			{/if}

			{#if rows.length}
				<p class="mt-4 text-sm text-muted-foreground">
					{#if searching}
						Checking {checked} of {rows.length}…
					{:else}
						{counts.available} of {rows.length} clear{counts.unknown
							? ` · ${counts.unknown} unknown`
							: ''}
					{/if}
				</p>

				<div class="mt-3 space-y-6">
					{#each groups as group (group.kind)}
						<section>
							<h2 class="mb-2 text-xs font-medium tracking-wide text-muted-foreground uppercase">
								{group.label}
							</h2>
							<ul class="divide-y overflow-hidden rounded-2xl border bg-card">
								{#each group.rows as row (row.id)}
									{@const status = STATUS[row.status]}
									{@const Icon = status.Icon}
									<li>
										<a
											href={row.url}
											target="_blank"
											rel="noopener noreferrer"
											class="flex items-center gap-3 px-3 py-2.5 transition-colors hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none focus-visible:ring-inset"
										>
											<Icon
												class="size-4 shrink-0 {status.cls} {status.spin ? 'animate-spin' : ''}"
											/>
											<div class="flex min-w-0 flex-1 flex-col">
												<span class="truncate font-mono text-sm">{row.name}</span>
												<span class="truncate text-xs text-muted-foreground">
													{row.label}{row.detail ? ` · ${row.detail}` : ''}
												</span>
											</div>
											<span class="shrink-0 text-xs {status.cls}">{status.text}</span>
										</a>
									</li>
								{/each}
							</ul>
						</section>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>
