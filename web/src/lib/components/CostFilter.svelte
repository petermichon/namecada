<script lang="ts">
	import DollarSignIcon from '@lucide/svelte/icons/dollar-sign';
	import DollarOffIcon from '$lib/components/icons/DollarOff.svelte';
	import type { CostFilter, CostGroup } from '$lib/platforms';

	let {
		cost = $bindable<CostFilter>(['free', 'paid']),
		costCounts
	}: {
		cost: CostFilter;
		costCounts: { free: number; paid: number };
	} = $props();

	const CHIP =
		'inline-flex min-h-10 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-lg border px-2.5 text-xs font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring sm:min-h-9';

	const costs = $derived([
		{
			value: 'free' as const,
			label: 'Free',
			count: costCounts.free,
			Icon: DollarOffIcon,
			color: 'text-success',
			active: 'border-success/40 bg-success/10 text-foreground shadow-xs hover:bg-success/15'
		},
		{
			value: 'paid' as const,
			label: 'Paid',
			count: costCounts.paid,
			Icon: DollarSignIcon,
			color: 'text-warning',
			active: 'border-warning/40 bg-warning/10 text-foreground shadow-xs hover:bg-warning/15'
		}
	]);

	function toggle(value: CostGroup) {
		cost = cost.includes(value) ? cost.filter((group) => group !== value) : [...cost, value];
	}
</script>

<div
	class="-mx-4 -my-1 flex [scrollbar-width:none] items-center gap-1 overflow-x-auto px-4 py-1 sm:mx-0 sm:my-0 sm:flex-wrap sm:overflow-visible sm:px-0 sm:py-0 [&::-webkit-scrollbar]:hidden"
	role="group"
	aria-label="Filter by cost"
>
	{#each costs as item (item.value)}
		<button
			type="button"
			class="{CHIP} {cost.includes(item.value)
				? item.active
				: 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground'}"
			aria-pressed={cost.includes(item.value)}
			onclick={() => toggle(item.value)}
		>
			<item.Icon class="size-3.5 shrink-0 {item.color}" aria-hidden="true" />
			{item.label} <span class="font-semibold tabular-nums">{item.count}</span>
		</button>
	{/each}
</div>
