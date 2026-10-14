<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import { onMount } from 'svelte';
	import { isAbort, streamFromApi } from '$lib/api';
	import { toLabel, type Status } from '$lib/providers';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as ToggleGroup from '$lib/components/ui/toggle-group/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';

	const QUERY_KEY = 'domain-browser:query';

	const EXAMPLES = ['acema', 'lumen', 'quartz', 'nimbus'];

	interface Row {
		id: string;
		platform: string;
		platformName: string;
		name: string;
		url: string;
		status: Status | 'pending';
		detail?: string;
	}

	const STATUS: Record<
		Row['status'],
		{ text: string; cls: string; Icon: typeof CheckCircle2Icon; spin?: boolean }
	> = {
		available: {
			text: 'available',
			cls: 'text-success',
			Icon: CheckCircle2Icon
		},
		taken: { text: 'unavailable', cls: 'text-muted-foreground', Icon: XCircleIcon },
		unknown: { text: 'unknown', cls: 'text-warning', Icon: HelpCircleIcon },
		pending: { text: 'checking', cls: 'text-muted-foreground', Icon: LoaderCircleIcon, spin: true }
	};

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

	let input = $state(loadQuery());
	let rows = $state<Row[]>([]);
	let error = $state('');
	let loading = $state(false);
	let lastLabel = $state<string | null>(null);
	let controller: AbortController | null = null;
	let generation = 0;
	let filter = $state<'all' | Status>('all');
	let copied = $state(false);

	function check({ sync = true }: { sync?: boolean } = {}) {
		const label = toLabel(input);
		const gen = ++generation;

		controller?.abort();
		controller = null;
		error = '';
		filter = 'all';

		if (!label) {
			rows = [];
			lastLabel = null;
			loading = false;
			const raw = input.trim();
			error = !raw
				? 'Enter a name to check.'
				: raw.includes('.')
					? 'Enter the name without the dot — e.g. “example”, not “example.com”.'
					: 'Names can use letters, numbers and hyphens, e.g. acema.';
			return;
		}

		try {
			localStorage.setItem(QUERY_KEY, input.trim());
		} catch {
			// storage unavailable: persistence is best-effort
		}
		if (sync) syncUrl(input.trim());

		const ac = new AbortController();
		controller = ac;
		lastLabel = label;
		loading = true;
		rows = [];

		streamFromApi(
			label,
			(result) => {
				if (gen !== generation || ac.signal.aborted) return;
				const next: Row = {
					id: result.id,
					platform: result.platform,
					platformName: result.platformName,
					name: result.name,
					url: result.url,
					status: result.status,
					detail: result.detail
				};
				rows = rows.some((row) => row.id === result.id)
					? rows.map((row) => (row.id === result.id ? next : row))
					: [...rows, next];
			},
			ac.signal
		)
			.catch((e) => {
				if (isAbort(e)) return;
				if (ac.signal.aborted || gen !== generation) return;
				// The backend is the sole source, so a failed request leaves no
				// rows to mark unreachable; surface it instead.
				if (!rows.length) error = 'Could not reach the backend. Is `serve` running?';
				rows = rows.map((row) =>
					row.status === 'pending' ? { ...row, status: 'unknown', detail: 'unreachable' } : row
				);
			})
			.finally(() => {
				if (gen === generation) loading = false;
			});
	}

	function onGlobalKeydown(e: KeyboardEvent) {
		if (e.key !== '/' || e.metaKey || e.ctrlKey || e.altKey) return;
		const target = e.target as HTMLElement | null;
		if (target && (target.isContentEditable || /^(input|textarea|select)$/i.test(target.tagName)))
			return;
		e.preventDefault();
		document.getElementById('domain')?.focus();
	}

	onMount(() => {
		if (toLabel(input)) check({ sync: false });

		const onPop = () => {
			input = queryParam();
			check({ sync: false });
		};
		addEventListener('popstate', onPop);
		addEventListener('keydown', onGlobalKeydown);
		return () => {
			removeEventListener('popstate', onPop);
			removeEventListener('keydown', onGlobalKeydown);
		};
	});

	const searching = $derived(loading || rows.some((row) => row.status === 'pending'));
	const checked = $derived(rows.filter((row) => row.status !== 'pending').length);

	const stale = $derived(
		!searching && rows.length > 0 && lastLabel !== null && toLabel(input) !== lastLabel
	);

	const visibleRows = $derived(
		filter === 'all' ? rows : rows.filter((row) => row.status === filter)
	);

	const counts = $derived.by(() => {
		const totals = { available: 0, taken: 0, unknown: 0 };
		for (const row of rows) {
			if (row.status !== 'pending') totals[row.status]++;
		}
		return totals;
	});

	const announcement = $derived(
		searching
			? ''
			: `${counts.available} available, ${counts.taken} unavailable, ${counts.unknown} unknown`
	);

	async function copyResults() {
		const lines = rows.map((row) => `${row.platformName}\t${row.name}\t${row.status}`);
		const text = `${lastLabel ?? ''}\n${lines.join('\n')}`;
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			setTimeout(() => (copied = false), 1500);
		} catch {
			// clipboard unavailable
		}
	}
</script>

<ModeWatcher />

