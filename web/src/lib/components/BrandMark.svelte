<script lang="ts">
	// The brand mark: a solid seal with a magnifier knocked out of it. A mask
	// keeps the lens truly transparent, so it works on the header's blurred
	// background and on any surface the mark is placed on.
	let { class: className = '' }: { class?: string } = $props();
</script>

<svg viewBox="0 0 24 24" class={className} aria-hidden="true">
	<defs>
		<mask id="brand-lens">
			<rect width="24" height="24" fill="black" />
			<path
				d="M3.85 8.62a4 4 0 0 1 4.78-4.77 4 4 0 0 1 6.74 0 4 4 0 0 1 4.78 4.78 4 4 0 0 1 0 6.74 4 4 0 0 1-4.77 4.78 4 4 0 0 1-6.75 0 4 4 0 0 1-4.78-4.77 4 4 0 0 1 0-6.76Z"
				fill="white"
			/>
			<g
				transform="translate(6.5 6.5) scale(0.5)"
				fill="none"
				stroke="black"
				stroke-width="2.8"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<circle cx="11" cy="11" r="8" />
				<path d="m21 21-4.34-4.34" />
			</g>
		</mask>
	</defs>
	<rect width="24" height="24" fill="currentColor" mask="url(#brand-lens)" />
</svg>
