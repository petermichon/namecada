<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import { lookupDomain, normalizeDomain, type LookupResult } from '$lib/rdap';
	import { getCached, setCached } from '$lib/cache';
	import { Button } from '$lib/components/ui/button/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';
	import Result from '$lib/components/Result.svelte';

	let input = $state('');
	let result = $state<LookupResult | null>(null);
	let loading = $state(false);
	let error = $state('');
	let controller: AbortController | null = null;

	async function check() {
		error = '';
		const normalized = normalizeDomain(input);
		if (!normalized) {
			result = null;
			error = 'Enter a valid domain label — letters, digits and inner hyphens only.';
			return;
		}

		const cached = getCached(normalized);
		if (cached) {
			result = { domain: normalized, status: cached.status };
			return;
		}

		controller?.abort();
		controller = new AbortController();
		loading = true;
		try {
			const res = await lookupDomain(normalized, controller.signal);
			result = res;
			if (res.status !== 'unknown') setCached(res.domain, res.status);
		} catch (e) {
			if (!(e instanceof DOMException && e.name === 'AbortError')) {
				result = { domain: normalized, status: 'unknown' };
				error = 'Network error — could not reach the registry.';
			}
		} finally {
			loading = false;
		}
	}
</script>

<ModeWatcher />

<div class="min-h-svh bg-background text-foreground">
	<div class="mx-auto max-w-xl px-4 py-16">
		<header class="mb-6 flex items-start justify-between gap-4">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">.com availability checker</h1>
				<p class="mt-1 text-sm text-muted-foreground">
					Authoritative registry lookup via Verisign RDAP.
				</p>
			</div>
			<Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
				<SunIcon class="size-4 dark:hidden" />
				<MoonIcon class="hidden size-4 dark:block" />
			</Button>
		</header>

		<LookupForm bind:value={input} onsubmit={check} busy={loading} />

		{#if error}
			<p class="mt-4 text-sm text-destructive">{error}</p>
		{/if}

		{#if result && !loading}
			<Result {result} />
		{/if}
	</div>
</div>
