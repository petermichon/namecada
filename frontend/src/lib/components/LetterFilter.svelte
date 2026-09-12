<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import type { LetterMode } from '$lib/search';

	const ALPHABET = 'aceimnorsuvwxz'.split('');

	let {
		letters = $bindable<string[]>([]),
		mode = $bindable<LetterMode>('include')
	}: {
		letters: string[];
		mode: LetterMode;
	} = $props();

	function toggle(letter: string) {
		letters = letters.includes(letter) ? letters.filter((l) => l !== letter) : [...letters, letter];
	}
</script>

<div class="mt-3 flex flex-wrap items-center gap-x-2 gap-y-1.5">
	<span class="text-xs text-muted-foreground">Letters</span>
	<div class="flex rounded-lg border p-0.5" role="group" aria-label="Letter filter mode">
		<Button
			variant={mode === 'include' ? 'secondary' : 'ghost'}
			size="xs"
			aria-pressed={mode === 'include'}
			onclick={() => (mode = 'include')}>must contain</Button
		>
		<Button
			variant={mode === 'exclude' ? 'secondary' : 'ghost'}
			size="xs"
			aria-pressed={mode === 'exclude'}
			onclick={() => (mode = 'exclude')}>exclude</Button
		>
	</div>
	<div class="flex flex-wrap gap-1">
		{#each ALPHABET as letter (letter)}
			{@const selected = letters.includes(letter)}
			<button
				type="button"
				class={cn(
					'h-7 w-7 rounded-md border font-mono text-sm transition-colors',
					selected
						? 'border-primary bg-primary text-primary-foreground'
						: 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
				)}
				aria-pressed={selected}
				aria-label="Toggle letter {letter}"
				onclick={() => toggle(letter)}>{letter}</button
			>
		{/each}
	</div>
	{#if letters.length > 0}
		<Button variant="ghost" size="xs" onclick={() => (letters = [])}>Clear</Button>
	{/if}
</div>
