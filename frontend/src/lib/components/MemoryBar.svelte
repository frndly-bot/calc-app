<script lang="ts">
	import { formatNumber } from '#lib/eval/normal';
	import { memory } from '#lib/stores/memory.svelte';
	import Key from '#lib/components/Key.svelte';

	let {
		current = 0,
		onrecall
	}: { current?: number; onrecall?: (value: number) => void } = $props();

	const label = $derived(memory.hasValue ? `M ${formatNumber(memory.value, { significant: 8 })}` : '');
</script>

<div class="memory">
	<Key label="MC" variant="memory" ariaLabel="Memory clear" onpress={() => memory.clear()} />
	<Key
		label="MR"
		variant="memory"
		ariaLabel="Memory recall"
		onpress={() => onrecall?.(memory.recall())}
	/>
	<Key
		label="M+"
		variant="memory"
		ariaLabel="Add to memory"
		onpress={() => memory.add(current)}
	/>
	<Key
		label="M−"
		variant="memory"
		ariaLabel="Subtract from memory"
		onpress={() => memory.subtract(current)}
	/>
	<span class="value" class:on={memory.hasValue}>{label}</span>
</div>

<style>
	.memory {
		display: grid;
		grid-template-columns: repeat(4, 54px) minmax(0, 1fr);
		align-items: center;
		gap: 2px;
		padding: 0 3px;
	}

	.value {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-align: right;
		color: transparent;
		font-family: var(--mono);
		font-size: 10px;
		letter-spacing: 0.04em;
		transition: color 0.2s ease;
	}

	.value.on {
		color: color-mix(in srgb, var(--accent) 80%, var(--muted));
	}
</style>
