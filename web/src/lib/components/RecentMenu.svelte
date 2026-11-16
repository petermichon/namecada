<script lang="ts">
	import HistoryIcon from '@lucide/svelte/icons/history';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { Button } from '$lib/components/ui/button/index.js';

	let {
		labels,
		onpick,
		onclear
	}: { labels: string[]; onpick: (label: string) => void; onclear: () => void } = $props();

	let open = $state(false);
	let root = $state<HTMLDivElement | null>(null);
	let trigger = $state<HTMLElement | null>(null);
	// The header mounts actions twice (phone and desktop), so the panel id has
	// to be unique per instance.
	const menuId = $props.id();

	function close(returnFocus = false) {
		open = false;
		if (returnFocus) trigger?.focus();
	}

	// A pointer anywhere outside the menu dismisses it, like a native popover.
	function onPointerDown(event: PointerEvent) {
		if (!open) return;
		if (root && !root.contains(event.target as Node)) close();
	}

	function onKeydown(event: KeyboardEvent) {
		if (!open || event.key !== 'Escape') return;
		event.stopPropagation();
		close(true);
	}
</script>

<svelte:window onpointerdown={onPointerDown} onkeydown={onKeydown} />

<div class="relative" bind:this={root}>
	<Button
		variant="ghost"
		size="icon"
		class="size-10 rounded-lg text-muted-foreground hover:text-foreground sm:size-9"
		bind:ref={trigger}
		aria-label="Recent checks"
		aria-expanded={open}
		aria-controls={menuId}
		onclick={() => (open = !open)}
	>
		<HistoryIcon class="size-4" />
	</Button>

	{#if open}
		<div
			id={menuId}
			class="absolute top-full right-0 z-30 mt-1 w-60 rounded-lg border border-border bg-card p-1 shadow-lg"
		>
			{#each labels as label (label)}
				<button
					type="button"
					class="flex h-11 w-full items-center rounded-md px-3 text-left text-sm hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
					onclick={() => {
						close(true);
						onpick(label);
					}}>{label}</button
				>
			{:else}
				<p class="px-3 py-2 text-sm text-muted-foreground">No recent checks.</p>
			{/each}

			{#if labels.length}
				<div class="my-1 h-px bg-border"></div>
				<button
					type="button"
					class="flex h-11 w-full items-center gap-2 rounded-md px-3 text-left text-sm text-muted-foreground hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
					onclick={() => {
						close(true);
						onclear();
					}}
				>
					<Trash2Icon class="size-3.5" aria-hidden="true" />
					Clear history
				</button>
			{/if}
		</div>
	{/if}
</div>
