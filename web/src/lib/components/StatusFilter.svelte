<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import CircleMinusIcon from '@lucide/svelte/icons/circle-minus';
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
			bar: 'bg-success',
			active: 'aria-pressed:bg-success/10 aria-pressed:hover:bg-success/20'
		},
		{
			value: 'taken' as const,
			label: 'unavailable',
			count: counts.taken,
			Icon: XCircleIcon,
			color: 'text-destructive',
			bar: 'bg-destructive',
			active: 'aria-pressed:bg-destructive/10 aria-pressed:hover:bg-destructive/20'
		},
		...(counts.unknown || filter.includes('unknown')
			? [
					{
						value: 'unknown' as const,
						label: 'unknown',
						count: counts.unknown,
						Icon: HelpCircleIcon,
						color: 'text-warning',
						bar: 'bg-warning',
						active: 'aria-pressed:bg-warning/10 aria-pressed:hover:bg-warning/20'
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
						bar: 'bg-muted-foreground/40',
						active: 'aria-pressed:bg-muted aria-pressed:hover:bg-foreground/[0.06]'
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

<div class="space-y-0.5" role="group" aria-label="Filter results by status">
	{#each rows as row (row.value)}
		<button
			type="button"
			aria-pressed={filter.includes(row.value)}
			onclick={() => toggle(row.value)}
			class="grid w-full grid-cols-[auto_6.5rem_2rem_1fr] items-center gap-x-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring {row.active}"
		>
			<row.Icon class="size-3.5 shrink-0 {row.color}" aria-hidden="true" />
			<span class="truncate text-foreground">{row.label}</span>
			<span class="text-right text-[13px] font-medium text-foreground tabular-nums"
				>{row.count}</span
			>
			{#if row.bar}
				<span class="h-[3px] w-full overflow-hidden rounded-full bg-border">
					<span
						class="block h-full rounded-full {row.bar} transition-[width] duration-300 motion-reduce:transition-none"
						style:width={share(row.count)}
					></span>
				</span>
			{:else}
				<span aria-hidden="true"></span>
			{/if}
		</button>
	{/each}
</div>
