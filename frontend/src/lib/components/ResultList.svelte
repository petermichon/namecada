<script lang="ts">
	import CopyIcon from '@lucide/svelte/icons/copy';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { Row } from '$lib/search';

	let { rows }: { rows: Row[] } = $props();

	let copied = $state('');

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

<ul class="divide-y overflow-hidden rounded-lg border">
	{#each rows as row (row.name)}
		<li>
			<button
				class="flex w-full items-center justify-between px-3 py-2 text-left transition-colors hover:bg-muted"
				onclick={() => copy(row.name)}
				title="Copy {row.name}.com"
			>
				<span class="font-mono text-sm"
					>{row.name}<span class="text-muted-foreground">.com</span></span
				>
				<span class="flex items-center gap-2">
					<Badge variant={row.status === 'available' ? 'default' : 'secondary'}>{row.status}</Badge>
					{#if copied === row.name}
						<CheckIcon class="size-4 text-muted-foreground" />
					{:else}
						<CopyIcon class="size-4 text-muted-foreground" />
					{/if}
				</span>
			</button>
		</li>
	{/each}
</ul>
