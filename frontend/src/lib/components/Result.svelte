<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
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
					label: 'not registered',
					variant: 'default' as const,
					note: 'No registry record — this domain is unregistered.',
					Icon: CheckCircle2Icon
				};
			case 'registered':
				return {
					label: 'registered',
					variant: 'secondary' as const,
					note: 'A registry record exists — this domain is taken.',
					Icon: XCircleIcon
				};
			default:
				return {
					label: 'unknown',
					variant: 'outline' as const,
					note: 'The registry did not return a definitive answer.',
					Icon: HelpCircleIcon
				};
		}
	});
</script>

{#snippet domainName(value: string)}
	{@const dot = value.indexOf('.')}
	<span class="truncate font-mono text-lg"
		>{value.slice(0, dot)}<span class="text-muted-foreground">{value.slice(dot)}</span></span
	>
{/snippet}

{#if pending}
	<div class="mt-6 rounded-lg border p-4">
		<div class="flex items-center justify-between gap-3">
			<div class="flex min-w-0 items-center gap-2">
				<LoaderCircleIcon class="size-5 shrink-0 animate-spin text-muted-foreground" />
				{@render domainName(pending)}
			</div>
			<Badge variant="outline">checking…</Badge>
		</div>
		<p class="mt-2 text-sm text-muted-foreground">Querying the registry…</p>
	</div>
{:else if result && view}
	<div class="mt-6 rounded-lg border p-4">
		<div class="flex items-center justify-between gap-3">
			<div class="flex min-w-0 items-center gap-2">
				<view.Icon class="size-5 shrink-0 text-muted-foreground" />
				{@render domainName(result.domain)}
			</div>
			<div class="flex shrink-0 items-center gap-2">
				{#if revalidating}
					<RefreshCwIcon class="size-3.5 animate-spin text-muted-foreground" />
				{/if}
				<Badge variant={view.variant}>{view.label}</Badge>
			</div>
		</div>
		<div class="mt-2 flex items-center justify-between gap-2">
			<p class="text-sm text-muted-foreground">{view.note}</p>
			{#if result.status === 'unknown' && onretry}
				<Button variant="outline" size="sm" onclick={onretry}>
					<RefreshCwIcon class="size-3.5" />
					Retry
				</Button>
			{/if}
		</div>
	</div>
{/if}
