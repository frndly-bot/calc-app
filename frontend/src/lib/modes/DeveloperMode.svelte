<script lang="ts">
	import { untrack } from 'svelte';
	import Display from '#lib/components/Display.svelte';
	import Key from '#lib/components/Key.svelte';
	import Keypad from '#lib/components/Keypad.svelte';
	import { errorMessage, calcService, type Formatted } from '#lib/bindings';
	import { history } from '#lib/stores/history.svelte';
	import { settings } from '#lib/stores/settings.svelte';
	import { ui } from '#lib/stores/ui.svelte';

	const wordSizes = [8, 16, 32, 64] as const;
	const bases = [
		{ id: 16, label: 'HEX', field: 'Hex' as const },
		{ id: 10, label: 'DEC', field: 'Dec' as const },
		{ id: 8, label: 'OCT', field: 'Oct' as const },
		{ id: 2, label: 'BIN', field: 'Bin' as const }
	];
	const ops = ['AND', 'OR', 'XOR', 'NAND', 'NOR', 'SHL', 'SHR', 'SAR', 'ROL', 'ROR', 'ADD', 'SUB', 'MUL', 'DIV', 'MOD'];

	let base = $state(16);
	let entry = $state('');
	let raw = $state('0');
	let formatted = $state<Formatted | null>(null);
	let pendingOp = $state<string | null>(null);
	let operandA = $state<string | null>(null);
	let operandLabel = $state('');
	let error = $state('');
	let busy = $state(false);
	let refreshToken = 0;
	let parseToken = 0;

	const bits = $derived(settings.current.devBits || 32);
	const signed = $derived(settings.current.devSigned ?? false);
	const baseInfo = $derived(bases.find((item) => item.id === base) ?? bases[0]);
	const activeValue = $derived(formatted ? formatted[baseInfo.field] : '0');
	const shownResult = $derived(entry || activeValue);
	const shownExpression = $derived(
		pendingOp
			? `${operandLabel} ${pendingOp}`
			: `${bits}-bit · ${signed ? 'signed' : 'unsigned'}`
	);
	const bitCells = $derived.by(() => {
		const bin = (formatted?.Bin ?? '').padStart(bits, '0');
		const cells: Array<{ index: number; set: boolean }> = [];
		for (let position = 0; position < bits; position++) {
			cells.push({ index: bits - 1 - position, set: bin[position] === '1' });
		}
		return cells;
	});

	function apply(result: { raw: string; formatted: Formatted }) {
		raw = result.raw;
		formatted = result.formatted;
	}

	function valueIn(baseId: number, values: Formatted): string {
		switch (baseId) {
			case 16:
				return values.Hex;
			case 8:
				return values.Oct;
			case 10:
				return values.Dec;
			default:
				return values.Bin;
		}
	}

	function prefixFor(baseId: number): string {
		if (baseId === 16) return '0x';
		if (baseId === 8) return '0o';
		if (baseId === 2) return '0b';
		return '';
	}

	function displayIn(values: Formatted): string {
		return `${prefixFor(base)}${valueIn(base, values)}`;
	}

	function digitAllowed(digit: string): boolean {
		return Number.parseInt(digit, 16) < base;
	}

	async function refresh(value: string, width = bits, isSigned = signed) {
		const token = ++refreshToken;
		try {
			const result = await calcService.devFormat(value, width, isSigned);
			if (token !== refreshToken) return;
			apply(result);
			error = '';
		} catch (err) {
			if (token === refreshToken) error = errorMessage(err);
		}
	}

	$effect(() => {
		const width = bits;
		const isSigned = signed;
		void refresh(untrack(() => raw), width, isSigned);
	});

	$effect(() => {
		const text = entry;
		const baseId = base;
		const width = bits;
		const isSigned = signed;
		if (!text) return;

		const token = ++parseToken;
		const timer = setTimeout(async () => {
			try {
				const parsed = await calcService.devParse(text, baseId, width, isSigned);
				if (token !== parseToken) return;
				apply(parsed);
				error = '';
			} catch {
				if (token === parseToken) error = '';
			}
		}, 140);

		return () => {
			clearTimeout(timer);
			if (token === parseToken) parseToken++;
		};
	});

	function pressDigit(digit: string) {
		if (!digitAllowed(digit)) return;
		error = '';
		entry = `${entry}${digit}`;
	}

	function backspace() {
		error = '';
		if (entry) {
			entry = entry.slice(0, -1);
			return;
		}
		pendingOp = null;
		operandA = null;
		operandLabel = '';
		void refresh('0');
	}

	function clear() {
		entry = '';
		pendingOp = null;
		operandA = null;
		operandLabel = '';
		error = '';
		void refresh('0');
	}

	async function selectBase(next: number) {
		if (next === base) return;
		base = next;
		entry = '';
		error = '';
	}

	async function setBits(next: number) {
		if (next === bits) return;
		entry = '';
		pendingOp = null;
		operandA = null;
		error = '';
		await settings.update({ devBits: next });
	}

	async function setSigned(next: boolean) {
		if (next === signed) return;
		entry = '';
		error = '';
		await settings.update({ devSigned: next });
	}

	async function pressOperator(op: string) {
		if (busy) return;
		error = '';
		if (pendingOp && operandA !== null && entry) {
			busy = true;
			try {
				const second = await calcService.devParse(entry, base, bits, signed);
				const combined = await calcService.devBinary(pendingOp, operandA, second.raw, bits, signed);
				apply(combined);
				operandA = combined.raw;
				operandLabel = displayIn(combined.formatted);
				entry = '';
				pendingOp = op;
			} catch (err) {
				error = errorMessage(err);
			} finally {
				busy = false;
			}
			return;
		}
		if (pendingOp) {
			pendingOp = op;
			return;
		}
		if (entry) {
			busy = true;
			try {
				const parsed = await calcService.devParse(entry, base, bits, signed);
				apply(parsed);
				operandA = parsed.raw;
				operandLabel = displayIn(parsed.formatted);
				entry = '';
				pendingOp = op;
			} catch (err) {
				error = errorMessage(err);
			} finally {
				busy = false;
			}
			return;
		}
		operandA = raw;
		operandLabel = displayIn(formatted ?? { Hex: '0', Dec: '0', Oct: '0', Bin: '0' });
		pendingOp = op;
	}

	async function applyUnary(op: string) {
		if (busy) return;
		busy = true;
		error = '';
		const label = entry || displayIn(formatted ?? { Hex: '0', Dec: '0', Oct: '0', Bin: '0' });
		try {
			let operand = raw;
			if (entry) {
				const parsed = await calcService.devParse(entry, base, bits, signed);
				operand = parsed.raw;
			}
			const result = await calcService.devUnary(op, operand, bits, signed);
			apply(result);
			entry = '';
			pendingOp = null;
			operandA = null;
			operandLabel = '';
			void history.add('developer', `${op} ${label}`, displayIn(result.formatted));
		} catch (err) {
			error = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	async function toggleBit(index: number) {
		if (busy) return;
		error = '';
		let current = raw;
		if (entry) {
			try {
				const parsed = await calcService.devParse(entry, base, bits, signed);
				current = parsed.raw;
				apply(parsed);
				entry = '';
			} catch (err) {
				error = errorMessage(err);
				return;
			}
		}
		const mask = (1n << BigInt(bits)) - 1n;
		const next = (BigInt(current) ^ (1n << BigInt(index))) & mask;
		await refresh(next.toString());
	}

	async function equals() {
		if (busy) return;
		error = '';
		if (pendingOp && operandA !== null) {
			if (!entry) {
				error = 'Enter a second operand';
				return;
			}
			busy = true;
			const rightText = entry;
			try {
				const second = await calcService.devParse(entry, base, bits, signed);
				const result = await calcService.devBinary(pendingOp, operandA, second.raw, bits, signed);
				const expr = `${operandLabel} ${pendingOp} ${prefixFor(base)}${rightText}`;
				apply(result);
				entry = '';
				pendingOp = null;
				operandA = null;
				operandLabel = '';
				void history.add('developer', expr, displayIn(result.formatted));
			} catch (err) {
				error = errorMessage(err);
			} finally {
				busy = false;
			}
			return;
		}
		if (!entry) return;
		busy = true;
		const text = entry;
		try {
			const result = await calcService.devParse(entry, base, bits, signed);
			apply(result);
			entry = '';
			void history.add(
				'developer',
				`${prefixFor(base)}${text}`,
				displayIn(result.formatted)
			);
		} catch (err) {
			error = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	function onKeydown(event: KeyboardEvent) {
		if (ui.historyOpen || event.metaKey || event.ctrlKey || event.altKey) return;
		const target = event.target as HTMLElement | null;
		if (target && /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName)) return;

		const key = event.key;
		const upper = key.toUpperCase();
		if (key >= '0' && key <= '9') {
			pressDigit(key);
		} else if (/^[A-F]$/.test(upper)) {
			pressDigit(upper);
		} else if (key === '+') {
			void pressOperator('ADD');
		} else if (key === '-') {
			void pressOperator('SUB');
		} else if (key === '*') {
			void pressOperator('MUL');
		} else if (key === '/') {
			void pressOperator('DIV');
		} else if (key === '%') {
			void pressOperator('MOD');
		} else if (key === 'Enter' || key === '=') {
			void equals();
		} else if (key === 'Backspace') {
			backspace();
		} else if (key === 'Escape' || key === 'Delete') {
			clear();
		} else {
			return;
		}
		event.preventDefault();
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div class="mode">
	<Display
		expression={shownExpression}
		result={shownResult}
		{error}
		badge={baseInfo.label}
		value={null}
	/>

	<div class="controls">
		{#each wordSizes as size (size)}
			<button
				type="button"
				class="chip"
				class:active={size === bits}
				onclick={() => void setBits(size)}
			>
				{size}
			</button>
		{/each}
		<div class="spacer"></div>
		<button
			type="button"
			class="chip"
			class:active={signed}
			onclick={() => void setSigned(!signed)}
			aria-pressed={signed}
		>
			{signed ? 'SIGNED' : 'UNSIGNED'}
		</button>
	</div>

	<div class="readouts">
		{#each bases as item (item.id)}
			<button
				type="button"
				class="readout"
				class:active={base === item.id}
				class:wide={item.id === 2}
				onclick={() => void selectBase(item.id)}
				aria-pressed={base === item.id}
			>
				<span class="tag">{item.label}</span>
				<span class="number">{formatted ? formatted[item.field] : '—'}</span>
			</button>
		{/each}
	</div>

	<div class="bitgrid" style:--bits={bits}>
		{#each bitCells as cell (cell.index)}
			<button
				type="button"
				class="bit"
				class:set={cell.set}
				onclick={() => void toggleBit(cell.index)}
				title={`Bit ${cell.index}`}
				aria-label={`Bit ${cell.index}, ${cell.set ? 'set' : 'clear'}`}
			>
				{cell.set ? 1 : 0}
			</button>
		{/each}
	</div>

	<div class="ops" aria-label="Bitwise operations">
		{#each ops as op (op)}
			<button
				type="button"
				class="op"
				class:active={pendingOp === op}
				onclick={() => void pressOperator(op)}
			>
				{op}
			</button>
		{/each}
	</div>

	<Keypad columns={4}>
		<Key label="A" onpress={() => pressDigit('A')} disabled={!digitAllowed('A')} />
		<Key label="B" onpress={() => pressDigit('B')} disabled={!digitAllowed('B')} />
		<Key label="C" onpress={() => pressDigit('C')} disabled={!digitAllowed('C')} />
		<Key label="D" onpress={() => pressDigit('D')} disabled={!digitAllowed('D')} />

		<Key label="E" onpress={() => pressDigit('E')} disabled={!digitAllowed('E')} />
		<Key label="F" onpress={() => pressDigit('F')} disabled={!digitAllowed('F')} />
		<Key label="AC" variant="danger" ariaLabel="Clear" onpress={clear} />
		<Key label="⌫" variant="function" ariaLabel="Backspace" onpress={backspace} />

		<Key label="7" onpress={() => pressDigit('7')} disabled={!digitAllowed('7')} />
		<Key label="8" onpress={() => pressDigit('8')} disabled={!digitAllowed('8')} />
		<Key label="9" onpress={() => pressDigit('9')} disabled={!digitAllowed('9')} />
		<Key label="NEG" variant="function" ariaLabel="Negate" onpress={() => void applyUnary('NEG')} />

		<Key label="4" onpress={() => pressDigit('4')} disabled={!digitAllowed('4')} />
		<Key label="5" onpress={() => pressDigit('5')} disabled={!digitAllowed('5')} />
		<Key label="6" onpress={() => pressDigit('6')} disabled={!digitAllowed('6')} />
		<Key
			label="NOT"
			variant="function"
			ariaLabel="Bitwise not"
			onpress={() => void applyUnary('NOT')}
		/>

		<Key label="1" onpress={() => pressDigit('1')} disabled={!digitAllowed('1')} />
		<Key label="2" onpress={() => pressDigit('2')} disabled={!digitAllowed('2')} />
		<Key label="3" onpress={() => pressDigit('3')} disabled={!digitAllowed('3')} />
		<Key label="=" variant="equals" ariaLabel="Equals" onpress={() => void equals()} />

		<Key label="0" span={4} onpress={() => pressDigit('0')} disabled={!digitAllowed('0')} />
	</Keypad>
</div>

<style>
	.mode {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 5px;
		flex-wrap: wrap;
	}

	.spacer {
		flex: 1;
	}

	.readouts {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 6px;
	}

	.readout {
		display: flex;
		align-items: baseline;
		gap: 7px;
		min-width: 0;
		padding: 8px 10px;
		border: 1px solid var(--border);
		border-radius: 11px;
		background: var(--surface);
		text-align: left;
		transition:
			border-color 0.16s ease,
			background 0.16s ease;
	}

	.readout:hover {
		background: var(--surface-strong);
	}

	.readout.active {
		border-color: color-mix(in srgb, var(--accent) 55%, transparent);
		background: color-mix(in srgb, var(--accent) 13%, var(--surface));
	}

	.readout.wide {
		grid-column: span 2;
	}

	.tag {
		color: var(--muted);
		font-size: 9px;
		font-weight: 750;
		letter-spacing: 0.08em;
		flex-shrink: 0;
	}

	.number {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-family: var(--mono);
		font-size: 13px;
		font-variant-numeric: tabular-nums;
	}

	.readout.wide .number {
		white-space: normal;
		overflow-wrap: anywhere;
		font-size: 11px;
		line-height: 1.35;
		max-height: 45px;
		overflow-y: auto;
	}

	.bitgrid {
		display: grid;
		grid-template-columns: repeat(8, minmax(0, 1fr));
		gap: 3px;
	}

	.bit {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 22px;
		border: 1px solid var(--border);
		border-radius: 5px;
		background: var(--surface);
		color: var(--muted);
		font-family: var(--mono);
		font-size: 9px;
		transition:
			background 0.12s ease,
			color 0.12s ease,
			border-color 0.12s ease;
	}

	.bit:hover {
		border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
	}

	.bit.set {
		background: color-mix(in srgb, var(--accent) 30%, var(--surface));
		border-color: color-mix(in srgb, var(--accent) 40%, transparent);
		color: var(--text);
		font-weight: 700;
	}

	.ops {
		display: flex;
		flex-wrap: wrap;
		gap: 5px;
	}

	.op {
		flex-shrink: 0;
		height: 27px;
		padding: 0 9px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface);
		color: var(--muted);
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.04em;
		transition:
			color 0.14s ease,
			background 0.14s ease,
			border-color 0.14s ease;
	}

	.op:hover {
		color: var(--text);
		background: var(--surface-hover);
	}

	.op.active {
		color: var(--text);
		background: color-mix(in srgb, var(--accent) 24%, transparent);
		border-color: color-mix(in srgb, var(--accent) 50%, transparent);
	}

</style>
