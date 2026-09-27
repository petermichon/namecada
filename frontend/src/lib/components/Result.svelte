<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import HelpCircleIcon from '@lucide/svelte/icons/help-circle';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { LookupResult } from '$lib/rdap';

	let { result }: { result: LookupResult } = $props();

	const view = $derived.by(() => {
		switch (result.status) {
			case 'unregistered':
				return {
					label: 'not registered',
					variant: 'default' as const,
					note: 'No registry record — this .com is unregistered.',
					Icon: CheckCircle2Icon
				};
			case 'registered':
				return {
					label: 'registered',
					variant: 'secondary' as const,
					note: 'A registry record exists — this .com is taken.',
					Icon: XCircleIcon
				};
			default:
				return {
					label: 'unknown',
					variant: 'outline' as const,
					note: 'Registry did not return a definitive answer.',
					Icon: HelpCircleIcon
				};
		}
	});
</script>

<div class="mt-6 rounded-lg border p-4">
	<div class="flex items-center justify-between gap-3">
		<div class="flex min-w-0 items-center gap-2">
			<view.Icon class="size-5 shrink-0 text-muted-foreground" />
			<span class="truncate font-mono text-lg"
				>{result.domain.replace(/\.com$/, '')}<span class="text-muted-foreground">.com</span></span
			>
		</div>
		<Badge variant={view.variant}>{view.label}</Badge>
	</div>
	<p class="mt-2 text-sm text-muted-foreground">{view.note}</p>
</div>
