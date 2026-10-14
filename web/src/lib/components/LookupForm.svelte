<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import * as InputGroup from '$lib/components/ui/input-group/index.js';
	import { autofocus } from '$lib/attachments';

	let {
		value = $bindable(''),
		loading = false,
		onsubmit
	}: {
		value: string;
		loading?: boolean;
		onsubmit?: () => void;
	} = $props();
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		onsubmit?.();
	}}
>
	<InputGroup.Root>
		<InputGroup.Addon>
			<SearchIcon />
		</InputGroup.Addon>
		<label for="domain" class="sr-only">Name</label>
		<InputGroup.Input
			id="domain"
			name="domain"
			bind:value
			{@attach autofocus}
			onkeydown={(e) => {
				if (e.key === 'Escape') value = '';
			}}
			placeholder="Search a name…"
			autocomplete="off"
			autocapitalize="off"
			spellcheck={false}
		/>
		<InputGroup.Addon align="inline-end">
			<kbd data-slot="kbd" class="hidden font-sans text-[11px] font-medium sm:inline-flex"
				>Enter</kbd
			>
			<InputGroup.Button size="icon-sm" type="submit" disabled={loading} aria-label="Check name">
				{#if loading}
					<LoaderCircleIcon class="size-4 animate-spin motion-reduce:animate-none" />
				{:else}
					<ArrowRightIcon class="size-4" />
				{/if}
			</InputGroup.Button>
		</InputGroup.Addon>
	</InputGroup.Root>
</form>
