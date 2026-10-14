<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import ResultRow from './ResultRow.svelte';
	import type { Row } from '$lib/useLookup.svelte';

	let { paidRows = [], freeRows = [] }: { paidRows?: Row[]; freeRows?: Row[] } = $props();
</script>

<Card.Root class="mt-3 gap-0 py-0">
	{#if paidRows.length}
		<div class="flex flex-wrap items-baseline gap-x-2 px-3 py-2">
			<h2 class="text-sm font-semibold tracking-tight">Paid</h2>
			<p class="text-xs text-muted-foreground">
				Costs money to register; the amount varies (premium names cost more).
			</p>
		</div>
		<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
			{#each paidRows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	{/if}
	{#if freeRows.length}
		<div class="border-t border-border px-3 py-2">
			<h2 class="text-sm font-semibold tracking-tight">Free</h2>
		</div>
		<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
			{#each freeRows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	{/if}
</Card.Root>
