<script lang="ts">
	import CopyIcon from '@lucide/svelte/icons/copy';
	import CheckIcon from '@lucide/svelte/icons/check';
	import StarIcon from '@lucide/svelte/icons/star';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { cn } from '$lib/utils.js';
	import { rowAt, type Dataset } from '$lib/search';

	const ROW_HEIGHT = 40;
	const OVERSCAN = 6;

	let {
		indices,
		dataset,
		starred = [],
		ontogglestar
	}: {
		indices: ArrayLike<number>;
		dataset: Dataset;
		starred?: string[];
		ontogglestar?: (name: string) => void;
	} = $props();

	let copied = $state('');
	let scrollTop = $state(0);
	let viewportH = $state(0);

	const count = $derived(indices.length);
	const start = $derived(Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN));
	const end = $derived(Math.min(count, Math.ceil((scrollTop + viewportH) / ROW_HEIGHT) + OVERSCAN));
	const visible = $derived(
		Array.from({ length: Math.max(0, end - start) }, (_, k) => rowAt(dataset, indices[start + k]))
	);

	async function copy(name: string) {
		try {
			await navigator.clipboard.writeText(name + '.com');
			copied = name;
			setTimeout(() => {
				if (copied === name) copied = '';
			}, 1200);
		} catch {
			copied = '';
		}
	}
</script>

<div
	class="max-h-[60vh] overflow-auto rounded-lg border"
	bind:clientHeight={viewportH}
	onscroll={(e) => (scrollTop = (e.currentTarget as HTMLElement).scrollTop)}
>
	<div style="height: {count * ROW_HEIGHT}px; position: relative;">
		<ul class="absolute inset-x-0" style="top: {start * ROW_HEIGHT}px;">
			{#each visible as row (row.name)}
				{@const isStarred = starred.includes(row.name)}
				<li
					class="flex items-center justify-between gap-3 border-b px-3 transition-colors hover:bg-muted"
					style="height: {ROW_HEIGHT}px"
				>
					<button
						class="flex min-w-0 flex-1 items-center gap-2 text-left"
						onclick={() => copy(row.name)}
						title="Copy {row.name}.com"
					>
						<span class="font-mono text-sm"
							>{row.name}<span class="text-muted-foreground">.com</span></span
						>
						{#if copied === row.name}
							<CheckIcon class="size-3.5 text-muted-foreground" />
						{:else}
							<CopyIcon class="size-3.5 text-muted-foreground" />
						{/if}
					</button>
					<div class="flex items-center gap-2">
						<Badge variant={row.status === 'available' ? 'default' : 'secondary'}
							>{row.status}</Badge
						>
						<button
							class={cn(
								'transition-colors',
								isStarred ? 'text-amber-500' : 'text-muted-foreground hover:text-foreground'
							)}
							onclick={() => ontogglestar?.(row.name)}
							aria-pressed={isStarred}
							aria-label={isStarred ? `Unstar ${row.name}` : `Star ${row.name}`}
							title={isStarred ? 'Unstar' : 'Star'}
						>
							<StarIcon class="size-4" fill={isStarred ? 'currentColor' : 'none'} />
						</button>
					</div>
				</li>
			{/each}
		</ul>
	</div>
</div>
