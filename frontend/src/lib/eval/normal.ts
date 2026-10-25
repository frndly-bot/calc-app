type Token =
	| { type: 'number'; value: number; pos: number }
	| { type: 'op'; op: string; pos: number }
	| { type: 'lparen'; pos: number }
	| { type: 'rparen'; pos: number };

type Node =
	| { type: 'number'; value: number }
	| { type: 'percent'; value: Node }
	| { type: 'unary'; op: '-' | '+'; value: Node }
	| { type: 'binary'; op: string; left: Node; right: Node };

const POSTFIX_PREC = 3;

function lex(src: string): Token[] {
	const tokens: Token[] = [];
	let i = 0;
	while (i < src.length) {
		const c = src[i];
		if (c === ' ' || c === '\t' || c === '\n' || c === '\r') {
			i++;
			continue;
		}
		if ((c >= '0' && c <= '9') || c === '.') {
			const start = i;
			let dots = 0;
			while (i < src.length && ((src[i] >= '0' && src[i] <= '9') || src[i] === '.')) {
				if (src[i] === '.') dots++;
				i++;
			}
			if (dots > 1 || src.slice(start, i) === '.') {
				throw new Error(`Invalid number at position ${start + 1}`);
			}
			if (i < src.length && (src[i] === 'e' || src[i] === 'E')) {
				let end = i + 1;
				if (end < src.length && (src[end] === '+' || src[end] === '-')) end++;
				let digits = 0;
				while (end < src.length && src[end] >= '0' && src[end] <= '9') {
					end++;
					digits++;
				}
				if (digits > 0) i = end;
			}
			const value = Number(src.slice(start, i));
			if (!Number.isFinite(value)) {
				throw new Error(`Invalid number at position ${start + 1}`);
			}
			tokens.push({ type: 'number', value, pos: start });
			continue;
		}
		if (c === '(') {
			tokens.push({ type: 'lparen', pos: i++ });
			continue;
		}
		if (c === ')') {
			tokens.push({ type: 'rparen', pos: i++ });
			continue;
		}
		if (c === '%') {
			tokens.push({ type: 'op', op: '%', pos: i++ });
			continue;
		}
		const glyph = c === '×' ? '*' : c === '÷' ? '/' : c === '−' ? '-' : c;
		if (glyph === '+' || glyph === '-' || glyph === '*' || glyph === '/') {
			tokens.push({ type: 'op', op: glyph, pos: i });
			i += c.length;
			continue;
		}
		throw new Error(`Unexpected character "${c}" at position ${i + 1}`);
	}
	return tokens;
}

function opPrec(op: string): number {
	if (op === '+' || op === '-') return 1;
	if (op === '*' || op === '/') return 2;
	return 0;
}

class Parser {
	private pos = 0;

	constructor(private readonly tokens: Token[]) {}

	parse(): Node {
		const node = this.expr(1);
		if (this.pos < this.tokens.length) {
			throw new Error('Unexpected input');
		}
		return node;
	}

	private peek(): Token | undefined {
		return this.tokens[this.pos];
	}

	private next(): Token | undefined {
		return this.tokens[this.pos++];
	}

	private expr(minPrec: number): Node {
		let left = this.unary();
		for (;;) {
			const token = this.peek();
			if (!token || token.type !== 'op') break;

			if (token.op === '%') {
				if (POSTFIX_PREC < minPrec) break;
				this.next();
				left = { type: 'percent', value: left };
				continue;
			}

			const prec = opPrec(token.op);
			if (prec < minPrec) break;
			this.next();
			const right = this.expr(prec + 1);
			left = { type: 'binary', op: token.op, left, right };
		}
		return left;
	}

	private unary(): Node {
		const token = this.peek();
		if (token?.type === 'op' && (token.op === '-' || token.op === '+')) {
			this.next();
			const value = this.unary();
			return token.op === '-' ? { type: 'unary', op: '-', value } : value;
		}
		return this.primary();
	}

	private primary(): Node {
		const token = this.next();
		if (!token) throw new Error('Unexpected end of expression');
		if (token.type === 'number') return { type: 'number', value: token.value };
		if (token.type === 'lparen') {
			const node = this.expr(1);
			const close = this.next();
			if (close?.type !== 'rparen') throw new Error('Missing closing parenthesis');
			return node;
		}
		throw new Error('Expected a number');
	}
}

function evalNode(node: Node): number {
	switch (node.type) {
		case 'number':
			return node.value;
		case 'percent':
			return evalNode(node.value) / 100;
		case 'unary':
			return node.op === '-' ? -evalNode(node.value) : evalNode(node.value);
		case 'binary': {
			const left = evalNode(node.left);
			const right = evalNode(node.right);
			return applyBinary(node.op, left, right, node.right);
		}
	}
}

function applyBinary(op: string, left: number, right: number, rightNode: Node): number {
	if ((op === '+' || op === '-') && rightNode.type === 'percent') {
		// "200 + 10%" means "200 + 10% of 200".
		const delta = left * right;
		return op === '+' ? left + delta : left - delta;
	}
	switch (op) {
		case '+':
			return left + right;
		case '-':
			return left - right;
		case '*':
			return left * right;
		case '/':
			if (right === 0) throw new Error("Can't divide by zero");
			return left / right;
	}
	throw new Error(`Unknown operator ${op}`);
}

export function evaluate(expr: string): number {
	const tokens = lex(expr);
	if (tokens.length === 0) throw new Error('Empty expression');
	const value = evalNode(new Parser(tokens).parse());
	if (!Number.isFinite(value)) throw new Error('Result is out of range');
	return Object.is(value, -0) ? 0 : value;
}

export interface FormatOptions {
	group?: boolean;
	significant?: number;
}

export function formatNumber(value: number, options: FormatOptions = {}): string {
	if (!Number.isFinite(value)) return 'Error';
	const significant = options.significant ?? 12;
	let rounded = Number(value.toPrecision(significant));
	if (Object.is(rounded, -0)) rounded = 0;

	const abs = Math.abs(rounded);
	if (abs !== 0 && (abs >= 1e12 || abs < 1e-9)) {
		return rounded
			.toExponential(significant - 1)
			.replace(/(\.\d*?)0+e/, '$1e')
			.replace(/\.e/, 'e')
			.replace('e+', 'e');
	}

	const integerDigits =
		abs >= 1 ? Math.floor(Math.log10(abs)) + 1 : 1;
	const fractionDigits =
		abs >= 1
			? Math.max(0, significant - integerDigits)
			: significant + Math.max(0, -Math.floor(Math.log10(abs)) - 1);

	return new Intl.NumberFormat('en-US', {
		useGrouping: options.group ?? true,
		maximumFractionDigits: Math.min(100, fractionDigits),
		minimumFractionDigits: 0
	}).format(rounded);
}

export function toExpressionNumber(value: number): string {
	return formatNumber(value, { group: false });
}
