<script lang="ts">
	let {
		name,
		size = 18,
		stroke = 1.7
	}: { name: string; size?: number; stroke?: number } = $props();
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width={stroke}
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	{#if name === 'sun'}
		<circle cx="12" cy="12" r="4.2" />
		<path d="M12 2.5v2M12 19.5v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M2.5 12h2M19.5 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4" />
	{:else if name === 'moon'}
		<path d="M20.6 14.6A8.6 8.6 0 0 1 9.4 3.4a8.6 8.6 0 1 0 11.2 11.2Z" />
	{:else if name === 'history'}
		<path d="M3.5 12a8.5 8.5 0 1 0 2.6-6.1L3.5 8.5" />
		<path d="M3.5 3.5v5h5" />
		<path d="M12 7.5V12l3 1.8" />
	{:else if name === 'close'}
		<path d="M6 6l12 12M18 6 6 18" />
	{:else if name === 'trash'}
		<path d="M4.5 7h15M9.5 7V4.8h5V7M6.5 7l.9 12.2h9.2L17.5 7M10 11v4.6M14 11v4.6" />
	{/if}
</svg>
