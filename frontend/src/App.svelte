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
	import {
		PLATFORMS,
		providers,
		streamAll,
		toLabel,
		type PlatformId,
		type Status
	} from '$lib/providers';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as ToggleGroup from '$lib/components/ui/toggle-group/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';

	getBootstrap().catch(() => {});

	const QUERY_KEY = 'domain-browser:query';

	const EXAMPLES = ['acema', 'lumen', 'quartz', 'nimbus'];

	interface Row {
		id: string;
		platform: PlatformId;
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
		if (sync) syncUrl(input.trim());

		const ac = new AbortController();
		controller = ac;
		lastLabel = label;
		rows = providers.map((provider) => ({
			id: provider.id,
			platform: provider.platform,
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

	const searching = $derived(rows.some((row) => row.status === 'pending'));
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
		const lines = rows.map((row) => `${PLATFORMS[row.platform]}\t${row.name}\t${row.status}`);
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
	<header class="sticky top-0 z-20 border-b border-border/60 bg-background/85 backdrop-blur">
		<div class="mx-auto flex max-w-xl items-center gap-3 px-4 py-2.5">
			<h1 class="shrink-0 text-sm font-semibold tracking-tight sm:text-base">Domain Browser</h1>
			<div class="min-w-0 flex-1">
				<LookupForm bind:value={input} loading={searching} onsubmit={check} />
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
	</header>

	<div id="main" class="mx-auto max-w-xl scroll-mt-16 px-4 py-5">
		<p class="text-sm text-muted-foreground">
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

			<div
				class="sticky top-[60px] z-10 -mx-4 mt-2 flex flex-wrap items-center gap-2 bg-background/90 px-4 py-2 backdrop-blur"
			>
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
					<button
						type="button"
						class="ml-auto rounded-md border border-border px-2.5 py-0.5 text-xs text-muted-foreground transition-colors hover:bg-muted/50"
						onclick={copyResults}
					>
						{copied ? 'Copied' : 'Copy'}
					</button>
				{/if}
			</div>

			<div class="mt-1 overflow-hidden rounded-2xl border bg-card">
				<ul class="divide-y">
					{#each visibleRows as row (row.id)}
						{@const pending = row.status === 'pending'}
						{@const status = STATUS[row.status]}
						{@const Icon = status.Icon}
						{@const text = row.url.replace(/^https?:\/\//, '')}
						<li>
							<a
								href={row.url}
								target="_blank"
								rel="noopener noreferrer"
								class="flex items-center gap-3 px-3 py-2.5 {row.status === 'available'
									? 'bg-success/10 hover:bg-success/15'
									: 'hover:bg-muted/50'} focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none focus-visible:ring-inset"
							>
								{#if pending}
									<span
										class="size-4 shrink-0 animate-pulse rounded-full bg-muted motion-reduce:animate-none"
									></span>
								{:else}
									<Icon class="size-4 shrink-0 {status.cls}" />
								{/if}
								<span
									class="hidden w-28 shrink-0 truncate text-xs font-medium text-muted-foreground sm:block"
									title={PLATFORMS[row.platform]}>{PLATFORMS[row.platform]}</span
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
									<span
										class="h-3 w-14 shrink-0 animate-pulse rounded-full bg-muted motion-reduce:animate-none"
									></span>
								{:else}
									<span class="shrink-0 text-xs {status.cls}">{status.text}</span>
								{/if}
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{:else if !error}
			<div class="mt-6">
				<p class="text-sm text-muted-foreground">Try a name:</p>
				<div class="mt-2 flex flex-wrap gap-2">
					{#each EXAMPLES as name (name)}
						<button
							type="button"
							class="rounded-full border border-border px-3 py-1 text-sm transition-colors hover:bg-muted/50"
							onclick={() => {
								input = name;
								check();
							}}
						>
							{name}
						</button>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>
