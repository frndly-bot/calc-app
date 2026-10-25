<script lang="ts">
	import { formatNumber } from '#lib/eval/normal';
	import { motion } from '#lib/stores/motion.svelte';

	let {
		expression = '',
		result = '0',
		error = '',
		badge = null,
		depth = 0,
		pulse = 0,
		value = null
	}: {
		expression?: string;
		result?: string;
		error?: string;
		badge?: string | null;
		depth?: number;
		pulse?: number;
		value?: number | null;
	} = $props();

	let animated = $state<string | null>(null);
	let frame = 0;
	let mounted = false;
	let lastPulse = 0;
	let lastValue = 0;

	$effect(() => {
		const nextPulse = pulse;
		const target = value;

		if (!mounted) {
			mounted = true;
			lastPulse = nextPulse;
			lastValue = target ?? 0;
			return;
		}

		if (nextPulse === lastPulse) return;
		lastPulse = nextPulse;
		cancelAnimationFrame(frame);

		if (motion.reduced || target === null || !Number.isFinite(target)) {
			animated = null;
			lastValue = target ?? 0;
			return;
		}

		const from = lastValue;
		lastValue = target;
		const start = performance.now();
		const duration = 460;

		const tick = (now: number) => {
			const progress = Math.min(1, (now - start) / duration);
			const eased = 1 - Math.pow(1 - progress, 3);
			animated = formatNumber(from + (target - from) * eased);
			if (progress < 1) {
				frame = requestAnimationFrame(tick);
			} else {
				animated = null;
			}
		};

		frame = requestAnimationFrame(tick);
		return () => cancelAnimationFrame(frame);
	});

	const shown = $derived(animated ?? (error ? 'Error' : result));
	const size = $derived.by(() => {
		const length = shown.length;
		if (length > 18) return 'tiny';
		if (length > 13) return 'small';
		if (length > 10) return 'medium';
		return 'large';
	});
</script>

<section class="display glass" aria-live="polite" aria-atomic="true">
	<div class="meta">
		{#if badge}<span class="badge">{badge}</span>{/if}
		{#if depth > 0}
			<span
				class="depth"
				title="{depth} unclosed {depth === 1 ? 'parenthesis' : 'parentheses'}"
				aria-label="{depth} unclosed {depth === 1 ? 'parenthesis' : 'parentheses'}"
			>( {depth}</span>
		{/if}
		<span class="expr {error ? 'error-text' : ''}">{error || expression || '\u00a0'}</span>
	</div>
	<div class="result {size}" class:error>{shown}</div>
</section>

<style>
	.display {
		position: relative;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		gap: 4px;
		min-height: 116px;
		padding: 14px 17px 15px;
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		overflow: hidden;
		flex-shrink: 0;
	}

	.display::before {
		content: '';
		position: absolute;
		inset: 0 0 auto;
		height: 1px;
		background: linear-gradient(
			90deg,
			transparent,
			color-mix(in srgb, var(--accent) 70%, transparent),
			transparent
		);
	}

	.meta {
		display: flex;
		align-items: center;
		gap: 7px;
		min-height: 18px;
	}

	.badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 19px;
		height: 18px;
		padding: 0 5px;
		border-radius: 6px;
		background: color-mix(in srgb, var(--accent) 22%, transparent);
		color: var(--accent);
		font-size: 10px;
		font-weight: 800;
		letter-spacing: 0.05em;
		flex-shrink: 0;
	}

	.depth {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		height: 18px;
		padding: 0 5px;
		border: 1px solid color-mix(in srgb, var(--accent) 30%, transparent);
		border-radius: 6px;
		color: color-mix(in srgb, var(--accent) 75%, var(--muted));
		font-family: var(--mono);
		font-size: 10px;
		font-weight: 700;
		flex-shrink: 0;
	}

	.expr {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--muted);
		font-family: var(--mono);
		font-size: 12px;
		letter-spacing: 0.01em;
	}

	.expr.error-text {
		color: var(--danger);
	}

	.result {
		text-align: right;
		font-variant-numeric: tabular-nums;
		font-weight: 600;
		letter-spacing: -0.02em;
		line-height: 1.12;
		overflow-wrap: anywhere;
		transition: color 0.2s ease;
	}

	.large {
		font-size: clamp(34px, 9vw, 44px);
	}

	.medium {
		font-size: clamp(28px, 7.4vw, 36px);
	}

	.small {
		font-size: clamp(22px, 6vw, 29px);
	}

	.tiny {
		font-size: clamp(17px, 4.6vw, 22px);
	}

	.error {
		color: var(--danger);
		font-size: clamp(20px, 5.4vw, 26px);
		font-weight: 600;
	}
</style>
