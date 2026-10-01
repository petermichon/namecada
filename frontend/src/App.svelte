<script lang="ts">
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import {
		lookupDomain,
		parseDomain,
		UnsupportedTldError,
		type LookupResult,
		type ParsedDomain
	} from '$lib/rdap';
	import { getBootstrap } from '$lib/bootstrap';
	import { getCached, isFresh, setCached } from '$lib/cache';
	import { Button } from '$lib/components/ui/button/index.js';
	import LookupForm from '$lib/components/LookupForm.svelte';
	import Result from '$lib/components/Result.svelte';

	getBootstrap().catch(() => {});

	const QUERY_KEY = 'domain-browser:query';

	function loadQuery(): string {
		if (typeof localStorage === 'undefined') return '';
		try {
			return localStorage.getItem(QUERY_KEY) ?? '';
		} catch {
			return '';
		}
	}

	function subdomainNotice(parsed: ParsedDomain): string {
		if (!parsed.subdomain) return '';
		return `You entered a subdomain — showing the registrable domain ${parsed.domain}.`;
	}

	let input = $state(loadQuery());
	let result = $state<LookupResult | null>(null);
	let pending = $state<string | null>(null);
	let notice = $state('');
	let loading = $state(false);
	let revalidating = $state(false);
	let error = $state('');
	let retryable = $state(false);
	let controller: AbortController | null = null;

	async function fetchStatus(domain: string, revalidate: boolean) {
		controller?.abort();
		const ac = new AbortController();
		controller = ac;
		if (revalidate) {
			revalidating = true;
		} else {
			pending = domain;
			loading = true;
		}
		error = '';
		retryable = false;
		try {
			const res = await lookupDomain(domain, ac.signal);
			if (ac.signal.aborted) return;
			if (res.status !== 'unknown') {
				result = res;
				setCached(res.domain, res.status, res.registeredOn, res.raw);
			} else if (!revalidate) {
				result = res;
				retryable = true;
			}
		} catch (e) {
			if (e instanceof DOMException && e.name === 'AbortError') return;
			if (e instanceof UnsupportedTldError) {
				result = null;
				error = `No RDAP endpoint is known for .${e.tld}.`;
				return;
			}
			if (revalidate) return;
			result = null;
			error = 'Network error — could not reach the registry.';
			retryable = true;
		} finally {
			if (controller === ac) {
				loading = false;
				pending = null;
				revalidating = false;
			}
		}
	}

	function check() {
		const parsed = parseDomain(input);
		if (!parsed) {
			result = null;
			notice = '';
			error = 'Enter a domain with a TLD, e.g. acema.com.';
			return;
		}
		notice = subdomainNotice(parsed);
		fetchStatus(parsed.domain, false);
	}

	$effect(() => {
		const parsed = parseDomain(input);

		controller?.abort();
		controller = null;
		pending = null;
		loading = false;
		revalidating = false;

		if (!parsed) {
			result = null;
			error = '';
			retryable = false;
			notice = '';
			return;
		}

		notice = subdomainNotice(parsed);
		const domain = parsed.domain;
		const cached = getCached(domain);
		if (cached) {
			result = {
				domain,
				status: cached.status,
				registeredOn: cached.registeredOn,
				raw: cached.raw
			};
			error = '';
			retryable = false;
			const missingData =
				cached.status === 'registered' &&
				(cached.registeredOn === undefined || cached.raw === undefined);
			if (!isFresh(cached) || missingData) fetchStatus(domain, true);
			return;
		}

		result = null;
		error = '';
		retryable = false;
		fetchStatus(domain, false);
	});

	$effect(() => {
		try {
			localStorage.setItem(QUERY_KEY, input);
		} catch {
			// storage unavailable: persistence is best-effort
		}
	});
</script>

<ModeWatcher />

<div class="min-h-svh bg-background text-foreground">
	<div class="mx-auto max-w-xl px-4 py-16">
		<header class="mb-6 flex items-start justify-between gap-4">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">Domain Browser</h1>
				<p class="mt-1 text-sm text-muted-foreground">Find domain names not registered yet.</p>
			</div>
			<Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
				<SunIcon class="size-4 dark:hidden" />
				<MoonIcon class="hidden size-4 dark:block" />
			</Button>
		</header>

		<LookupForm bind:value={input} onsubmit={check} />

		{#if notice}
			<p class="mt-3 text-xs text-muted-foreground">{notice}</p>
		{/if}

		{#if loading && pending}
			<Result {pending} />
		{/if}

		<div aria-live="polite" aria-atomic="true">
			{#if error}
				<div class="mt-4 flex items-center justify-between gap-3">
					<p class="text-sm text-destructive">{error}</p>
					{#if retryable}
						<Button variant="outline" size="sm" onclick={check}>
							<RefreshCwIcon class="size-3.5" />
							Retry
						</Button>
					{/if}
				</div>
			{/if}

			{#if result && !loading}
				<Result {result} {revalidating} onretry={check} />
			{/if}
		</div>
	</div>
</div>
