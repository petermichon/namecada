<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	let {
		examples,
		onselect
	}: {
		examples: string[];
		onselect: (name: string) => void;
	} = $props();
</script>

<div class="mt-6">
	<p class="text-sm text-muted-foreground">Try a name:</p>
	<div class="mt-2 flex flex-wrap gap-2">
		{#each examples as name (name)}
			<Button variant="outline" size="sm" onclick={() => onselect(name)}>
				{name}
			</Button>
		{/each}
	</div>
</div>
