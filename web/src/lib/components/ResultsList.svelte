<script lang="ts">
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

{#snippet group(title: string, rows: Row[])}
	<section class="mt-8 first:mt-6">
		<div class="flex items-baseline gap-3">
			<h2 class="text-sm font-semibold tracking-tight">{title}</h2>
			{#if verdicts(rows).length}
				<p class="text-xs text-muted-foreground">
					{availableCount(rows)} of {verdicts(rows).length} available
				</p>
			{/if}
			<span class="ml-1 h-px flex-1 self-center bg-border" aria-hidden="true"></span>
		</div>
		<ul class="mt-1 grid grid-cols-1 lg:grid-cols-2 lg:gap-x-10">
			{#each rows as row (row.id)}
				<ResultRow {row} />
			{/each}
		</ul>
	</section>
{/snippet}

{#if paidRows.length}
	{@render group('Paid platforms', paidRows)}
{/if}

{#if freeRows.length}
	{@render group('Free platforms', freeRows)}
{/if}
