<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import CircleMinusIcon from '@lucide/svelte/icons/circle-minus';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import type { Row } from '$lib/useLookup.svelte';

	const STATUS: Record<
		Row['status'],
		{ text: string; cls: string; Icon?: typeof CheckCircle2Icon }
	> = {
		available: { text: 'available', cls: 'text-success', Icon: CheckCircle2Icon },
		taken: { text: 'unavailable', cls: 'text-destructive', Icon: XCircleIcon },
		unknown: { text: 'unknown', cls: 'text-warning', Icon: HelpCircleIcon },
		unsupported: { text: 'not checked', cls: 'text-muted-foreground', Icon: CircleMinusIcon },
		pending: { text: 'checking', cls: 'text-muted-foreground' }
	};

	let { row }: { row: Row } = $props();

	const pending = $derived(row.status === 'pending');
	const status = $derived(STATUS[row.status]);
	const text = $derived(row.url.replace(/^https?:\/\//, ''));
	const Icon = $derived(status.Icon);
</script>

<li>
	<a
		href={pending ? undefined : row.url}
		target="_blank"
		rel="noopener noreferrer"
		class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-0.5 rounded-md px-2 py-2 transition-colors hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none focus-visible:ring-inset sm:grid-cols-[7.5rem_minmax(0,1fr)_auto]"
	>
		{#if pending}
			<Skeleton class="col-start-1 row-start-1 h-3 w-24 rounded-sm sm:w-20" />
			<div class="col-start-1 row-start-2 min-w-0 sm:col-start-2 sm:row-start-1">
				<Skeleton class="h-4 w-48 max-w-full rounded-sm" />
			</div>
			<div class="col-start-2 row-span-2 row-start-1 justify-self-end sm:col-start-3 sm:row-span-1">
				<Skeleton class="h-3 w-16 rounded-full" />
			</div>
		{:else}
			<span
				class="col-start-1 row-start-1 flex min-w-0 items-center gap-1.5 text-[11px] font-medium text-muted-foreground"
			>
				<span class="truncate">{row.platformName}</span>
				{#if row.paid}
					<span
						class="shrink-0 rounded border border-border bg-muted/60 px-1 py-px text-[9px] font-semibold tracking-[0.08em] uppercase"
						>Paid</span
					>
				{/if}
			</span>

			<div
				class="col-start-1 row-start-2 flex min-w-0 items-center gap-2 sm:col-start-2 sm:row-start-1"
			>
				<span
					class="truncate text-sm {row.status === 'taken'
						? 'text-muted-foreground'
						: 'text-foreground'}">{text}</span
				>
				{#if row.detail}
					<span class="hidden shrink-0 text-xs text-muted-foreground sm:inline">{row.detail}</span>
				{/if}
			</div>

			<span
				class="col-start-2 row-span-2 row-start-1 flex items-center gap-1.5 justify-self-end text-xs sm:col-start-3 sm:row-span-1 {status.cls}"
			>
				{#if Icon}<Icon class="size-3.5 shrink-0" aria-hidden="true" />{/if}
				<span class="w-20 text-right">{status.text}</span>
			</span>
		{/if}
	</a>
</li>
