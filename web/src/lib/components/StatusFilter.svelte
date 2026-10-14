<script lang="ts">
	import * as ToggleGroup from '$lib/components/ui/toggle-group/index.js';
	import type { Filter } from '$lib/useLookup.svelte';

	let {
		filter = $bindable<Filter>('all'),
		counts,
		disabled = false
	}: {
		filter: Filter;
		counts: { available: number; taken: number; unknown: number };
		disabled?: boolean;
	} = $props();
</script>

<ToggleGroup.Root
	type="single"
	value={filter}
	onValueChange={(value) => (filter = (value || 'all') as Filter)}
	variant="outline"
	size="sm"
	spacing={2}
	{disabled}
	class="text-xs"
>
	<ToggleGroup.Item
		value="available"
		class="data-[state=on]:border-success/40 data-[state=on]:bg-success/15 data-[state=on]:text-success"
	>
		{counts.available} available
	</ToggleGroup.Item>
	<ToggleGroup.Item value="taken" class="data-[state=on]:text-foreground">
		{counts.taken} unavailable
	</ToggleGroup.Item>
	{#if counts.unknown}
		<ToggleGroup.Item
			value="unknown"
			class="data-[state=on]:border-warning/40 data-[state=on]:bg-warning/15 data-[state=on]:text-warning"
		>
			{counts.unknown} unknown
		</ToggleGroup.Item>
	{/if}
</ToggleGroup.Root>