<div class="min-h-svh bg-background text-foreground">
	<a
		href="#main"
		class="sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-4 focus:z-30 focus:rounded-md focus:border focus:border-border focus:bg-card focus:px-3 focus:py-1.5 focus:text-sm"
		>Skip to results</a
	>
	<header
		class="sticky top-0 z-20 border-b border-border/60 bg-background/85 backdrop-blur [@media(max-height:34rem)]:static"
	>
		<div class="mx-auto max-w-2xl px-4 py-2.5 lg:max-w-5xl">
			<div class="flex items-center gap-3">
				<h1 class="shrink-0 text-sm font-semibold tracking-tight sm:text-base">Domain Browser</h1>
				<div class="flex min-w-0 flex-1 justify-center">
					<div class="w-full max-w-xl">
						<LookupForm bind:value={input} loading={searching} onsubmit={check} />
					</div>
				</div>
				<Button
					variant="ghost"
					size="icon"
					class="shrink-0"
					onclick={toggleMode}
					aria-label="Toggle theme"
				>
					<SunIcon class="size-4 dark:hidden" />
					<MoonIcon class="hidden size-4 dark:block" />
				</Button>
			</div>

			{#if rows.length}
				<div class="mt-2 flex flex-wrap items-center gap-2">
					{#if searching}
						<span class="text-xs text-muted-foreground">{lastLabel}</span>
					{:else}
						<ToggleGroup.Root
							type="single"
							value={filter}
							onValueChange={(value) => (filter = (value || 'all') as 'all' | Status)}
							variant="outline"
							size="sm"
							spacing={2}
							class="text-xs"
						>
							<ToggleGroup.Item
								value="available"
								class="data-[state=on]:border-success/40 data-[state=on]:bg-success/15 data-[state=on]:text-success"
							>
								{counts.available} available
							</ToggleGroup.Item>
							<ToggleGroup.Item value="taken" class="data-[state=on]:text-foreground">
								{counts.taken} unavailable
							</ToggleGroup.Item>
							{#if counts.unknown}
								<ToggleGroup.Item
									value="unknown"
									class="data-[state=on]:border-warning/40 data-[state=on]:bg-warning/15 data-[state=on]:text-warning"
								>
									{counts.unknown} unknown
								</ToggleGroup.Item>
							{/if}
						</ToggleGroup.Root>
						<Button variant="outline" size="sm" class="ml-auto" onclick={copyResults}>
							{copied ? 'Copied' : 'Copy'}
						</Button>
					{/if}
				</div>
			{/if}
		</div>
	</header>

	<div id="main" class="mx-auto max-w-2xl px-4 py-5 lg:max-w-5xl">
		<p class="max-w-prose text-sm text-muted-foreground">
			Best-effort from public sources. <em>Unavailable</em> means the name is taken or the platform
			blocks it (too short, reserved, suspended); <em>available</em> means none of that was found — not
			a guarantee you can register it.
		</p>

		{#if stale}
			<p class="mt-3 text-xs text-muted-foreground">Press Enter to check the updated name.</p>
		{/if}

		<span class="sr-only" role="status" aria-live="polite">{announcement}</span>

		{#if error}
			<p class="mt-4 text-sm text-destructive">{error}</p>
		{/if}

		{#if rows.length}
			{#if searching}
				<p class="mt-4 text-sm text-muted-foreground">Checking {checked} of {rows.length}…</p>
			{:else}
				<p class="mt-4 text-lg font-semibold tracking-tight">
					{counts.available} of {rows.length} available
				</p>
			{/if}

			<Card.Root class="mt-3 gap-0 py-0">
				<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
					{#each visibleRows as row (row.id)}
						{@const pending = row.status === 'pending'}
						{@const status = STATUS[row.status]}
						{@const Icon = status.Icon}
						{@const text = row.url.replace(/^https?:\/\//, '')}
						<li class="bg-card">
							<a
								href={row.url}
								target="_blank"
								rel="noopener noreferrer"
								class="flex items-center gap-3 px-3 py-3 {row.status === 'available'
									? 'bg-success/[0.07] hover:bg-success/[0.12]'
									: 'hover:bg-muted/50'} focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none focus-visible:ring-inset"
							>
								<span
									class="hidden w-28 shrink-0 truncate text-xs font-medium text-muted-foreground sm:block"
									title={row.platformName}>{row.platformName}</span
								>
								<span
									class="min-w-0 flex-1 truncate text-sm {row.status === 'taken'
										? 'text-muted-foreground'
										: ''}">{text}</span
								>
								{#if row.detail}
									<span class="hidden shrink-0 text-xs text-muted-foreground sm:inline"
										>{row.detail}</span
									>
								{/if}
								{#if pending}
									<Skeleton class="h-3 w-14 shrink-0 rounded-full motion-reduce:animate-none" />
								{:else}
									<span class="flex shrink-0 items-center gap-1.5 text-xs {status.cls}">
										<Icon class="size-3.5 shrink-0" />
										<span class="w-20">{status.text}</span>
									</span>
								{/if}
							</a>
						</li>
					{/each}
				</ul>
			</Card.Root>
		{:else if !error}
			<div class="mt-6">
				<p class="text-sm text-muted-foreground">Try a name:</p>
				<div class="mt-2 flex flex-wrap gap-2">
					{#each EXAMPLES as name (name)}
						<Button
							variant="outline"
							size="sm"
							onclick={() => {
								input = name;
								check();
							}}
						>
							{name}
						</Button>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>
