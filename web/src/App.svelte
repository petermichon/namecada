<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';
	import ResultsList from '$lib/components/ResultsList.svelte';
	import StatusFilter from '$lib/components/StatusFilter.svelte';
	import { useLookup } from '$lib/useLookup.svelte';

	const lookup = useLookup();

	const totalSources = $derived(lookup.rows.length);

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

{#snippet themeButton()}
	<Button
		variant="ghost"
		size="icon"
		class="-mr-1.5 shrink-0"
		onclick={toggleMode}
		aria-label="Toggle theme"
	>
		<SunIcon class="size-4 dark:hidden" />
		<MoonIcon class="hidden size-4 dark:block" />
	</Button>
{/snippet}

{#snippet feedback()}
	{#if lookup.error}
		<p id="name-error" role="alert" class="mt-2 text-sm text-destructive">{lookup.error}</p>
	{/if}
	{#if !lookup.error}
		<span class="sr-only" role="status" aria-live="polite">{lookup.announcement}</span>
	{/if}
{/snippet}

<ModeWatcher />

<div class="flex min-h-svh flex-col bg-background text-foreground">
	<a
		href="#main"
		class="sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-4 focus:z-30 focus:rounded-md focus:border focus:border-border focus:bg-card focus:px-3 focus:py-1.5 focus:text-sm"
		>Skip to main content</a
	>

	{#if lookup.lastLabel === null}
		<header class="flex justify-end px-4 py-3">
			{@render themeButton()}
		</header>

		<main id="main" class="flex flex-1 flex-col items-center justify-center px-4 pb-28 sm:pb-36">
			<div class="w-full max-w-xl">
				<h1 class="text-center text-4xl font-semibold tracking-tight sm:text-5xl">Namecheck</h1>
				<p class="mt-2 text-center text-sm text-muted-foreground">See where a name is available.</p>

				<div class="mt-8">
					<LookupForm
						bind:value={lookup.input}
						loading={lookup.searching}
						invalid={!!lookup.error}
						errorId="name-error"
						onsubmit={lookup.check}
					/>
				</div>

				{@render feedback()}

				{#if !lookup.error}
					<p class="mt-3 text-center text-xs text-muted-foreground">
						Checks domains, packages and usernames.
					</p>
				{/if}
			</div>
		</main>
	{:else}
		<header class="sticky top-0 z-20 border-b border-border/60 bg-background/85 backdrop-blur">
			<div class="mx-auto flex max-w-2xl items-center gap-3 px-4 py-2.5 lg:max-w-5xl">
				<h1 class="shrink-0 text-base font-semibold tracking-tight">
					<a
						href="./"
						onclick={onHomeClick}
						class="rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
						>Namecheck</a
					>
				</h1>
				<div class="flex min-w-0 flex-1 justify-center">
					<div class="w-full max-w-xl">
						<LookupForm
							bind:value={lookup.input}
							loading={lookup.searching}
							invalid={!!lookup.error}
							errorId="name-error"
							onsubmit={lookup.check}
						/>
					</div>
				</div>
				{@render themeButton()}
			</div>
		</header>

		<main id="main" class="mx-auto w-full max-w-2xl flex-1 px-4 py-5 lg:max-w-5xl">
			{@render feedback()}

			{#if lookup.rows.length}
				<div class="mt-5">
					<StatusFilter bind:filter={lookup.filter} counts={lookup.counts} total={totalSources} />
				</div>

				{#if lookup.filter.length === 0}
					<p class="mt-8 text-sm text-muted-foreground">No statuses selected.</p>
				{:else}
					<div aria-busy={lookup.searching}>
						<ResultsList paidRows={lookup.paidRows} freeRows={lookup.freeRows} />
					</div>
				{/if}
			{/if}

			<footer
				class="mt-10 max-w-prose border-t border-border/60 pt-4 text-xs leading-relaxed text-muted-foreground"
			>
				<p>
					Best-effort from public sources. <em>Available</em> means no public block was found — not a
					guarantee you can register it. Private accounts and organizations can appear available.
				</p>
			</footer>
		</main>
	{/if}
</div>
