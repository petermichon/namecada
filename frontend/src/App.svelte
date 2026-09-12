<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import NAMES_RAW from '$lib/domains-names.txt?raw';
	import AVAIL_RAW from '$lib/domains-avail.txt?raw';
	import { parseDataset, search, type Mode } from '$lib/search';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import ResultList from '$lib/components/ResultList.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	const dataset = parseDataset(NAMES_RAW, AVAIL_RAW);
	const total = dataset.names.length;

	let rawQuery = $state('');
	let query = $state('');
	let mode = $state<Mode>('prefix');
	let onlyAvailable = $state(true);

	$effect(() => {
		const v = rawQuery;
		const t = setTimeout(() => (query = v), 120);
		return () => clearTimeout(t);
	});

	const result = $derived(search(dataset, { query, mode, onlyAvailable }));
</script>

<ModeWatcher />

<div class="min-h-svh bg-background text-foreground">
	<div class="mx-auto max-w-3xl px-4 py-10">
		<header class="mb-6 flex items-start justify-between gap-4">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">5-letter .com browser</h1>
				<p class="mt-1 text-sm text-muted-foreground">
					{total.toLocaleString()} names scanned · {dataset.availableTotal.toLocaleString()} available
				</p>
			</div>
			<Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
				<SunIcon class="size-4 dark:hidden" />
				<MoonIcon class="hidden size-4 dark:block" />
			</Button>
		</header>

		<SearchBar bind:query={rawQuery} bind:mode bind:onlyAvailable />

		{#if query.trim().length === 0}
			<p class="mt-10 text-center text-sm text-muted-foreground">
				Start typing to search {total.toLocaleString()} names.
			</p>
		{:else}
			<div class="mt-4 flex items-center justify-between text-xs text-muted-foreground">
				<span>{result.total.toLocaleString()} match{result.total === 1 ? '' : 'es'}</span>
				{#if result.truncated}<span>showing first {result.rows.length}</span>{/if}
			</div>
			<div class="mt-2">
				<ResultList rows={result.rows} />
			</div>
			{#if result.rows.length === 0}
				<p class="mt-10 text-center text-sm text-muted-foreground">No matches.</p>
			{/if}
		{/if}
	</div>
</div>
