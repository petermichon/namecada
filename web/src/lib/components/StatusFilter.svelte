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
			color: 'text-success'
		},
		{
			value: 'taken' as const,
			label: 'unavailable',
			count: counts.taken,
			Icon: XCircleIcon,
			color: 'text-destructive'
		},
		...(counts.unknown || filter.includes('unknown')
			? [
					{
						value: 'unknown' as const,
						label: 'unknown',
						count: counts.unknown,
						Icon: HelpCircleIcon,
						color: 'text-warning'
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
						color: 'text-muted-foreground'
					}
				]
			: [])
	]);

	// Keep the segments mounted even at zero width so 0 → N transitions.
	const segments = $derived([
		{ value: 'available' as const, count: counts.available, bar: 'bg-success' },
		{ value: 'taken' as const, count: counts.taken, bar: 'bg-destructive' },
		{ value: 'unknown' as const, count: counts.unknown, bar: 'bg-warning' }
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
		{#each segments as segment (segment.value)}
			<span
				class="h-full transition-[width] duration-500 motion-reduce:transition-none {segment.bar}"
				style:width={share(segment.count)}
			></span>
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
