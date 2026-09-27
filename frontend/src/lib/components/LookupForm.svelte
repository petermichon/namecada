<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { Input } from '$lib/components/ui/input/index.js';
	import { autofocus } from '$lib/attachments';

	let {
		value = $bindable(''),
		onsubmit
	}: {
		value: string;
		onsubmit: () => void;
	} = $props();
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		onsubmit();
	}}
>
	<div class="relative">
		<label for="domain" class="sr-only">Domain name</label>
		<SearchIcon
			class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
		/>
		<Input
			id="domain"
			name="domain"
			bind:value
			{@attach autofocus}
			placeholder="search, e.g. acema.com"
			class="h-10 pl-9"
			autocomplete="off"
			autocapitalize="off"
			spellcheck={false}
		/>
	</div>
</form>
