<script lang="ts">
	import { fly } from 'svelte/transition';
	import { motion } from '#lib/stores/motion.svelte';
	import { toast } from '#lib/stores/toast.svelte';
	import Icon from '#lib/components/Icon.svelte';
</script>

<div class="layer">
	{#if toast.message}
		<div
			class="toast glass"
			role="status"
			transition:fly={{ y: motion.reduced ? 0 : 18, duration: motion.reduced ? 0 : 220 }}
		>
			<span>{toast.message}</span>
			<button type="button" class="dismiss" aria-label="Dismiss" onclick={() => toast.dismiss()}>
				<Icon name="close" size={13} />
			</button>
		</div>
	{/if}
</div>

<style>
	.layer {
		position: fixed;
		left: 0;
		right: 0;
		bottom: 18px;
		z-index: 30;
		display: flex;
		justify-content: center;
		padding: 0 16px;
		pointer-events: none;
	}

	.toast {
		display: flex;
		align-items: center;
		gap: 10px;
		max-width: 400px;
		padding: 9px 9px 9px 14px;
		border-radius: 13px;
		border-color: color-mix(in srgb, var(--danger) 32%, var(--border));
		background: color-mix(in srgb, var(--bg-1) 80%, var(--danger) 8%);
		box-shadow: var(--shadow-lg);
		color: var(--text);
		font-size: 12.5px;
		line-height: 1.4;
		pointer-events: auto;
	}

	.dismiss {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		flex-shrink: 0;
		border: 0;
		border-radius: 8px;
		background: transparent;
		color: var(--muted);
	}

	.dismiss:hover {
		background: var(--surface-hover);
		color: var(--text);
	}
</style>
