<script lang="ts">
	import { CATEGORIES, type CategoryFilter, type CostFilter } from '$lib/platforms';

	let {
		category = $bindable<CategoryFilter>('all'),
		cost = $bindable<CostFilter>('all'),
		categoryCounts,
		costCounts,
		total
	}: {
		category: CategoryFilter;
		cost: CostFilter;
		categoryCounts: Record<string, number>;
		costCounts: { all: number; free: number; paid: number };
		total: number;
	} = $props();

	const CHIP =
		'inline-flex min-h-9 shrink-0 items-center gap-1.5 rounded-lg border border-transparent px-2.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-pressed:border-border aria-pressed:bg-card aria-pressed:text-foreground aria-pressed:shadow-xs';

	const costs = $derived([
		{ value: 'all' as const, label: 'All', count: costCounts.all },
		{ value: 'free' as const, label: 'Free', count: costCounts.free },
		{ value: 'paid' as const, label: 'Paid', count: costCounts.paid }
	]);
</script>

<div class="space-y-2">
	<div class="space-y-1">
		<p class="text-[11px] font-semibold tracking-[0.16em] text-muted-foreground uppercase">Tags</p>
		<div
			class="-mx-1 flex [scrollbar-width:none] gap-1 overflow-x-auto px-1 pb-0.5 sm:mx-0 sm:flex-wrap sm:overflow-visible sm:px-0 [&::-webkit-scrollbar]:hidden"
			role="group"
			aria-label="Filter by tag"
		>
			<button
				type="button"
				class={CHIP}
				aria-pressed={category === 'all'}
				onclick={() => (category = 'all')}
			>
				All <span class="font-semibold tabular-nums">{total}</span>
			</button>
			{#each CATEGORIES as item (item.id)}
				{@const count = categoryCounts[item.id] ?? 0}
				{#if count > 0 || category === item.id}
					<button
						type="button"
						class={CHIP}
						aria-pressed={category === item.id}
						onclick={() => (category = item.id)}
					>
						{item.label} <span class="font-semibold tabular-nums">{count}</span>
					</button>
				{/if}
			{/each}
		</div>
	</div>

	<div class="space-y-1">
		<p class="text-[11px] font-semibold tracking-[0.16em] text-muted-foreground uppercase">Cost</p>
		<div class="flex flex-wrap gap-1" role="group" aria-label="Filter by cost">
			{#each costs as item (item.value)}
				<button
					type="button"
					class={CHIP}
					aria-pressed={cost === item.value}
					onclick={() => (cost = item.value)}
				>
					{item.label} <span class="font-semibold tabular-nums">{item.count}</span>
				</button>
			{/each}
		</div>
	</div>
</div>
