<script lang="ts">
	import ResultRow from './ResultRow.svelte';
	import type { Row } from '$lib/useLookup.svelte';

	let { rows = [] }: { rows?: Row[] } = $props();
</script>

<ul class="mt-2 grid grid-cols-1 gap-y-0.5 lg:grid-cols-2 lg:gap-x-10">
	{#each rows as row (row.id)}
		<ResultRow {row} />
	{/each}
</ul>
