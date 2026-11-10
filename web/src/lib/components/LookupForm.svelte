<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import XIcon from '@lucide/svelte/icons/x';
	import * as InputGroup from '$lib/components/ui/input-group/index.js';
	import { autofocus } from '$lib/attachments';

	let {
		value = $bindable(''),
		loading = false,
		invalid = false,
		errorId,
		onsubmit
	}: {
		value: string;
		loading?: boolean;
		invalid?: boolean;
		errorId?: string;
		onsubmit?: () => void;
	} = $props();
</script>

<form
	role="search"
	onsubmit={(e) => {
		e.preventDefault();
		onsubmit?.();
	}}
>
	<InputGroup.Root class="h-11">
		<InputGroup.Addon>
			<SearchIcon />
		</InputGroup.Addon>
		<label for="domain" class="sr-only">Name</label>
		<InputGroup.Input
			id="domain"
			name="domain"
			bind:value
			class="h-full"
			{@attach autofocus}
			onkeydown={(e) => {
				if (e.key === 'Escape') value = '';
			}}
			placeholder="Search a bare name…"
			autocomplete="off"
			autocapitalize="off"
			spellcheck={false}
			aria-invalid={invalid ? 'true' : undefined}
			aria-describedby={invalid && errorId ? errorId : undefined}
		/>
		<InputGroup.Addon align="inline-end">
			{#if value}
				<InputGroup.Button
					size="icon-sm"
					type="button"
					onmousedown={(e) => e.preventDefault()}
					onclick={() => (value = '')}
					aria-label="Clear name"
				>
					<XIcon class="size-3.5" />
				</InputGroup.Button>
			{/if}
			<InputGroup.Button
				size="icon-sm"
				class="size-10 text-foreground"
				type="submit"
				disabled={loading}
				aria-label="Check name"
			>
				{#if loading}
					<LoaderCircleIcon class="size-4 animate-spin motion-reduce:animate-none" />
				{:else}
					<ArrowRightIcon class="size-4" />
				{/if}
			</InputGroup.Button>
		</InputGroup.Addon>
	</InputGroup.Root>
</form>
