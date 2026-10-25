<script lang="ts">
	import Display from '#lib/components/Display.svelte';
	import Key from '#lib/components/Key.svelte';
	import Keypad from '#lib/components/Keypad.svelte';
	import MemoryBar from '#lib/components/MemoryBar.svelte';
	import { errorMessage, calcService } from '#lib/bindings';
	import { formatNumber, toExpressionNumber } from '#lib/eval/normal';
	import { history } from '#lib/stores/history.svelte';
	import { settings } from '#lib/stores/settings.svelte';
	import { ui } from '#lib/stores/ui.svelte';

	const angleUnits = ['deg', 'rad', 'grad'] as const;

	let expr = $state('');
	let displayExpr = $state<string | null>(null);
	let resultText = $state('0');
	let resultValue = $state(0);
	let previewText = $state<string | null>(null);
	let error = $state('');
	let second = $state(false);
	let justEvaluated = $state(false);
	let pulse = $state(0);
	let busy = $state(false);
	let previewToken = 0;

	const angleUnit = $derived(settings.current.angleUnit || 'deg');
	const shownExpression = $derived(displayExpr ?? prettyScientific(expr));
	const shownResult = $derived(error ? '' : previewText ?? resultText);
	const parenDepth = $derived.by(() => {
		let depth = 0;
		for (const char of expr) {
			if (char === '(') depth++;
			else if (char === ')') depth = Math.max(0, depth - 1);
		}
		return depth;
	});

	function prettyScientific(source: string): string {
		return source
			.replace(/\bexp10\(/g, '10^(')
			.replace(/\bexp\(/g, 'e^(')
			.replace(/\bsqrt\(/g, '√(')
			.replace(/\bcbrt\(/g, '∛(')
			.replace(/\bpi\b/g, 'π')
			.replace(/\btau\b/g, 'τ')
			.replace(/\*/g, '×')
			.replace(/\//g, '÷')
			.replace(/-/g, '−');
	}

	function isValueEnd(source: string): boolean {
		if (!source) return false;
		return /[\d)!πτ]/.test(source.slice(-1)) || /\b(?:pi|e|tau)$/.test(source);
	}

	function startEditing() {
		if (justEvaluated) {
			expr = '';
			justEvaluated = false;
		}
		error = '';
		displayExpr = null;
	}

	function continueEditing() {
		justEvaluated = false;
		error = '';
		displayExpr = null;
	}

	function pressDigit(digit: string) {
		startEditing();
		if (isValueEnd(expr)) {
			const tail = expr.match(/(\d*\.?\d*)$/)?.[1] ?? '';
			if (tail === '') {
				if (/(?:pi|e|tau)$/.test(expr)) expr += '*';
			} else if (tail === '0') {
				expr = expr.slice(0, -1) + digit;
				return;
			}
		}
		expr += digit;
	}

	function pressDecimal() {
		startEditing();
		const tail = expr.match(/(\d*\.?\d*)$/)?.[1] ?? '';
		if (tail.includes('.')) return;
		if (tail === '') {
			if (/(?:pi|e|tau)$/.test(expr)) expr += '*0.';
			else expr += '0.';
		} else {
			expr += '.';
		}
	}

	function pressOperator(op: string) {
		continueEditing();
		if (expr === '') {
			if (op === '-') expr = '-';
			return;
		}
		const last = expr.slice(-1);
		if ('+-*/^'.includes(last)) {
			if (expr.length === 1) {
				if (op === '-') return;
				expr = '';
				return;
			}
			expr = expr.slice(0, -1) + op;
		} else if (last === '(') {
			if (op === '-') expr += '-';
		} else {
			expr += op;
		}
	}

	function insertFunction(name: string) {
		startEditing();
		if (isValueEnd(expr)) expr += '*';
		expr += `${name}(`;
	}

	function insertConstant(name: string) {
		startEditing();
		if (isValueEnd(expr)) expr += '*';
		expr += name;
	}

	function pressParen(paren: '(' | ')') {
		startEditing();
		if (paren === '(') {
			if (isValueEnd(expr)) expr += '*';
			expr += '(';
			return;
		}
		const opens = (expr.match(/\(/g) ?? []).length;
		const closes = (expr.match(/\)/g) ?? []).length;
		if (opens > closes && !/[+\-*/^(%]$/.test(expr)) expr += ')';
	}

	function applyExponent(power: string) {
		continueEditing();
		if (!isValueEnd(expr)) return;
		expr += `^${power}`;
	}

	function pressFactorial() {
		continueEditing();
		if (isValueEnd(expr)) expr += '!';
	}

	function pressModulo() {
		continueEditing();
		if (isValueEnd(expr) && !expr.endsWith('%')) expr += '%';
	}

	function modOrAbs() {
		if (second) insertFunction('abs');
		else pressModulo();
	}

	function toggleSign() {
		continueEditing();
		if (expr.startsWith('-')) expr = expr.slice(1);
		else expr = `-${expr}`;
	}

	function backspace() {
		error = '';
		displayExpr = null;
		justEvaluated = false;
		expr = expr.slice(0, -1);
		if (!expr) {
			resultText = '0';
			resultValue = 0;
			previewText = null;
		}
	}

	function clear() {
		expr = '';
		displayExpr = null;
		resultText = '0';
		resultValue = 0;
		previewText = null;
		error = '';
		second = false;
		justEvaluated = false;
	}

	function recall(value: number) {
		startEditing();
		if (isValueEnd(expr)) expr += '*';
		expr += toExpressionNumber(value);
	}

	function cycleAngle() {
		const index = angleUnits.indexOf(angleUnit as (typeof angleUnits)[number]);
		const next = angleUnits[(index + 1) % angleUnits.length];
		void settings.update({ angleUnit: next });
	}

	async function equals() {
		if (!expr || busy) return;
		busy = true;
		error = '';
		try {
			const value = await calcService.scientificEvaluate(expr, angleUnit);
			const formatted = formatNumber(value);
			const shown = prettyScientific(expr);
			resultValue = value;
			resultText = formatted;
			previewText = formatted;
			displayExpr = `${shown} =`;
			justEvaluated = true;
			expr = toExpressionNumber(value);
			pulse += 1;
			void history.add('scientific', shown, formatted);
		} catch (err) {
			error = errorMessage(err);
			justEvaluated = false;
		} finally {
			busy = false;
		}
	}

	$effect(() => {
		const source = expr;
		const unit = angleUnit;
		if (!source || justEvaluated) return;

		const token = ++previewToken;
		const timer = setTimeout(async () => {
			try {
				const value = await calcService.scientificEvaluate(source, unit);
				if (token === previewToken) previewText = formatNumber(value);
			} catch {
				if (token === previewToken) previewText = null;
			}
		}, 160);

		return () => {
			clearTimeout(timer);
			if (token === previewToken) previewToken++;
		};
	});

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
		} else if (key === '^') {
			pressOperator('^');
		} else if (key === '%') {
			pressModulo();
		} else if (key === '!') {
			pressFactorial();
		} else if (key === '(' || key === ')') {
			pressParen(key);
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
		badge={second ? '2nd' : null}
		depth={parenDepth}
		{pulse}
		value={resultValue}
	/>

	<MemoryBar current={resultValue} onrecall={recall} />

	<Keypad columns={5}>
		<Key
			label="2nd"
			variant="function"
			active={second}
			onpress={() => (second = !second)}
			ariaLabel="Second function"
		/>
		<Key
			label={angleUnit.toUpperCase()}
			variant="function"
			onpress={cycleAngle}
			ariaLabel="Angle unit: {angleUnit}"
			title="Cycle degrees, radians and gradians"
		/>
		<Key
			label={second ? 'sin⁻¹' : 'sin'}
			variant="function"
			onpress={() => insertFunction(second ? 'asin' : 'sin')}
		/>
		<Key
			label={second ? 'cos⁻¹' : 'cos'}
			variant="function"
			onpress={() => insertFunction(second ? 'acos' : 'cos')}
		/>
		<Key
			label={second ? 'tan⁻¹' : 'tan'}
			variant="function"
			onpress={() => insertFunction(second ? 'atan' : 'tan')}
		/>

		<Key
			label={second ? 'eˣ' : 'ln'}
			variant="function"
			onpress={() => insertFunction(second ? 'exp' : 'ln')}
		/>
		<Key
			label={second ? '10ˣ' : 'log'}
			variant="function"
			onpress={() => insertFunction(second ? 'exp10' : 'log')}
		/>
		<Key
			label={second ? 'x³' : 'x²'}
			variant="function"
			onpress={() => applyExponent(second ? '3' : '2')}
		/>
		<Key label="xʸ" variant="function" onpress={() => pressOperator('^')} />
		<Key
			label={second ? '∛' : '√'}
			variant="function"
			onpress={() => insertFunction(second ? 'cbrt' : 'sqrt')}
		/>

		<Key
			label={second ? 'τ' : 'π'}
			variant="function"
			onpress={() => insertConstant(second ? 'tau' : 'pi')}
		/>
		<Key label="e" variant="function" onpress={() => insertConstant('e')} />
		<Key label="(" variant="function" onpress={() => pressParen('(')} />
		<Key label=")" variant="function" onpress={() => pressParen(')')} />
		<Key label="⌫" variant="function" ariaLabel="Backspace" onpress={backspace} />

		<Key label="7" onpress={() => pressDigit('7')} />
		<Key label="8" onpress={() => pressDigit('8')} />
		<Key label="9" onpress={() => pressDigit('9')} />
		<Key label="÷" variant="operator" ariaLabel="Divide" onpress={() => pressOperator('/')} />
		<Key label="AC" variant="danger" ariaLabel="Clear" onpress={clear} />

		<Key label="4" onpress={() => pressDigit('4')} />
		<Key label="5" onpress={() => pressDigit('5')} />
		<Key label="6" onpress={() => pressDigit('6')} />
		<Key label="×" variant="operator" ariaLabel="Multiply" onpress={() => pressOperator('*')} />
		<Key label="!" variant="function" ariaLabel="Factorial" onpress={pressFactorial} />

		<Key label="1" onpress={() => pressDigit('1')} />
		<Key label="2" onpress={() => pressDigit('2')} />
		<Key label="3" onpress={() => pressDigit('3')} />
		<Key label="−" variant="operator" ariaLabel="Subtract" onpress={() => pressOperator('-')} />
		<Key
			label={second ? 'abs' : 'mod'}
			variant="function"
			onpress={modOrAbs}
			ariaLabel={second ? 'Absolute value' : 'Modulo'}
		/>

		<Key label="0" onpress={() => pressDigit('0')} />
		<Key label="." ariaLabel="Decimal point" onpress={pressDecimal} />
		<Key label="±" variant="function" ariaLabel="Toggle sign" onpress={toggleSign} />
		<Key label="+" variant="operator" ariaLabel="Add" onpress={() => pressOperator('+')} />
		<Key label="=" variant="equals" ariaLabel="Equals" onpress={() => void equals()} />
	</Keypad>
</div>

<style>
	.mode {
		display: flex;
		flex-direction: column;
		gap: 9px;
	}
</style>
