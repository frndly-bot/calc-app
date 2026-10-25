<script lang="ts">
	let {
		label,
		onpress,
		variant = 'digit',
		span = 1,
		disabled = false,
		active = false,
		ariaLabel,
		title
	}: {
		label: string;
		onpress: () => void;
		variant?: 'digit' | 'operator' | 'equals' | 'function' | 'danger' | 'memory';
		span?: number;
		disabled?: boolean;
		active?: boolean;
		ariaLabel?: string;
		title?: string;
	} = $props();
</script>

<button
	type="button"
	class="key {variant}"
	class:active
	style:grid-column={span > 1 ? `span ${span}` : null}
	onclick={onpress}
	{disabled}
	aria-label={ariaLabel ?? label}
	aria-pressed={active ? true : undefined}
	{title}
>
	<span>{label}</span>
</button>

<style>
	.key {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 46px;
		border: 1px solid var(--border);
		border-radius: 15px;
		background: var(--surface);
		color: var(--text);
		font-size: 19px;
		font-weight: 500;
		box-shadow: var(--shadow-key);
		overflow: hidden;
		user-select: none;
		-webkit-tap-highlight-color: transparent;
		transition:
			transform 0.1s ease,
			background 0.16s ease,
			border-color 0.16s ease,
			box-shadow 0.16s ease,
			color 0.16s ease;
	}

	.key::after {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			circle at center,
			color-mix(in srgb, var(--accent) 45%, transparent),
			transparent 65%
		);
		opacity: 0;
		transition: opacity 0.35s ease;
		pointer-events: none;
	}

	.key:hover {
		background: var(--surface-hover);
		border-color: var(--border-strong);
	}

	.key:active {
		transform: scale(0.955);
	}

	.key:active::after {
		opacity: 0.55;
		transition: none;
	}

	.key:disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}

	.key.active {
		background: color-mix(in srgb, var(--accent) 24%, var(--surface));
		border-color: color-mix(in srgb, var(--accent) 55%, transparent);
		color: var(--text);
	}

	.key.operator {
		background: color-mix(in srgb, var(--accent) 15%, var(--surface));
		border-color: color-mix(in srgb, var(--accent) 30%, var(--border));
		font-size: 21px;
	}

	.key.operator:hover {
		background: color-mix(in srgb, var(--accent) 24%, var(--surface));
	}

	.key.equals {
		background: linear-gradient(
			140deg,
			color-mix(in srgb, var(--accent) 88%, white 12%),
			var(--accent)
		);
		border-color: transparent;
		color: var(--accent-contrast);
		font-weight: 650;
		box-shadow: 0 10px 28px color-mix(in srgb, var(--accent) 38%, transparent);
	}

	.key.equals:hover {
		filter: brightness(1.06);
	}

	.key.function {
		font-size: 13px;
		font-weight: 650;
		letter-spacing: 0.015em;
		color: var(--muted);
	}

	.key.function:hover {
		color: var(--text);
	}

	.key.danger {
		color: var(--danger);
		background: color-mix(in srgb, var(--danger) 11%, var(--surface));
		border-color: color-mix(in srgb, var(--danger) 24%, var(--border));
		font-size: 15px;
		font-weight: 700;
	}

	.key.memory {
		min-height: 29px;
		border-radius: 999px;
		background: transparent;
		box-shadow: none;
		color: var(--muted);
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.09em;
	}

	.key.equals :global(span),
	.key.operator :global(span) {
		font-variant-numeric: tabular-nums;
	}
</style>
