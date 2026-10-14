<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';
	import ResultsList from '$lib/components/ResultsList.svelte';
	import StatusFilter from '$lib/components/StatusFilter.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { useLookup } from '$lib/useLookup.svelte';

	const EXAMPLES = ['acema', 'lumen', 'quartz', 'nimbus'];

	const lookup = useLookup();

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
						<LookupForm
							bind:value={lookup.input}
							loading={lookup.searching}
							onsubmit={lookup.check}
						/>
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

			{#if lookup.rows.length}
				<div class="mt-2 flex flex-wrap items-center gap-2">
					{#if lookup.searching}
						<span class="text-xs text-muted-foreground">{lookup.lastLabel}</span>
					{:else}
						<StatusFilter bind:filter={lookup.filter} counts={lookup.counts} />
						<Button variant="outline" size="sm" class="ml-auto" onclick={lookup.copyResults}>
							{lookup.copied ? 'Copied' : 'Copy'}
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

		<span class="sr-only" role="status" aria-live="polite">{lookup.announcement}</span>

		{#if lookup.error}
			<p class="mt-4 text-sm text-destructive">{lookup.error}</p>
		{/if}

		{#if lookup.rows.length}
			{#if lookup.searching}
				<p class="mt-4 text-sm text-muted-foreground">
					Checking {lookup.checked} of {lookup.rows.length}…
				</p>
			{:else}
				<p class="mt-4 text-lg font-semibold tracking-tight">
					{#if lookup.stale}
						Press Enter to check the updated name.
					{:else}
						{lookup.counts.available} of {lookup.rows.length} available
					{/if}
				</p>
			{/if}

			<ResultsList paidRows={lookup.paidRows} freeRows={lookup.freeRows} />
		{:else if !lookup.error}
			<EmptyState
				examples={EXAMPLES}
				onselect={(name) => {
					lookup.input = name;
					lookup.check();
				}}
			/>
		{/if}
	</div>
</div>
