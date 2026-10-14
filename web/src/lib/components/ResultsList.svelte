<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import ResultRow from './ResultRow.svelte';
	import type { Row } from '$lib/useLookup.svelte';

	let {
		paidRows = [],
		freeRows = [],
		paidUnsupportedRows = [],
		freeUnsupportedRows = []
	}: {
		paidRows?: Row[];
		freeRows?: Row[];
		paidUnsupportedRows?: Row[];
		freeUnsupportedRows?: Row[];
	} = $props();

	const availableCount = (rows: Row[]) => rows.filter((row) => row.status === 'available').length;
</script>

{#if paidRows.length || paidUnsupportedRows.length}
	<Card.Root class="mt-3 gap-0 py-0">
		<div class="flex flex-wrap items-baseline gap-x-2 px-3 py-2">
			<h2 class="text-sm font-semibold tracking-tight">Paid</h2>
			{#if paidRows.length}
				<p class="text-xs text-muted-foreground">
					{availableCount(paidRows)} of {paidRows.length} available
				</p>
			{/if}
		</div>
		<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
			{#each paidRows as row (row.id)}
				<ResultRow {row} />
			{/each}
			{#each paidUnsupportedRows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	</Card.Root>
{/if}

{#if freeRows.length || freeUnsupportedRows.length}
	<Card.Root class="mt-3 gap-0 py-0">
		<div class="flex flex-wrap items-baseline gap-x-2 px-3 py-2">
			<h2 class="text-sm font-semibold tracking-tight">Free</h2>
			{#if freeRows.length}
				<p class="text-xs text-muted-foreground">
					{availableCount(freeRows)} of {freeRows.length} available
				</p>
			{/if}
		</div>
		<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
			{#each freeRows as row (row.id)}
				<ResultRow {row} />
			{/each}
			{#each freeUnsupportedRows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	</Card.Root>
{/if}
