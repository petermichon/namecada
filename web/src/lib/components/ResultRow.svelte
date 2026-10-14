<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import CircleMinusIcon from '@lucide/svelte/icons/circle-minus';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import type { Row } from '$lib/useLookup.svelte';

	const STATUS: Record<
		Row['status'],
		{ text: string; cls: string; Icon?: typeof CheckCircle2Icon }
	> = {
		available: { text: 'available', cls: 'text-success', Icon: CheckCircle2Icon },
		taken: { text: 'unavailable', cls: 'text-muted-foreground', Icon: XCircleIcon },
		unknown: { text: 'unknown', cls: 'text-warning', Icon: HelpCircleIcon },
		unsupported: { text: 'unsupported', cls: 'text-muted-foreground' },
		pending: { text: 'checking', cls: 'text-muted-foreground' }
	};

	let { row }: { row: Row } = $props();

	const pending = $derived(row.status === 'pending');
	const status = $derived(STATUS[row.status]);
	const text = $derived(row.url.replace(/^https?:\/\//, ''));
	const Icon = $derived(status.Icon ?? CheckCircle2Icon);
</script>

<li class="bg-card">
	{#if row.status === 'unsupported'}
		<div
			class="flex items-center gap-3 px-3 py-3 opacity-60"
			aria-disabled="true"
			title={`${row.platformName} is not checked`}
		>
			<span class="w-28 shrink-0 truncate text-xs font-medium text-muted-foreground">
				{row.platformName}
			</span>
			<span class="min-w-0 flex-1 truncate text-sm text-muted-foreground">{row.detail}</span>
			<span class="flex shrink-0 items-center gap-1.5 text-xs text-muted-foreground">
				<CircleMinusIcon class="size-3.5 shrink-0" />
				<span class="w-20">unsupported</span>
			</span>
		</div>
	{:else}
		<a
			href={pending ? undefined : row.url}
			target="_blank"
			rel="noopener noreferrer"
			class="flex items-center gap-3 px-3 py-3 {row.status === 'available'
				? 'bg-success/[0.07] hover:bg-success/[0.12]'
				: 'hover:bg-muted/50'} focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none focus-visible:ring-inset"
		>
			<span
				class="hidden w-28 shrink-0 truncate text-xs font-medium text-muted-foreground sm:block"
				title={row.platformName}>{row.platformName}</span
			>
			{#if pending}
				<Skeleton class="h-5 w-40 max-w-full min-w-0 flex-1 rounded motion-reduce:animate-none" />
			{:else}
				<span
					class="min-w-0 flex-1 truncate text-sm {row.status === 'taken'
						? 'text-muted-foreground'
						: ''}">{text}</span
				>
			{/if}
			{#if row.detail}
				<span class="hidden shrink-0 text-xs text-muted-foreground sm:inline">{row.detail}</span>
			{/if}
			{#if pending}
				<span class="flex shrink-0 items-center text-xs">
					<Skeleton class="h-3 w-14 shrink-0 rounded-full motion-reduce:animate-none" />
				</span>
			{:else}
				<span class="flex shrink-0 items-center gap-1.5 text-xs {status.cls}">
					<Icon class="size-3.5 shrink-0" />
					<span class="w-20">{status.text}</span>
				</span>
			{/if}
		</a>
	{/if}
</li>
