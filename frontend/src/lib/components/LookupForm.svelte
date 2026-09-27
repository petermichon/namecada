<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { autofocus } from '$lib/attachments';

	let {
		value = $bindable(''),
		onsubmit,
		busy = false
	}: {
		value: string;
		onsubmit: () => void;
		busy?: boolean;
	} = $props();
</script>

<form
	class="flex gap-2"
	onsubmit={(e) => {
		e.preventDefault();
		onsubmit();
	}}
>
	<div class="relative flex-1">
		<SearchIcon
			class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
		/>
		<Input
			bind:value
			{@attach autofocus}
			placeholder="search, e.g. acema"
			class="h-10 pl-9"
			autocomplete="off"
			autocapitalize="off"
			spellcheck={false}
		/>
	</div>
	<Button type="submit" class="h-10" disabled={busy}>
		{busy ? 'Checking…' : 'Check'}
	</Button>
</form>
