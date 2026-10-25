<script lang="ts">
	import Display from '#lib/components/Display.svelte';
	import Key from '#lib/components/Key.svelte';
	import Keypad from '#lib/components/Keypad.svelte';
	import MemoryBar from '#lib/components/MemoryBar.svelte';
	import { evaluate, formatNumber, toExpressionNumber } from '#lib/eval/normal';
	import { history } from '#lib/stores/history.svelte';
	import { ui } from '#lib/stores/ui.svelte';

	let input = $state('');
	let displayExpr = $state<string | null>(null);
	let resultText = $state('0');
	let resultValue = $state(0);
	let error = $state('');
	let justEvaluated = $state(false);
	let pulse = $state(0);

	const preview = $derived.by((): { text: string; value: number } | null => {
		if (!input) return null;
		try {
			const value = evaluate(input);
			return { text: formatNumber(value), value };
		} catch {
			return null;
		}
	});

	const shownExpression = $derived(displayExpr ?? pretty(input));
	const shownResult = $derived(error ? '' : preview?.text ?? resultText);
	const memoryBase = $derived(preview?.value ?? resultValue);

	function pretty(expr: string): string {
		return expr.replace(/\*/g, '×').replace(/\//g, '÷').replace(/-/g, '−');
	}

	function pressDigit(digit: string) {
		if (justEvaluated) {
			input = '';
			justEvaluated = false;
		}
		error = '';
		displayExpr = null;
		const tail = input.match(/(\d*\.?\d*)$/)?.[1] ?? '';
		if (tail === '') {
			if (input && /[%)]$/.test(input)) input += '*';
			input += digit;
		} else if (tail === '0') {
			input = input.slice(0, -1) + digit;
		} else {
			input += digit;
		}
	}

	function pressDecimal() {
		if (justEvaluated) {
			input = '';
			justEvaluated = false;
		}
		error = '';
		displayExpr = null;
		const tail = input.match(/(\d*\.?\d*)$/)?.[1] ?? '';
		if (tail.includes('.')) return;
		if (tail === '') {
			if (input && /[%)]$/.test(input)) input += '*0.';
			else input += '0.';
		} else {
			input += '.';
		}
	}

	function pressOperator(op: string) {
		error = '';
		displayExpr = null;
		justEvaluated = false;
		if (input === '') {
			if (op === '-') input = '-';
			return;
		}
		const last = input.slice(-1);
		if ('+-*/'.includes(last)) {
			if (input.length === 1) {
				if (op === '-') return;
				input = '';
				return;
			}
			input = input.slice(0, -1) + op;
		} else if (last === '(') {
			if (op === '-') input += '-';
		} else {
			input += op;
		}
	}

	function pressPercent() {
		error = '';
		displayExpr = null;
		justEvaluated = false;
		if (input === '' || /[+\-*/%]$/.test(input) || input.endsWith('(')) return;
		input = `${input}%`;
	}

	function pressParen(paren: '(' | ')') {
		if (justEvaluated) {
			input = '';
			justEvaluated = false;
		}
		error = '';
		displayExpr = null;
		if (paren === '(') {
			if (input && !/[+\-*/(]$/.test(input)) input += '*';
			input += '(';
			return;
		}
		const opens = (input.match(/\(/g) ?? []).length;
		const closes = (input.match(/\)/g) ?? []).length;
		if (opens > closes && !/[+\-*/(]$/.test(input)) input += ')';
	}

	function toggleSign() {
		error = '';
		displayExpr = null;
		if (justEvaluated) {
			justEvaluated = false;
			const next = -resultValue;
			resultValue = next;
			resultText = formatNumber(next);
			input = toExpressionNumber(next);
			pulse += 1;
			return;
		}
		if (input === '') {
			input = '-';
			return;
		}
		if (input === '-') {
			input = '';
			return;
		}
		const match = input.match(/(\d*\.?\d+%?)$/);
		if (!match) return;
		const before = input.slice(0, input.length - match[1].length);
		const unaryNegative =
			before.endsWith('-') &&
			(before.length === 1 || /[+\-*/(]$/.test(before.slice(0, -1)));
		if (unaryNegative) input = before.slice(0, -1) + match[1];
		else input = before + '-' + match[1];
	}

	function backspace() {
		error = '';
		displayExpr = null;
		justEvaluated = false;
		input = input.slice(0, -1);
		if (!input) {
			resultText = '0';
			resultValue = 0;
		}
	}

	function clear() {
		input = '';
		displayExpr = null;
		resultText = '0';
		resultValue = 0;
		error = '';
		justEvaluated = false;
	}

	function recall(value: number) {
		if (justEvaluated) {
			input = '';
			justEvaluated = false;
		}
		error = '';
		displayExpr = null;
		const text = toExpressionNumber(value);
		if (input === '' || /[+\-*/(]$/.test(input)) input += text;
		else input += `*${text}`;
	}

	function equals() {
		if (!input) return;
		try {
			const value = evaluate(input);
			const formatted = formatNumber(value);
			const expr = pretty(input);
			resultValue = value;
			resultText = formatted;
			error = '';
			displayExpr = `${expr} =`;
			justEvaluated = true;
			input = toExpressionNumber(value);
			pulse += 1;
			void history.add('normal', expr, formatted);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			justEvaluated = false;
		}
	}

	function onKeydown(event: KeyboardEvent) {
		if (ui.historyOpen || event.metaKey || event.ctrlKey || event.altKey) return;
		const target = event.target as HTMLElement | null;
		if (target && /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName)) return;

		const key = event.key;
		if (key >= '0' && key <= '9') {
			pressDigit(key);
		} else if (key === '.' || key === ',') {
			pressDecimal();
		} else if (key === '+' || key === '-') {
			pressOperator(key);
		} else if (key === '*' || key.toLowerCase() === 'x') {
			pressOperator('*');
		} else if (key === '/') {
			pressOperator('/');
		} else if (key === '%') {
			pressPercent();
		} else if (key === '(' || key === ')') {
			pressParen(key);
		} else if (key === 'Enter' || key === '=') {
			equals();
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
		badge={null}
		{pulse}
		value={resultValue}
	/>

	<MemoryBar current={memoryBase} onrecall={recall} />

	<Keypad columns={4}>
		<Key label="AC" variant="danger" ariaLabel="Clear" onpress={clear} />
		<Key label="⌫" variant="function" ariaLabel="Backspace" onpress={backspace} />
		<Key label="%" variant="operator" ariaLabel="Percent" onpress={pressPercent} />
		<Key label="÷" variant="operator" ariaLabel="Divide" onpress={() => pressOperator('/')} />

		<Key label="7" onpress={() => pressDigit('7')} />
		<Key label="8" onpress={() => pressDigit('8')} />
		<Key label="9" onpress={() => pressDigit('9')} />
		<Key label="×" variant="operator" ariaLabel="Multiply" onpress={() => pressOperator('*')} />

		<Key label="4" onpress={() => pressDigit('4')} />
		<Key label="5" onpress={() => pressDigit('5')} />
		<Key label="6" onpress={() => pressDigit('6')} />
		<Key label="−" variant="operator" ariaLabel="Subtract" onpress={() => pressOperator('-')} />

		<Key label="1" onpress={() => pressDigit('1')} />
		<Key label="2" onpress={() => pressDigit('2')} />
		<Key label="3" onpress={() => pressDigit('3')} />
		<Key label="+" variant="operator" ariaLabel="Add" onpress={() => pressOperator('+')} />

		<Key label="±" variant="function" ariaLabel="Toggle sign" onpress={toggleSign} />
		<Key label="0" onpress={() => pressDigit('0')} />
		<Key label="." ariaLabel="Decimal point" onpress={pressDecimal} />
		<Key label="=" variant="equals" ariaLabel="Equals" onpress={equals} />
	</Keypad>
</div>

<style>
	.mode {
		display: flex;
		flex-direction: column;
		gap: 9px;
	}
</style>
