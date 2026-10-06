<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
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
	<div
		class="flex h-10 items-center gap-2 rounded-full border border-border bg-card pr-1 pl-3.5 shadow-sm focus-within:border-ring focus-within:ring-4 focus-within:ring-ring/10 dark:shadow-none"
	>
		<SearchIcon class="size-4 shrink-0 text-muted-foreground" />
		<label for="domain" class="sr-only">Name</label>
		<Input
			id="domain"
			name="domain"
			bind:value
			{@attach autofocus}
			placeholder="Search a name…"
			class="h-full flex-1 rounded-none border-0 bg-transparent px-0 focus-visible:ring-0 dark:bg-transparent"
			autocomplete="off"
			autocapitalize="off"
			spellcheck={false}
		/>
		<kbd
			class="hidden h-6 shrink-0 items-center rounded-md border border-border bg-muted px-1.5 font-sans text-[11px] font-medium text-muted-foreground sm:flex"
			>Enter</kbd
		>
		<Button
			type="submit"
			size="icon"
			class="size-8 shrink-0 rounded-full"
			disabled={loading}
			aria-label="Check name"
		>
			{#if loading}
				<LoaderCircleIcon class="size-4 animate-spin" />
			{:else}
				<ArrowRightIcon class="size-4" />
			{/if}
		</Button>
	</div>
</form>
