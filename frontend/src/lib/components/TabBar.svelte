<script lang="ts">
	import { mode, modes } from '#lib/stores/mode.svelte';
</script>

<nav class="tabs glass" aria-label="Calculator mode">
	{#each modes as item, index (item.id)}
		<button
			type="button"
			class="tab"
			class:active={mode.current === item.id}
			onclick={() => mode.set(item.id)}
			aria-pressed={mode.current === item.id}
			title={item.hint}
		>
			<span class="dot" style:background={item.accent}></span>
			<span class="label">{item.label}</span>
		</button>
	{/each}
	<span
		class="indicator"
		style:transform="translateX({mode.index * 100}%)"
		style:background={modes[mode.index]?.accent}
	></span>
</nav>

<style>
	.tabs {
		position: relative;
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		padding: 4px;
		border-radius: 16px;
		overflow: hidden;
		flex-shrink: 0;
	}

	.tab {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		height: 36px;
		border: 0;
		border-radius: 12px;
		background: transparent;
		color: var(--muted);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.01em;
		transition:
			color 0.2s ease,
			background 0.2s ease;
	}

	.tab:hover {
		color: var(--text);
		background: var(--surface);
	}

	.tab.active {
		color: var(--text);
		background: color-mix(in srgb, var(--accent) 13%, transparent);
	}

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		opacity: 0.45;
		transition:
			opacity 0.2s ease,
			box-shadow 0.2s ease;
	}

	.tab.active .dot {
		opacity: 1;
		box-shadow: 0 0 10px currentColor;
	}

	.indicator {
		position: absolute;
		left: 4px;
		bottom: 3px;
		width: calc((100% - 8px) / 4);
		height: 2px;
		border-radius: 999px;
		transition:
			transform 0.32s cubic-bezier(0.22, 1, 0.36, 1),
			background 0.3s ease;
		box-shadow: 0 0 14px currentColor;
	}

	@media (max-width: 380px) {
		.label {
			display: none;
		}
	}
</style>
