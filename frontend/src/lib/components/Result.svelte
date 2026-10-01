<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import { Button } from '$lib/components/ui/button/index.js';
	import Details from '$lib/components/Details.svelte';
	import type { LookupResult } from '$lib/rdap';

	let {
		result = null,
		pending = null,
		revalidating = false,
		onretry
	}: {
		result?: LookupResult | null;
		pending?: string | null;
		revalidating?: boolean;
		onretry?: () => void;
	} = $props();

	const view = $derived.by(() => {
		switch (result?.status) {
			case 'unregistered':
				return {
					verdict: 'Not registered',
					note: 'Not in the registry — may still be reserved or premium.',
					Icon: CheckCircle2Icon
				};
			case 'registered':
				return {
					verdict: 'Registered',
					note: '',
					Icon: XCircleIcon
				};
			default:
				return {
					verdict: 'Unknown',
					note: 'The registry did not return a definitive answer.',
					Icon: HelpCircleIcon
				};
		}
	});

	const showRetry = $derived(result?.status === 'unknown' && !!onretry);
</script>

{#snippet domainName(value: string)}
	{@const dot = value.indexOf('.')}
	<span class="truncate font-mono text-sm text-muted-foreground"
		>{value.slice(0, dot)}<span class="opacity-60">{value.slice(dot)}</span></span
	>
{/snippet}

{#if pending}
	<div class="mt-6 rounded-lg border p-4">
		<div class="flex items-center gap-3">
			<LoaderCircleIcon class="size-6 shrink-0 animate-spin text-muted-foreground" />
			<div class="flex min-w-0 flex-col">
				<span class="text-lg font-semibold">Checking…</span>
				{@render domainName(pending)}
			</div>
		</div>
	</div>
{:else if result && view}
	<div class="mt-6 rounded-lg border p-4">
		<div class="flex items-center gap-3">
			<view.Icon class="size-6 shrink-0" />
			<div class="flex min-w-0 flex-1 flex-col">
				<span class="text-lg font-semibold">{view.verdict}</span>
				{@render domainName(result.domain)}
			</div>
			{#if revalidating}
				<RefreshCwIcon class="size-4 shrink-0 animate-spin text-muted-foreground" />
			{/if}
		</div>
		{#if view.note || showRetry}
			<div class="mt-3 flex items-center justify-between gap-3">
				{#if view.note}
					<p class="text-sm text-muted-foreground">{view.note}</p>
				{/if}
				{#if showRetry}
					<Button variant="outline" size="sm" class="ml-auto shrink-0" onclick={onretry}>
						<RefreshCwIcon class="size-3.5" />
						Retry
					</Button>
				{/if}
			</div>
		{/if}
	</div>
	{#if result.raw !== undefined}
		<Details raw={result.raw} />
	{/if}
{/if}
