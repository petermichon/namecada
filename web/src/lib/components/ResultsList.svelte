<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import ResultRow from './ResultRow.svelte';
	import type { Row } from '$lib/useLookup.svelte';

	let { rows }: { rows: Row[] } = $props();
</script>

<Card.Root class="mt-3 gap-0 py-0">
	<ul class="grid grid-cols-1 gap-px bg-border lg:grid-cols-2">
		{#each rows as row (row.id)}
			<ResultRow {row} />
		{/each}
	</ul>
</Card.Root>
