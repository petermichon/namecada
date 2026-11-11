<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import CircleMinusIcon from '@lucide/svelte/icons/circle-minus';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import type { Filter, FilterStatus } from '$lib/useLookup.svelte';

	let {
		filter = $bindable<Filter>([]),
		counts,
		total
	}: {
		filter: Filter;
		counts: { available: number; taken: number; unknown: number; unsupported: number };
		total: number;
	} = $props();

	const rows = $derived([
		{
			value: 'available' as const,
			label: 'available',
			count: counts.available,
			Icon: CheckCircle2Icon,
			color: 'text-success',
			bar: 'bg-success'
		},
		{
			value: 'taken' as const,
			label: 'unavailable',
			count: counts.taken,
			Icon: XCircleIcon,
			color: 'text-destructive',
			bar: 'bg-destructive'
		},
		...(counts.unknown || filter.includes('unknown')
			? [
					{
						value: 'unknown' as const,
						label: 'unknown',
						count: counts.unknown,
						Icon: HelpCircleIcon,
						color: 'text-warning',
						bar: 'bg-warning'
					}
				]
			: []),
		...(counts.unsupported || filter.includes('unsupported')
			? [
					{
						value: 'unsupported' as const,
						label: 'not checked',
						count: counts.unsupported,
						Icon: CircleMinusIcon,
						color: 'text-muted-foreground',
						bar: 'bg-muted-foreground/40'
					}
				]
			: [])
	]);

	function toggle(value: FilterStatus) {
		filter = filter.includes(value)
			? filter.filter((status) => status !== value)
			: [...filter, value];
	}

	function share(count: number): string {
		return total > 0 ? `${(count / total) * 100}%` : '0%';
	}
</script>

<div class="space-y-3">
	<div class="flex h-2 w-full gap-px overflow-hidden rounded-full bg-muted" aria-hidden="true">
		{#each rows as row (row.value)}
			{#if row.count > 0}
				<span
					class="h-full transition-[width] duration-500 motion-reduce:transition-none {row.bar}"
					style:width={share(row.count)}
				></span>
			{/if}
		{/each}
	</div>

	<div class="flex flex-wrap items-center gap-1" role="group" aria-label="Filter results by status">
		{#each rows as row (row.value)}
			<button
				type="button"
				aria-pressed={filter.includes(row.value)}
				onclick={() => toggle(row.value)}
				class="inline-flex min-h-9 items-center gap-1.5 rounded-lg border border-transparent px-2.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-pressed:border-border aria-pressed:bg-card aria-pressed:text-foreground aria-pressed:shadow-xs"
			>
				<row.Icon class="size-3.5 shrink-0 {row.color}" aria-hidden="true" />
				<span>{row.label}</span>
				<span class="font-semibold tabular-nums">{row.count}</span>
			</button>
		{/each}
	</div>
</div>
