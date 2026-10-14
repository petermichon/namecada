<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import ResultRow from './ResultRow.svelte';
	import type { Row } from '$lib/useLookup.svelte';

	let {
		paidRows = [],
		freeRows = []
	}: {
		paidRows?: Row[];
		freeRows?: Row[];
	} = $props();

	// Unsupported rows are not verdicts: they render in the list like any other
	// row but stay out of the "N of M available" line, so count only rows that
	// carry a verdict.
	const verdicts = (rows: Row[]) => rows.filter((row) => row.status !== 'unsupported');
	const availableCount = (rows: Row[]) =>
		verdicts(rows).filter((row) => row.status === 'available').length;
</script>

{#if paidRows.length}
	<Card.Root class="mt-3 gap-0 py-0">
		<div class="flex flex-wrap items-baseline gap-x-2 px-3 py-2">
			<h2 class="text-sm font-semibold tracking-tight">Paid</h2>
			{#if verdicts(paidRows).length}
				<p class="text-xs text-muted-foreground">
					{availableCount(paidRows)} of {verdicts(paidRows).length} available
				</p>
			{/if}
		</div>
		<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
			{#each paidRows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	</Card.Root>
{/if}

{#if freeRows.length}
	<Card.Root class="mt-3 gap-0 py-0">
		<div class="flex flex-wrap items-baseline gap-x-2 px-3 py-2">
			<h2 class="text-sm font-semibold tracking-tight">Free</h2>
			{#if verdicts(freeRows).length}
				<p class="text-xs text-muted-foreground">
					{availableCount(freeRows)} of {verdicts(freeRows).length} available
				</p>
			{/if}
		</div>
		<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
			{#each freeRows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	</Card.Root>
{/if}
