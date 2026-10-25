<script lang="ts">
	let {
		label,
		value = $bindable(0),
		suffix = '',
		step = 'any',
		min,
		placeholder = ''
	}: {
		label: string;
		value?: number;
		suffix?: string;
		step?: string | number;
		min?: number;
		placeholder?: string;
	} = $props();
</script>

<label class="field">
	<span class="label">{label}</span>
	<span class="box">
		<input type="number" bind:value {step} {min} {placeholder} inputmode="decimal" />
		{#if suffix}<span class="suffix">{suffix}</span>{/if}
	</span>
</label>

<style>
	.field {
		display: flex;
		flex-direction: column;
		gap: 5px;
		min-width: 0;
	}

	.label {
		color: var(--muted);
		font-size: 10px;
		font-weight: 650;
		letter-spacing: 0.07em;
		text-transform: uppercase;
	}

	.box {
		display: flex;
		align-items: center;
		gap: 4px;
		height: 42px;
		padding: 0 11px;
		border: 1px solid var(--border);
		border-radius: 12px;
		background: var(--surface);
		transition:
			border-color 0.16s ease,
			background 0.16s ease;
	}

	.box:focus-within {
		border-color: color-mix(in srgb, var(--accent) 55%, var(--border));
		background: var(--surface-strong);
	}

	input {
		width: 100%;
		min-width: 0;
		border: 0;
		background: transparent;
		color: var(--text);
		font-size: 16px;
		font-weight: 550;
		font-variant-numeric: tabular-nums;
		outline: none;
	}

	input::-webkit-outer-spin-button,
	input::-webkit-inner-spin-button {
		appearance: none;
		margin: 0;
	}

	input[type='number'] {
		-moz-appearance: textfield;
		appearance: textfield;
	}

	input::placeholder {
		color: color-mix(in srgb, var(--muted) 60%, transparent);
	}

	.suffix {
		flex-shrink: 0;
		color: var(--muted);
		font-size: 11px;
		font-weight: 600;
	}
</style>
