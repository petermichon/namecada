<script lang="ts">
	import {
		extractDetails,
		isNotableStatus,
		nameserverHint,
		type DomainDetails
	} from '$lib/details';
	import { cn } from '$lib/utils.js';

	let { raw }: { raw: unknown } = $props();

	const details = $derived<DomainDetails | null>(extractDetails(raw));
	const nsHint = $derived(nameserverHint(details?.nameservers ?? []));

	const DATE_LABELS: Record<string, string> = {
		registration: 'Registered',
		expiration: 'Expires',
		'last changed': 'Last changed',
		transfer: 'Transferred',
		'last update of RDAP database': 'RDAP updated'
	};

	function formatDate(iso: string): string {
		const date = new Date(iso);
		return Number.isNaN(date.getTime()) ? iso : date.toISOString().slice(0, 10);
	}
</script>

{#if details}
	<div class="mt-4 rounded-lg border p-4">
		<p class="mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
			Registry record
		</p>
		<dl class="grid gap-2 text-sm">
			{#if details.statuses.length}
				<div class="flex gap-3">
					<dt class="w-32 shrink-0 text-muted-foreground">Status</dt>
					<dd class="flex flex-wrap gap-1">
						{#each details.statuses as status (status)}
							<span
								class={cn(
									'rounded-sm border px-1.5 py-0.5 font-mono text-xs',
									isNotableStatus(status)
										? 'border-foreground/60 font-semibold text-foreground'
										: 'border-border text-muted-foreground opacity-70'
								)}>{status}</span
							>
						{/each}
					</dd>
				</div>
			{/if}
			{#if details.events.length}
				<div class="flex gap-3">
					<dt class="w-32 shrink-0 text-muted-foreground">Events</dt>
					<dd class="grid gap-0.5">
						{#each details.events as event (event.action)}
							<span>
								{DATE_LABELS[event.action] ?? event.action}
								<span class="text-muted-foreground">· {formatDate(event.date)}</span>
							</span>
						{/each}
					</dd>
				</div>
			{/if}
			{#if details.registrar}
				<div class="flex gap-3">
					<dt class="w-32 shrink-0 text-muted-foreground">Registrar</dt>
					<dd>
						{details.registrar.name ?? 'Unknown'}
						{#if details.registrar.id}
							<span class="text-muted-foreground">(IANA {details.registrar.id})</span>
						{/if}
						{#if details.registrar.abuseEmail}
							<div class="text-xs text-muted-foreground">{details.registrar.abuseEmail}</div>
						{/if}
					</dd>
				</div>
			{/if}
			{#if details.nameservers.length || nsHint}
				<div class="flex gap-3">
					<dt class="w-32 shrink-0 text-muted-foreground">Nameservers</dt>
					<dd class="min-w-0">
						{#if details.nameservers.length}
							<div class="grid gap-0.5 font-mono text-xs">
								{#each details.nameservers as ns (ns)}
									<span class="break-all">{ns}</span>
								{/each}
							</div>
						{/if}
						{#if nsHint?.kind === 'none'}
							<div class="text-xs text-muted-foreground">None — not delegated</div>
						{:else if nsHint?.kind === 'parked'}
							<div class="text-xs font-medium">
								Parking nameservers{nsHint.provider ? ` (${nsHint.provider})` : ''} — possibly for sale
							</div>
						{/if}
					</dd>
				</div>
			{/if}
			{#if details.dnssec !== undefined}
				<div class="flex gap-3">
					<dt class="w-32 shrink-0 text-muted-foreground">DNSSEC</dt>
					<dd>{details.dnssec ? 'signed' : 'unsigned'}</dd>
				</div>
			{/if}
			{#if details.related}
				<div class="flex gap-3">
					<dt class="w-32 shrink-0 text-muted-foreground">Registrar RDAP</dt>
					<dd class="min-w-0">
						<a
							class="text-xs break-all underline underline-offset-2"
							href={details.related}
							target="_blank"
							rel="noreferrer">{details.related}</a
						>
					</dd>
				</div>
			{/if}
		</dl>
	</div>
{/if}
