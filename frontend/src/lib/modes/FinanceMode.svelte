<script lang="ts">
	import { fly } from 'svelte/transition';
	import Field from '#lib/components/Field.svelte';
	import Key from '#lib/components/Key.svelte';
	import Keypad from '#lib/components/Keypad.svelte';
	import { calcService, errorMessage, type LoanSummary, type SavingsResult } from '#lib/bindings';
	import { evaluate, formatNumber } from '#lib/eval/normal';
	import { history } from '#lib/stores/history.svelte';
	import { motion } from '#lib/stores/motion.svelte';
	import { settings } from '#lib/stores/settings.svelte';
	import { ui } from '#lib/stores/ui.svelte';

	type Tab = 'loan' | 'savings' | 'roi' | 'cagr' | 'percent';

	const tabs: Array<{ id: Tab; label: string; hint: string }> = [
		{ id: 'loan', label: 'Loan', hint: 'Amortized payment and full schedule' },
		{ id: 'savings', label: 'Savings', hint: 'Monthly deposits with compound interest' },
		{ id: 'roi', label: 'ROI', hint: 'Return on an investment' },
		{ id: 'cagr', label: 'CAGR', hint: 'Compound annual growth rate' },
		{ id: 'percent', label: 'Change', hint: 'Percentage change between values' }
	];
	const currencies = ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'INR'];

	let tab = $state<Tab>('loan');
	let principal = $state(250000);
	let loanRate = $state(6.5);
	let loanYears = $state(30);
	let paymentsPerYear = $state(12);

	let monthlyDeposit = $state(500);
	let savingsRate = $state(5);
	let savingsYears = $state(15);

	let initial = $state(10000);
	let final = $state(16000);
	let cagrYears = $state(5);
	let fromValue = $state(120);
	let toValue = $state(150);

	let loanResult = $state<LoanSummary | null>(null);
	let savingsResult = $state<SavingsResult | null>(null);
	let scalar = $state<{ title: string; value: string; caption: string } | null>(null);
	let error = $state('');
	let busy = $state(false);

	type FieldKey =
		| 'principal'
		| 'loanRate'
		| 'loanYears'
		| 'monthlyDeposit'
		| 'savingsRate'
		| 'savingsYears'
		| 'initial'
		| 'final'
		| 'cagrYears'
		| 'fromValue'
		| 'toValue';

	interface FieldSpec {
		label: string;
		get: () => number;
		set: (value: number) => void;
		whole?: boolean;
		percent?: boolean;
	}

	const fields: Record<FieldKey, FieldSpec> = {
		principal: { label: 'Principal', get: () => principal, set: (v) => (principal = v) },
		loanRate: {
			label: 'Interest rate',
			get: () => loanRate,
			set: (v) => (loanRate = v),
			percent: true
		},
		loanYears: { label: 'Term', get: () => loanYears, set: (v) => (loanYears = v), whole: true },
		monthlyDeposit: {
			label: 'Monthly deposit',
			get: () => monthlyDeposit,
			set: (v) => (monthlyDeposit = v)
		},
		savingsRate: {
			label: 'Interest rate',
			get: () => savingsRate,
			set: (v) => (savingsRate = v),
			percent: true
		},
		savingsYears: {
			label: 'Horizon',
			get: () => savingsYears,
			set: (v) => (savingsYears = v),
			whole: true
		},
		initial: { label: 'Initial value', get: () => initial, set: (v) => (initial = v) },
		final: { label: 'Final value', get: () => final, set: (v) => (final = v) },
		cagrYears: { label: 'Years', get: () => cagrYears, set: (v) => (cagrYears = v), whole: true },
		fromValue: { label: 'From', get: () => fromValue, set: (v) => (fromValue = v) },
		toValue: { label: 'To', get: () => toValue, set: (v) => (toValue = v) }
	};

	const defaultField: Record<Tab, FieldKey> = {
		loan: 'principal',
		savings: 'monthlyDeposit',
		roi: 'initial',
		cagr: 'initial',
		percent: 'fromValue'
	};

	const operators = ['+', '−', '×', '÷'];

	let activeField = $state<FieldKey>('principal');
	let pad = $state('');

	const activeSpec = $derived(fields[activeField]);
	const padPreview = $derived.by(() => {
		if (!pad) return formatNumber(activeSpec.get(), { significant: 10 });
		try {
			const value = evaluate(pad);
			const shown = activeSpec.whole ? Math.round(value) : value;
			return `${pad} = ${formatNumber(shown, { significant: 10 })}`;
		} catch {
			return pad;
		}
	});

	const currency = $derived(settings.current.currency || 'USD');
	const activeTab = $derived(tabs.find((item) => item.id === tab) ?? tabs[0]);
	const periodLabel = $derived(
		paymentsPerYear === 12
			? 'month'
			: paymentsPerYear === 4
				? 'quarter'
				: paymentsPerYear === 1
					? 'year'
					: 'period'
	);
	const actionLabel = $derived(
		tab === 'loan'
			? 'Calculate loan'
			: tab === 'savings'
				? 'Project savings'
				: tab === 'roi'
					? 'Calculate ROI'
					: tab === 'cagr'
						? 'Calculate CAGR'
						: 'Calculate change'
	);

	function money(value: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency,
			maximumFractionDigits: 2
		}).format(value);
	}

	function pct(value: number): string {
		return `${formatNumber(value, { significant: 10 })}%`;
	}

	function setCurrency(event: Event) {
		const value = (event.currentTarget as HTMLSelectElement).value;
		void settings.update({ currency: value });
	}

	function selectTab(id: Tab) {
		tab = id;
		activeField = defaultField[id];
		pad = '';
		error = '';
	}

	function activateField(event: FocusEvent) {
		const target = event.target as HTMLElement | null;
		const slot = target?.closest<HTMLElement>('[data-field]');
		const key = slot?.dataset.field as FieldKey | undefined;
		if (!key || key === activeField) return;
		activeField = key;
		pad = '';
		error = '';
	}

	function pressOperator(op: string) {
		if (op === '%' && activeSpec.percent) return;
		if (!pad) pad = String(activeSpec.get());
		const last = pad.slice(-1);
		if (op === '%' && (last === '%' || operators.includes(last))) return;
		if (op !== '%' && operators.includes(last) && op !== '−') {
			pad = pad.slice(0, -1) + op;
			return;
		}
		pad += op;
	}

	function appendEntry(token: string) {
		error = '';
		if (token === 'AC') {
			pad = '';
			return;
		}
		if (token === '⌫') {
			pad = pad.slice(0, -1);
			return;
		}
		if (token === '=') {
			applyPad();
			return;
		}
		if (token === '±') {
			toggleSign();
			return;
		}
		if (token === '.') {
			const tail = pad.match(/[0-9.]*$/)?.[0] ?? '';
			if (!tail.includes('.')) pad += tail ? '.' : '0.';
			return;
		}
		if (operators.includes(token) || token === '%') {
			pressOperator(token);
			return;
		}
		pad += token;
	}

	function toggleSign() {
		error = '';
		if (!pad) {
			pad = String(-activeSpec.get());
			return;
		}
		try {
			pad = String(-evaluate(pad));
		} catch {
			return;
		}
	}

	function applyPad(andCalculate = false) {
		if (!pad) {
			if (andCalculate) void calculate();
			return;
		}
		try {
			const value = evaluate(pad);
			const spec = activeSpec;
			spec.set(spec.whole ? Math.round(value) : value);
			pad = '';
			error = '';
			if (andCalculate) void calculate();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Invalid entry';
		}
	}

	function positive(value: number, label: string): boolean {
		if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) {
			error = `${label} must be a positive number`;
			return false;
		}
		return true;
	}

	function nonNegative(value: number, label: string): boolean {
		if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
			error = `${label} cannot be negative`;
			return false;
		}
		return true;
	}

	function wholeYears(value: number): boolean {
		if (typeof value !== 'number' || !Number.isInteger(value) || value <= 0) {
			error = 'Years must be a positive whole number';
			return false;
		}
		return true;
	}

	async function calculate() {
		if (busy) return;
		error = '';

		if (tab === 'loan') {
			if (!positive(principal, 'Principal')) return;
			if (!nonNegative(loanRate, 'Interest rate')) return;
			if (!wholeYears(loanYears)) return;
		} else if (tab === 'savings') {
			if (!positive(monthlyDeposit, 'Monthly deposit')) return;
			if (!nonNegative(savingsRate, 'Interest rate')) return;
			if (!wholeYears(savingsYears)) return;
		} else if (tab === 'roi') {
			if (!positive(initial, 'Initial value')) return;
			if (!nonNegative(final, 'Final value')) return;
		} else if (tab === 'cagr') {
			if (!positive(initial, 'Initial value')) return;
			if (!positive(final, 'Final value')) return;
			if (!wholeYears(cagrYears)) return;
		} else {
			if (!positive(fromValue, 'Starting value')) return;
			if (!nonNegative(toValue, 'Ending value')) return;
		}

		busy = true;
		try {
			if (tab === 'loan') {
				const result = await calcService.calcLoan(principal, loanRate, loanYears, paymentsPerYear);
				loanResult = result;
				void history.add(
					'finance',
					`Loan ${money(principal)} @ ${loanRate}% × ${loanYears}y`,
					`${money(result.paymentPerPeriod)}/${periodLabel}`
				);
			} else if (tab === 'savings') {
				const result = await calcService.calcSavings(monthlyDeposit, savingsRate, savingsYears);
				savingsResult = result;
				void history.add(
					'finance',
					`Savings ${money(monthlyDeposit)}/mo @ ${savingsRate}% × ${savingsYears}y`,
					money(result.futureValue)
				);
			} else if (tab === 'roi') {
				const value = await calcService.calcROI(initial, final);
				scalar = {
					title: 'Return on investment',
					value: pct(value),
					caption: `${money(initial)} → ${money(final)}`
				};
				void history.add(
					'finance',
					`ROI ${money(initial)} → ${money(final)}`,
					pct(value)
				);
			} else if (tab === 'cagr') {
				const value = await calcService.calcCAGR(initial, final, cagrYears);
				scalar = {
					title: 'Compound annual growth',
					value: pct(value),
					caption: `${money(initial)} → ${money(final)} over ${cagrYears} years`
				};
				void history.add(
					'finance',
					`CAGR ${money(initial)} → ${money(final)} / ${cagrYears}y`,
					pct(value)
				);
			} else {
				const value = await calcService.calcPercentChange(fromValue, toValue);
				scalar = {
					title: 'Percentage change',
					value: pct(value),
					caption: `${formatNumber(fromValue, { significant: 10 })} → ${formatNumber(toValue, { significant: 10 })}`
				};
				void history.add(
					'finance',
					`Change ${formatNumber(fromValue, { significant: 10 })} → ${formatNumber(toValue, { significant: 10 })}`,
					pct(value)
				);
			}
		} catch (err) {
			error = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	function onKeydown(event: KeyboardEvent) {
		if (ui.historyOpen) return;
		if (event.metaKey || event.ctrlKey || event.altKey) return;
		const target = event.target as HTMLElement | null;
		if (event.key === 'Escape') {
			if (pad || error) {
				pad = '';
				error = '';
				event.preventDefault();
			}
			return;
		}
		if (event.key === 'Enter') {
			const onPadKey =
				target?.tagName === 'BUTTON' &&
				target.closest('.pad') !== null &&
				!target.classList.contains('pad-clear');
			if (onPadKey) {
				event.preventDefault();
				applyPad(true);
				return;
			}
			if (target?.tagName === 'BUTTON' || target?.tagName === 'SELECT') return;
			event.preventDefault();
			applyPad(true);
			return;
		}
		if (target?.tagName === 'INPUT' || target?.tagName === 'SELECT') return;
		if (event.key === 'Backspace') {
			event.preventDefault();
			appendEntry('⌫');
			return;
		}
		const alias: Record<string, string> = {
			'+': '+',
			'-': '−',
			'*': '×',
			'/': '÷',
			'%': '%',
			'=': '='
		};
		const token = /^[0-9]$/.test(event.key) || event.key === '.' ? event.key : alias[event.key];
		if (token) {
			event.preventDefault();
			appendEntry(token);
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div class="mode">
	<div class="segmented glass" role="tablist" aria-label="Finance calculation">
		{#each tabs as item (item.id)}
			<button
				type="button"
				role="tab"
				class="segment"
				class:active={tab === item.id}
				aria-selected={tab === item.id}
				onclick={() => selectTab(item.id)}
			>
				{item.label}
			</button>
		{/each}
	</div>

	<p class="hint">{activeTab.hint}</p>

	<div class="form">
		<div class="form-head">
			<label class="currency">
				<span>Currency</span>
				<select value={currency} onchange={setCurrency}>
					{#each currencies as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
			</label>
		</div>

		<div class="grid" onfocusin={activateField}>
			{#if tab === 'loan'}
				<div class="field-slot" data-field="principal">
					<Field label="Principal" bind:value={principal} suffix={currency} min={0} />
				</div>
				<div class="field-slot" data-field="loanRate">
					<Field label="Interest rate" bind:value={loanRate} suffix="% / yr" min={0} step={0.1} />
				</div>
				<div class="field-slot" data-field="loanYears">
					<Field label="Term" bind:value={loanYears} suffix="years" min={1} step={1} />
				</div>
				<label class="field">
					<span class="label">Payments</span>
					<span class="select-box">
						<select bind:value={paymentsPerYear}>
							<option value={12}>Monthly (12)</option>
							<option value={4}>Quarterly (4)</option>
							<option value={1}>Yearly (1)</option>
						</select>
					</span>
				</label>
			{:else if tab === 'savings'}
				<div class="field-slot" data-field="monthlyDeposit">
					<Field label="Monthly deposit" bind:value={monthlyDeposit} suffix={currency} min={0} />
				</div>
				<div class="field-slot" data-field="savingsRate">
					<Field label="Interest rate" bind:value={savingsRate} suffix="% / yr" min={0} step={0.1} />
				</div>
				<div class="field-slot" data-field="savingsYears">
					<Field label="Horizon" bind:value={savingsYears} suffix="years" min={1} step={1} />
				</div>
			{:else if tab === 'roi'}
				<div class="field-slot" data-field="initial">
					<Field label="Initial value" bind:value={initial} suffix={currency} min={0} />
				</div>
				<div class="field-slot" data-field="final">
					<Field label="Final value" bind:value={final} suffix={currency} min={0} />
				</div>
			{:else if tab === 'cagr'}
				<div class="field-slot" data-field="initial">
					<Field label="Initial value" bind:value={initial} suffix={currency} min={0} />
				</div>
				<div class="field-slot" data-field="final">
					<Field label="Final value" bind:value={final} suffix={currency} min={0} />
				</div>
				<div class="field-slot" data-field="cagrYears">
					<Field label="Years" bind:value={cagrYears} suffix="years" min={1} step={1} />
				</div>
			{:else}
				<div class="field-slot" data-field="fromValue">
					<Field label="From" bind:value={fromValue} suffix={currency} min={0} />
				</div>
				<div class="field-slot" data-field="toValue">
					<Field label="To" bind:value={toValue} suffix={currency} min={0} />
				</div>
			{/if}
		</div>

		<div class="pad" role="group" aria-label="Finance input keypad">
			<div class="pad-head">
				<span class="pad-target">Editing <strong>{activeSpec.label}</strong></span>
				<span class="pad-preview" aria-live="polite">{padPreview}</span>
				<button
					type="button"
					class="pad-clear"
					disabled={!pad}
					onclick={() => appendEntry('AC')}
				>
					AC
				</button>
			</div>
			<Keypad columns={5} gap={5}>
				<Key label="7" onpress={() => appendEntry('7')} />
				<Key label="8" onpress={() => appendEntry('8')} />
				<Key label="9" onpress={() => appendEntry('9')} />
				<Key label="÷" variant="operator" ariaLabel="Divide" onpress={() => appendEntry('÷')} />
				<Key label="⌫" variant="function" ariaLabel="Backspace" onpress={() => appendEntry('⌫')} />
				<Key label="4" onpress={() => appendEntry('4')} />
				<Key label="5" onpress={() => appendEntry('5')} />
				<Key label="6" onpress={() => appendEntry('6')} />
				<Key label="×" variant="operator" ariaLabel="Multiply" onpress={() => appendEntry('×')} />
				<Key
					label="%"
					variant="operator"
					ariaLabel="Percent"
					disabled={activeSpec.percent}
					title={activeSpec.percent ? 'Interest rates are already percentages' : undefined}
					onpress={() => appendEntry('%')}
				/>
				<Key label="1" onpress={() => appendEntry('1')} />
				<Key label="2" onpress={() => appendEntry('2')} />
				<Key label="3" onpress={() => appendEntry('3')} />
				<Key label="−" variant="operator" ariaLabel="Subtract" onpress={() => appendEntry('−')} />
				<Key label="±" variant="operator" ariaLabel="Negate" onpress={() => appendEntry('±')} />
				<Key label="0" span={2} onpress={() => appendEntry('0')} />
				<Key label="." onpress={() => appendEntry('.')} />
				<Key label="+" variant="operator" ariaLabel="Add" onpress={() => appendEntry('+')} />
				<Key label="=" variant="equals" ariaLabel="Apply value" onpress={() => appendEntry('=')} />
			</Keypad>
		</div>

		<button type="button" class="calculate" onclick={() => applyPad(true)} disabled={busy}>
			{busy ? 'Calculating…' : actionLabel}
		</button>

		{#if error}<p class="error-text">{error}</p>{/if}
	</div>

	{#if tab === 'loan' && loanResult}
		<section
			class="result"
			transition:fly={{ y: motion.reduced ? 0 : 10, duration: motion.reduced ? 0 : 220 }}
		>
			<p class="eyebrow">Payment per {periodLabel}</p>
			<p class="hero">
				{money(loanResult.paymentPerPeriod)}<span>/{periodLabel}</span>
			</p>
			<div class="stats">
				<div>
					<span>Total paid</span>
					<strong>{money(loanResult.totalPaid)}</strong>
				</div>
				<div>
					<span>Total interest</span>
					<strong>{money(loanResult.totalInterest)}</strong>
				</div>
				<div>
					<span>Payments</span>
					<strong>{loanResult.numPeriods}</strong>
				</div>
			</div>
			<div class="schedule">
				<div class="schedule-row head">
					<span>#</span>
					<span>Payment</span>
					<span>Principal</span>
					<span>Interest</span>
					<span>Balance</span>
				</div>
				{#each loanResult.schedule as row (row.number)}
					<div class="schedule-row">
						<span>{row.number}</span>
						<span>{money(row.payment)}</span>
						<span>{money(row.principal)}</span>
						<span>{money(row.interest)}</span>
						<span>{money(row.balance)}</span>
					</div>
				{/each}
			</div>
		</section>
	{:else if tab === 'savings' && savingsResult}
		<section
			class="result"
			transition:fly={{ y: motion.reduced ? 0 : 10, duration: motion.reduced ? 0 : 220 }}
		>
			<p class="eyebrow">Future value</p>
			<p class="hero">{money(savingsResult.futureValue)}</p>
			<div class="stats">
				<div>
					<span>Deposited</span>
					<strong>{money(savingsResult.totalDeposited)}</strong>
				</div>
				<div>
					<span>Interest earned</span>
					<strong>{money(savingsResult.totalInterest)}</strong>
				</div>
			</div>
		</section>
	{:else if tab !== 'loan' && tab !== 'savings' && scalar}
		<section
			class="result"
			transition:fly={{ y: motion.reduced ? 0 : 10, duration: motion.reduced ? 0 : 220 }}
		>
			<p class="eyebrow">{scalar.title}</p>
			<p class="hero">{scalar.value}</p>
			<p class="caption">{scalar.caption}</p>
		</section>
	{/if}
</div>

<style>
	.mode {
		display: flex;
		flex-direction: column;
		gap: 9px;
	}

	.segmented {
		display: grid;
		grid-template-columns: repeat(5, minmax(0, 1fr));
		gap: 3px;
		padding: 4px;
		border-radius: 14px;
		flex-shrink: 0;
	}

	.segment {
		height: 32px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		color: var(--muted);
		font-size: 11px;
		font-weight: 650;
		transition:
			color 0.16s ease,
			background 0.16s ease;
	}

	.segment:hover {
		color: var(--text);
	}

	.segment.active {
		color: var(--text);
		background: color-mix(in srgb, var(--accent) 20%, transparent);
		box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 32%, transparent);
	}

	.hint {
		margin: -2px 2px 0;
		color: var(--muted);
		font-size: 11px;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.form-head {
		display: flex;
		justify-content: flex-end;
	}

	.currency {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--muted);
		font-size: 10px;
		font-weight: 650;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.currency select {
		height: 26px;
		padding: 0 8px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface);
		color: var(--text);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.02em;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 9px;
	}

	.field-slot {
		min-width: 0;
	}

	.pad {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-top: 2px;
	}

	.pad-head {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 22px;
	}

	.pad-target {
		flex-shrink: 0;
		color: var(--muted);
		font-size: 10px;
		font-weight: 650;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.pad-target strong {
		color: var(--text);
	}

	.pad-preview {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-align: right;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--muted);
		font-family: var(--mono);
		font-size: 11px;
	}

	.pad-clear {
		flex-shrink: 0;
		height: 22px;
		padding: 0 7px;
		border: 1px solid var(--border);
		border-radius: 7px;
		background: transparent;
		color: var(--muted);
		font-size: 9px;
		font-weight: 700;
		letter-spacing: 0.06em;
	}

	.pad-clear:hover:not(:disabled) {
		color: var(--text);
		border-color: var(--border-strong);
	}

	.pad-clear:disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}

	.pad :global(.key) {
		min-height: 34px;
		border-radius: 11px;
		font-size: 15px;
	}

	.pad :global(.key.operator) {
		font-size: 17px;
	}

	.pad :global(.key.equals) {
		font-size: 16px;
	}

	.pad :global(.key.function) {
		font-size: 12px;
	}

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

	.select-box {
		display: flex;
		align-items: center;
		height: 42px;
		padding: 0 8px 0 4px;
		border: 1px solid var(--border);
		border-radius: 12px;
		background: var(--surface);
	}

	.select-box:focus-within {
		border-color: color-mix(in srgb, var(--accent) 55%, var(--border));
	}

	.select-box select {
		width: 100%;
		border: 0;
		background: transparent;
		color: var(--text);
		font-size: 14px;
		font-weight: 550;
		outline: none;
	}

	.calculate {
		height: 46px;
		border: 0;
		border-radius: 14px;
		background: linear-gradient(
			140deg,
			color-mix(in srgb, var(--accent) 88%, white 12%),
			var(--accent)
		);
		color: var(--accent-contrast);
		font-size: 14px;
		font-weight: 700;
		letter-spacing: 0.01em;
		box-shadow: 0 10px 26px color-mix(in srgb, var(--accent) 30%, transparent);
		transition:
			filter 0.16s ease,
			transform 0.1s ease,
			opacity 0.16s ease;
	}

	.calculate:hover:not(:disabled) {
		filter: brightness(1.06);
	}

	.calculate:active:not(:disabled) {
		transform: scale(0.985);
	}

	.calculate:disabled {
		opacity: 0.6;
		cursor: progress;
	}

	.error-text {
		margin: 0;
		color: var(--danger);
		font-size: 11.5px;
		line-height: 1.4;
	}

	.result {
		padding: 14px;
		border: 1px solid color-mix(in srgb, var(--accent) 26%, var(--border));
		border-radius: 18px;
		background:
			linear-gradient(
				160deg,
				color-mix(in srgb, var(--accent) 12%, transparent),
				transparent 55%
			),
			var(--surface);
	}

	.eyebrow {
		margin: 0;
		color: var(--muted);
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.hero {
		margin: 5px 0 0;
		font-size: 30px;
		font-weight: 650;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
		line-height: 1.15;
	}

	.hero span {
		color: var(--muted);
		font-size: 14px;
		font-weight: 600;
		letter-spacing: 0;
	}

	.caption {
		margin: 5px 0 0;
		color: var(--muted);
		font-size: 11.5px;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 8px;
		margin-top: 12px;
	}

	.stats div {
		display: flex;
		flex-direction: column;
		gap: 3px;
		min-width: 0;
		padding: 8px 9px;
		border-radius: 10px;
		background: color-mix(in srgb, var(--bg-0) 35%, transparent);
	}

	.stats span {
		color: var(--muted);
		font-size: 9.5px;
		font-weight: 650;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.stats strong {
		overflow: hidden;
		text-overflow: ellipsis;
		font-size: 12.5px;
		font-weight: 650;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.schedule {
		margin-top: 12px;
		max-height: 168px;
		overflow-y: auto;
		border-top: 1px solid var(--border);
	}

	.schedule-row {
		display: grid;
		grid-template-columns: 26px repeat(4, minmax(0, 1fr));
		gap: 5px;
		padding: 5px 2px;
		border-bottom: 1px solid color-mix(in srgb, var(--border) 55%, transparent);
		font-family: var(--mono);
		font-size: 10px;
		font-variant-numeric: tabular-nums;
		text-align: right;
	}

	.schedule-row span:first-child {
		text-align: left;
		color: var(--muted);
	}

	.schedule-row.head {
		position: sticky;
		top: 0;
		z-index: 1;
		background: var(--bg-1);
		color: var(--muted);
		font-family: var(--font);
		font-size: 9px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	@media (max-width: 380px) {
		.grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
