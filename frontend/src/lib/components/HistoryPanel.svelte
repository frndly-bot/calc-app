<script lang="ts">
	import { flip } from 'svelte/animate';
	import { fade, fly } from 'svelte/transition';
	import { accentFor, modeLabel } from '#lib/stores/mode.svelte';
	import { history } from '#lib/stores/history.svelte';
	import { motion } from '#lib/stores/motion.svelte';
	import Icon from '#lib/components/Icon.svelte';

	let { open = false, onclose }: { open?: boolean; onclose: () => void } = $props();

	function relative(iso: string): string {
		const time = new Date(iso).getTime();
		if (!Number.isFinite(time)) return '';
		const diff = Date.now() - time;
		const minute = 60_000;
		if (diff < minute) return 'just now';
		if (diff < 60 * minute) return `${Math.round(diff / minute)}m ago`;
		if (diff < 24 * 60 * minute) return `${Math.round(diff / (60 * minute))}h ago`;
		if (diff < 7 * 24 * 60 * minute) return `${Math.round(diff / (24 * 60 * minute))}d ago`;
		return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function focusOnMount(node: HTMLElement) {
		node.focus();
	}
</script>

{#if open}
	<button
		type="button"
		class="scrim"
		aria-label="Close history"
		transition:fade={{ duration: motion.reduced ? 0 : 180 }}
		onclick={onclose}
	></button>

	<div
		class="panel"
		role="dialog"
		aria-modal="true"
		aria-label="Calculation history"
		tabindex="-1"
		use:focusOnMount
		transition:fly={{
			x: motion.reduced ? 0 : 330,
			duration: motion.reduced ? 0 : 300,
			opacity: motion.reduced ? 1 : 0.35
		}}
	>
		<header class="head">
			<div>
				<h2>History</h2>
				<p>{history.entries.length} saved {history.entries.length === 1 ? 'entry' : 'entries'}</p>
			</div>
			<div class="tools">
				<button
					type="button"
					class="clear"
					disabled={history.entries.length === 0}
					onclick={() => void history.clear()}
				>
					Clear all
				</button>
				<button type="button" class="icon-button" aria-label="Close history" onclick={onclose}>
					<Icon name="close" size={16} />
				</button>
			</div>
		</header>

		<div class="list">
			{#if history.entries.length === 0}
				<div class="empty" in:fade={{ duration: motion.reduced ? 0 : 150 }}>
					<Icon name="history" size={26} />
					<p>No calculations yet</p>
					<span>Results you confirm with = are stored here.</span>
				</div>
			{:else}
				{#each history.entries as entry (entry.id)}
					<article
						class="row"
						animate:flip={{ duration: motion.reduced ? 0 : 220 }}
						out:fly={{
							x: motion.reduced ? 0 : 34,
							duration: motion.reduced ? 0 : 170,
							opacity: motion.reduced ? 1 : 0
						}}
					>
						<span class="dot" style:background={accentFor(entry.mode)}></span>
						<div class="body">
							<p class="expr">{entry.expr}</p>
							<p class="result">= {entry.result}</p>
							<p class="meta">
								{modeLabel(entry.mode)}
								<span aria-hidden="true">·</span>
								{relative(entry.createdAt)}
							</p>
						</div>
						<button
							type="button"
							class="remove"
							aria-label="Delete {entry.expr}"
							onclick={() => void history.remove(entry.id)}
						>
							<Icon name="trash" size={14} />
						</button>
					</article>
				{/each}
			{/if}
		</div>
	</div>
{/if}

<style>
	.scrim {
		position: fixed;
		inset: 0;
		z-index: 20;
		border: 0;
		padding: 0;
		background: rgba(3, 6, 14, 0.55);
		backdrop-filter: blur(3px);
		cursor: default;
	}

	.panel {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		z-index: 21;
		display: flex;
		flex-direction: column;
		width: min(340px, 88vw);
		padding: 18px 16px 14px;
		background: color-mix(in srgb, var(--bg-1) 88%, var(--accent) 4%);
		border-left: 1px solid var(--border-strong);
		box-shadow: -24px 0 70px rgba(0, 0, 0, 0.45);
		outline: none;
	}

	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
		padding-bottom: 14px;
		border-bottom: 1px solid var(--border);
	}

	h2 {
		margin: 0;
		font-size: 17px;
		font-weight: 650;
		letter-spacing: -0.01em;
	}

	.head p {
		margin: 3px 0 0;
		color: var(--muted);
		font-size: 11px;
	}

	.tools {
		display: flex;
		align-items: center;
		gap: 7px;
	}

	.clear {
		height: 32px;
		padding: 0 11px;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: transparent;
		color: var(--muted);
		font-size: 11px;
		font-weight: 650;
		transition:
			color 0.16s ease,
			border-color 0.16s ease,
			background 0.16s ease;
	}

	.clear:hover:not(:disabled) {
		color: var(--danger);
		border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
		background: color-mix(in srgb, var(--danger) 10%, transparent);
	}

	.clear:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 12px 2px 8px;
		overflow-y: auto;
		flex: 1;
	}

	.row {
		display: grid;
		grid-template-columns: 7px minmax(0, 1fr) 28px;
		align-items: start;
		gap: 9px;
		padding: 10px 9px 10px 11px;
		border: 1px solid transparent;
		border-radius: 13px;
		background: var(--surface);
		transition:
			border-color 0.16s ease,
			background 0.16s ease;
	}

	.row:hover {
		background: var(--surface-strong);
		border-color: var(--border);
	}

	.dot {
		width: 7px;
		height: 7px;
		margin-top: 5px;
		border-radius: 50%;
		box-shadow: 0 0 9px currentColor;
	}

	.body {
		min-width: 0;
	}

	.expr {
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--muted);
		font-family: var(--mono);
		font-size: 11.5px;
	}

	.result {
		margin: 2px 0 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-family: var(--mono);
		font-size: 14.5px;
		font-weight: 600;
	}

	.meta {
		margin: 5px 0 0;
		color: var(--muted);
		font-size: 10px;
		letter-spacing: 0.02em;
	}

	.remove {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border: 0;
		border-radius: 9px;
		background: transparent;
		color: var(--muted);
		opacity: 0;
		transition:
			opacity 0.16s ease,
			color 0.16s ease,
			background 0.16s ease;
	}

	.row:hover .remove,
	.remove:focus-visible {
		opacity: 1;
	}

	.remove:hover {
		color: var(--danger);
		background: color-mix(in srgb, var(--danger) 13%, transparent);
	}

	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		margin: auto;
		padding: 30px 18px;
		color: var(--muted);
		text-align: center;
	}

	.empty p {
		margin: 6px 0 0;
		color: var(--text);
		font-size: 14px;
		font-weight: 600;
	}

	.empty span {
		font-size: 11.5px;
		line-height: 1.5;
		max-width: 200px;
	}

	@media (hover: none) {
		.remove {
			opacity: 0.7;
		}
	}
</style>
