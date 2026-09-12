<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { Mode } from '$lib/search';

	let {
		query = $bindable(''),
		mode = $bindable<Mode>('prefix'),
		onlyAvailable = $bindable(true)
	}: {
		query: string;
		mode: Mode;
		onlyAvailable: boolean;
	} = $props();
</script>

<div class="flex gap-2">
	<div class="relative flex-1">
		<SearchIcon
			class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
		/>
		<Input bind:value={query} placeholder="search, e.g. acema" class="h-10 pl-9" autofocus />
	</div>
	<div class="flex rounded-lg border p-1" role="group" aria-label="Search mode">
		<Button
			variant={mode === 'prefix' ? 'secondary' : 'ghost'}
			size="sm"
			aria-pressed={mode === 'prefix'}
			onclick={() => (mode = 'prefix')}>Prefix</Button
		>
		<Button
			variant={mode === 'contains' ? 'secondary' : 'ghost'}
			size="sm"
			aria-pressed={mode === 'contains'}
			onclick={() => (mode = 'contains')}>Contains</Button
		>
	</div>
</div>

<label class="mt-3 flex w-fit items-center gap-2 text-sm text-muted-foreground select-none">
	<input
		type="checkbox"
		bind:checked={onlyAvailable}
		class="size-4 rounded-sm accent-primary focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none"
	/>
	Only available
</label>
