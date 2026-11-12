<script lang="ts">
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import SearchIcon from '@lucide/svelte/icons/search';
	import XIcon from '@lucide/svelte/icons/x';
	import * as InputGroup from '$lib/components/ui/input-group/index.js';
	import { autofocus } from '$lib/attachments';

	let {
		value = $bindable(''),
		loading = false,
		invalid = false,
		errorId,
		hintId,
		onsubmit,
		oninput
	}: {
		value: string;
		loading?: boolean;
		invalid?: boolean;
		errorId?: string;
		hintId?: string;
		onsubmit?: () => void;
		oninput?: () => void;
	} = $props();

	const describedBy = $derived(
		[hintId, invalid && errorId ? errorId : undefined].filter(Boolean).join(' ') || undefined
	);
</script>

<form
	role="search"
	onsubmit={(e) => {
		e.preventDefault();
		onsubmit?.();
	}}
>
	<InputGroup.Root class="h-11 rounded-xl border-border bg-card shadow-xs">
		<InputGroup.Addon class="pl-3">
			<SearchIcon class="size-4" aria-hidden="true" />
		</InputGroup.Addon>
		<label for="domain" class="sr-only">Name</label>
		<InputGroup.Input
			id="domain"
			name="domain"
			bind:value
			class="h-full text-base md:text-sm"
			{@attach autofocus}
			oninput={() => oninput?.()}
			onkeydown={(e) => {
				if (e.key === 'Escape') value = '';
			}}
			placeholder="Search a name…"
			autocomplete="off"
			autocapitalize="off"
			spellcheck={false}
			aria-invalid={invalid ? 'true' : undefined}
			aria-describedby={describedBy}
			aria-keyshortcuts="/"
		/>
		<InputGroup.Addon align="inline-end">
			{#if value}
				<InputGroup.Button
					size="icon-sm"
					type="button"
					class="size-10 rounded-lg sm:size-8"
					onmousedown={(e) => e.preventDefault()}
					onclick={() => (value = '')}
					aria-label="Clear name"
				>
					<XIcon class="size-3.5" />
				</InputGroup.Button>
			{/if}
			<InputGroup.Button
				variant="default"
				size="icon-sm"
				class="size-10 rounded-lg sm:size-9"
				type="submit"
				aria-label={loading ? 'Check name again' : 'Check name'}
			>
				<ArrowRightIcon class="size-4 text-white dark:text-primary-foreground" />
			</InputGroup.Button>
		</InputGroup.Addon>
	</InputGroup.Root>
</form>
