<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		columns = 4,
		children,
		gap = 8
	}: { columns?: number; children: Snippet; gap?: number } = $props();
</script>

<div class="keypad" style:--cols={columns} style:--gap="{gap}px">
	{@render children()}
</div>

<style>
	.keypad {
		display: grid;
		grid-template-columns: repeat(var(--cols), minmax(0, 1fr));
		gap: var(--gap);
		align-content: start;
	}
</style>
