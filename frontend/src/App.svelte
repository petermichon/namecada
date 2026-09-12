<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import ShuffleIcon from '@lucide/svelte/icons/shuffle';
	import StarIcon from '@lucide/svelte/icons/star';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import NAMES_RAW from '$lib/domains-names.txt?raw';
	import AVAIL_RAW from '$lib/domains-avail.txt?raw';
	import {
		allIndices,
		availableIndices,
		indexOfName,
		matchingIndices,
		parseDataset,
		sampleFromPool
	} from '$lib/search';
	import { stars } from '$lib/stars.svelte';
	import { settings } from '$lib/settings.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import LetterFilter from '$lib/components/LetterFilter.svelte';
	import ResultList from '$lib/components/ResultList.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	const SAMPLE_SIZE = 5;

	const dataset = parseDataset(NAMES_RAW, AVAIL_RAW);
	const availablePool = availableIndices(dataset);
	const allPool = allIndices(dataset);
	const total = dataset.names.length;

	let query = $state(settings.query);
	let randomIndices = $state<number[]>([]);
	let randomPoolSize = $state(0);
	let randomFiltered = $state(false);
	let showStarred = $state(false);
	let copiedAll = $state(false);
	let lastMode = settings.mode;
	let lastOnlyAvailable = settings.onlyAvailable;
	let lastLetters = settings.letters.join('');
	let lastLetterMode = settings.letterMode;

	$effect(() => {
		const v = settings.query;
		const t = setTimeout(() => {
			query = v;
			if (v.trim().length > 0) {
				randomIndices = [];
				showStarred = false;
			}
		}, 120);
		return () => clearTimeout(t);
	});

	$effect(() => {
		const mode = settings.mode;
		const onlyAvailable = settings.onlyAvailable;
		const letters = settings.letters.join('');
		const letterMode = settings.letterMode;
		const changed =
			mode !== lastMode ||
			onlyAvailable !== lastOnlyAvailable ||
			letters !== lastLetters ||
			letterMode !== lastLetterMode;
		lastMode = mode;
		lastOnlyAvailable = onlyAvailable;
		lastLetters = letters;
		lastLetterMode = letterMode;
		if (changed && randomIndices.length > 0) showRandom();
	});

	const hasFilter = $derived(query.trim().length > 0 || settings.letters.length > 0);
	const viewPool = $derived(
		hasFilter
			? matchingIndices(dataset, {
					query,
					mode: settings.mode,
					onlyAvailable: settings.onlyAvailable,
					letters: settings.letters,
					letterMode: settings.letterMode
				})
			: settings.onlyAvailable
				? availablePool
				: allPool
	);
	const viewTotal = $derived(viewPool.length);
	const starredIndices = $derived(
		stars.items.map((n) => indexOfName(dataset, n)).filter((i) => i >= 0)
	);

	function showRandom() {
		showStarred = false;
		const filtered = settings.query.trim().length > 0 || settings.letters.length > 0;
		if (filtered) {
			const pool = matchingIndices(dataset, {
				query: settings.query,
				mode: settings.mode,
				onlyAvailable: settings.onlyAvailable,
				letters: settings.letters,
				letterMode: settings.letterMode
			});
			randomPoolSize = pool.length;
			randomFiltered = true;
			randomIndices = sampleFromPool(pool, SAMPLE_SIZE);
		} else {
			const pool = settings.onlyAvailable ? availablePool : allPool;
			randomPoolSize = pool.length;
			randomFiltered = false;
			randomIndices = sampleFromPool(pool, SAMPLE_SIZE);
		}
	}

	function clearRandom() {
		randomIndices = [];
	}

	function toggleStarredView() {
		showStarred = !showStarred;
		if (showStarred) {
			settings.query = '';
			query = '';
			randomIndices = [];
		}
	}

	async function copyAllStarred() {
		if (stars.items.length === 0) return;
		try {
			await navigator.clipboard.writeText(stars.items.map((n) => `${n}.com`).join('\n'));
			copiedAll = true;
			setTimeout(() => (copiedAll = false), 1500);
		} catch {
			copiedAll = false;
		}
	}
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
			<div class="flex items-center gap-1">
				<Button
					variant={showStarred ? 'secondary' : 'ghost'}
					size="sm"
					onclick={toggleStarredView}
					aria-pressed={showStarred}
					title="Show starred domains"
				>
					<StarIcon class="size-4" fill={showStarred ? 'currentColor' : 'none'} />
					Starred ({stars.items.length})
				</Button>
				<Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
					<SunIcon class="size-4 dark:hidden" />
					<MoonIcon class="hidden size-4 dark:block" />
				</Button>
			</div>
		</header>

		<SearchBar
			bind:query={settings.query}
			bind:mode={settings.mode}
			bind:onlyAvailable={settings.onlyAvailable}
			onrandom={showRandom}
		/>
		<LetterFilter bind:letters={settings.letters} bind:mode={settings.letterMode} />

		{#if showStarred}
			<div class="mt-4 flex items-center justify-between text-xs text-muted-foreground">
				<span>{starredIndices.length} starred</span>
				{#if starredIndices.length > 0}
					<Button variant="ghost" size="sm" onclick={copyAllStarred}>
						{#if copiedAll}
							<CheckIcon class="size-4" />
							Copied
						{:else}
							<CopyIcon class="size-4" />
							Copy all
						{/if}
					</Button>
				{/if}
			</div>
			<div class="mt-2">
				<ResultList
					indices={starredIndices}
					{dataset}
					starred={stars.items}
					ontogglestar={(n) => stars.toggle(n)}
				/>
			</div>
			{#if starredIndices.length === 0}
				<p class="mt-10 text-center text-sm text-muted-foreground">
					No starred domains yet. Click the star on any result to save it.
				</p>
			{/if}
		{:else if randomIndices.length > 0}
			<div class="mt-4 flex items-center justify-between text-xs text-muted-foreground">
				<span>
					{#if randomFiltered}
						{randomIndices.length} of {randomPoolSize.toLocaleString()} match{randomPoolSize === 1
							? ''
							: 'es'}
					{:else}
						{randomIndices.length} random{settings.onlyAvailable ? ' available' : ''}
					{/if}
				</span>
				<div class="flex items-center gap-1">
					<Button variant="ghost" size="sm" onclick={showRandom}>
						<ShuffleIcon class="size-4" />
						Shuffle
					</Button>
					<Button variant="ghost" size="sm" onclick={clearRandom}>Clear</Button>
				</div>
			</div>
			<div class="mt-2">
				<ResultList
					indices={randomIndices}
					{dataset}
					starred={stars.items}
					ontogglestar={(n) => stars.toggle(n)}
				/>
			</div>
		{:else}
			<div class="mt-4 flex items-center justify-between text-xs text-muted-foreground">
				<span>
					{#if hasFilter}
						{viewTotal.toLocaleString()} match{viewTotal === 1 ? '' : 'es'}
					{:else}
						{viewTotal.toLocaleString()}{settings.onlyAvailable ? ' available' : ' names'}
					{/if}
				</span>
			</div>
			<div class="mt-2">
				<ResultList
					indices={viewPool}
					{dataset}
					starred={stars.items}
					ontogglestar={(n) => stars.toggle(n)}
				/>
			</div>
			{#if viewTotal === 0}
				<p class="mt-10 text-center text-sm text-muted-foreground">No matches.</p>
			{/if}
		{/if}
	</div>
</div>
