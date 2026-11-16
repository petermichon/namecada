<script lang="ts">
	import type { Snippet } from 'svelte';
	import BrandMark from './BrandMark.svelte';

	// One header for the landing page and the report. The search field turns it
	// into the two-row mobile layout; actions sit beside the brand on phones and
	// at the far right on wider screens.
	let {
		sticky = false,
		onhome,
		search,
		actions
	}: {
		sticky?: boolean;
		onhome: (event: MouseEvent) => void;
		search?: Snippet;
		actions?: Snippet;
	} = $props();
</script>

<header
	class="z-20 w-full border-b {sticky
		? 'sticky top-0 border-border/70 bg-background'
		: 'border-transparent'}"
>
	<div
		class="mx-auto flex w-full max-w-6xl flex-col gap-2 px-4 pt-[calc(0.625rem+env(safe-area-inset-top))] pb-2.5 sm:h-16 sm:flex-row sm:items-center sm:gap-3 sm:px-6"
	>
		<div class="flex items-center justify-between gap-3 sm:justify-start">
			<a
				href="./"
				onclick={onhome}
				class="inline-flex items-center gap-2 rounded-lg transition-opacity hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring"
			>
				<BrandMark class="size-7 text-primary" />
				<span class="text-lg font-semibold tracking-[-0.02em]">Namecada</span>
			</a>
			{#if actions}
				<div class="flex items-center gap-1 sm:hidden">{@render actions()}</div>
			{/if}
		</div>

		{#if search}
			<div class="min-w-0 flex-1">
				<div class="mx-auto w-full max-w-xl">{@render search()}</div>
			</div>
		{/if}

		{#if actions}
			<div class="hidden items-center gap-1 sm:flex">{@render actions()}</div>
		{/if}
	</div>
</header>
