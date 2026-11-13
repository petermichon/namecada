<script lang="ts">
	import InfoIcon from '@lucide/svelte/icons/info';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import SunIcon from '@lucide/svelte/icons/sun';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CostFilter from '$lib/components/CostFilter.svelte';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import LookupForm from '$lib/components/LookupForm.svelte';
	import ResultRow from '$lib/components/ResultRow.svelte';
	import StatusFilter from '$lib/components/StatusFilter.svelte';
	import { useLookup } from '$lib/useLookup.svelte';

	const lookup = useLookup();

	// Home is "no scan yet"; results is everything after.
	const home = $derived(lookup.lastLabel === null);

	// One empty block; only the copy changes with the cause.
	function emptyState(statusCount: number, rowCount: number) {
		if (statusCount === 0) {
			return {
				title: 'No statuses selected.',
				body: 'Choose a status above, or show everything.'
			};
		}
		if (rowCount === 0) {
			return {
				title: 'No sources match these filters.',
				body: 'Clear a filter above, or show everything.'
			};
		}
		return null;
	}

	const empty = $derived(emptyState(lookup.filter.length, lookup.visibleRows.length));

	function onHomeClick(e: MouseEvent) {
		if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
		e.preventDefault();
		lookup.goHome();
		scrollTo(0, 0);
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
		addEventListener('keydown', onGlobalKeydown);
		return () => removeEventListener('keydown', onGlobalKeydown);
	});
</script>

{#snippet brand()}
	<a
		href="./"
		onclick={onHomeClick}
		class="inline-flex items-center gap-2 rounded-lg transition-opacity hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring"
	>
		<BrandMark class="size-7 text-primary" />
		<span class="text-lg font-semibold tracking-[-0.02em]">Namecada</span>
	</a>
{/snippet}

{#snippet themeButton()}
	<Button
		variant="ghost"
		size="icon"
		class="size-10 rounded-lg text-muted-foreground hover:text-foreground sm:size-9"
		onclick={toggleMode}
		aria-label="Toggle theme"
	>
		<SunIcon class="size-4 dark:hidden" />
		<MoonIcon class="hidden size-4 dark:block" />
	</Button>
{/snippet}

{#snippet feedback()}
	<span class="sr-only" role="status" aria-live="polite"
		>{lookup.failure ? '' : lookup.announcement}</span
	>
{/snippet}

<ModeWatcher />

<div class="flex min-h-svh flex-col bg-background text-foreground">
	<a
		href="#main"
		class="sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-4 focus:z-30 focus:rounded-md focus:border focus:border-border focus:bg-card focus:px-3 focus:py-1.5 focus:text-sm"
		>Skip to main content</a
	>

	{#if home}
		<header
			class="mx-auto flex w-full max-w-6xl items-center justify-between border-b border-transparent px-4 py-4 sm:h-[65px] sm:px-6 sm:py-0"
		>
			{@render brand()}
			{@render themeButton()}
		</header>

		<main
			id="main"
			class="mx-auto flex w-full max-w-6xl flex-1 flex-col justify-center px-4 pb-24 sm:px-6 sm:pb-32"
		>
			<div class="mx-auto flex w-full max-w-2xl flex-col items-center text-center">
				<h1 class="max-w-2xl text-4xl font-semibold tracking-[-0.035em] text-balance sm:text-6xl">
					Check a name across the open web.
				</h1>
				<p class="mt-4 max-w-xl text-base leading-7 text-pretty text-muted-foreground">
					Domains, usernames and package registries, checked in one pass.
				</p>

				<div class="mt-8 w-full max-w-xl">
					<LookupForm
						bind:value={lookup.input}
						loading={lookup.searching}
						onsubmit={lookup.check}
					/>
				</div>

				<div class="relative w-full [&>p]:justify-center">
					<div class="absolute inset-x-0 top-0">{@render feedback()}</div>
				</div>
			</div>
		</main>
	{:else}
		<header class="sticky top-0 z-20 border-b border-border/70 bg-background">
			<div
				class="mx-auto flex w-full max-w-6xl flex-col gap-2 px-4 py-2.5 sm:h-16 sm:flex-row sm:items-center sm:gap-3 sm:px-6"
			>
				<div class="flex items-center justify-between gap-3 sm:justify-start">
					{@render brand()}
					<div class="sm:hidden">{@render themeButton()}</div>
				</div>
				<div class="min-w-0 flex-1">
					<div class="mx-auto w-full max-w-xl">
						<LookupForm
							bind:value={lookup.input}
							loading={lookup.searching}
							onsubmit={lookup.check}
						/>
					</div>
				</div>
				<div class="hidden sm:block">{@render themeButton()}</div>
			</div>
		</header>

		<main id="main" class="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:px-6 sm:py-8">
			{@render feedback()}

			{#if lookup.failure}
				<div
					role="alert"
					class="mt-2 flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
				>
					<TriangleAlertIcon class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
					<p class="min-w-0 flex-1">
						{lookup.failure}
						<button
							type="button"
							class="ml-1 font-medium underline underline-offset-4 hover:no-underline focus-visible:rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-destructive"
							onclick={() => lookup.retry()}>Try again</button
						>
					</p>
				</div>
			{/if}

			{#if lookup.rows.length}
				<section class="mt-4 flex flex-col gap-4 sm:mt-2 sm:gap-5">
					<div class="flex flex-wrap items-end justify-between gap-x-8 gap-y-3">
						<div class="min-w-0">
							<h1 class="text-3xl font-semibold tracking-[-0.03em] break-all sm:text-4xl">
								{lookup.lastLabel}
							</h1>
						</div>

						<div
							class="flex max-w-sm items-start gap-1.5 text-sm leading-5 text-pretty text-muted-foreground"
						>
							<InfoIcon class="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
							<p>Availability doesn’t guarantee you can register the name.</p>
						</div>
					</div>

					<StatusFilter
						bind:filter={lookup.filter}
						counts={lookup.counts}
						total={lookup.statusTotal}
					/>

					<CostFilter bind:cost={lookup.cost} costCounts={lookup.costCounts} />
				</section>
			{/if}

			{#if !lookup.failure && (lookup.rows.length || !lookup.searching)}
				<div class="mt-8" aria-busy={lookup.searching}>
					{#if empty}
						<div class="rounded-lg border border-dashed border-border px-4 py-14 text-center">
							<p class="text-sm font-medium">{empty.title}</p>
							<p class="mt-1 text-sm text-muted-foreground">{empty.body}</p>
							<Button
								variant="outline"
								size="sm"
								class="mt-4 rounded-lg"
								onclick={() => lookup.showAll()}
							>
								Show all
							</Button>
						</div>
					{:else}
						<ul class="mt-2 grid grid-cols-1 gap-y-0.5 lg:grid-cols-2 lg:gap-x-10">
							{#each lookup.visibleRows as row (row.id)}
								<ResultRow {row} />
							{/each}
						</ul>
					{/if}
				</div>
			{/if}
		</main>
	{/if}
</div>
