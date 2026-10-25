import { describe, expect, it } from 'vitest';
import { evaluate, formatNumber, toExpressionNumber } from '#lib/eval/normal';

describe('evaluate', () => {
	const cases: Array<[string, number]> = [
		['1+2', 3],
		['2*3+4', 10],
		['2+3*4', 14],
		['(2+3)*4', 20],
		['10/4', 2.5],
		['-2*3', -6],
		['2*-3', -6],
		['2--3', 5],
		['2 × 3 − 4 ÷ 2', 4],
		['0.1+0.2', 0.3],
		['50%', 0.5],
		['200+10%', 220],
		['200-10%', 180],
		['200*10%', 20],
		['200/10%', 2000],
		['10%+5', 5.1],
		['1.5*2', 3],
		['1e3', 1000],
		['1.5e-3', 0.0015],
		['2e+3*2', 4000]
	];

	it.each(cases)('evaluates %s', (expr, want) => {
		expect(evaluate(expr)).toBeCloseTo(want, 10);
	});

	it.each(['1+', '(1+2', '1/0', '', 'abc', '2.3.4', '()'])(
		'rejects %s',
		(expr) => {
			expect(() => evaluate(expr)).toThrow();
		}
	);
});

describe('formatNumber', () => {
	it('hides floating point noise', () => {
		expect(formatNumber(0.1 + 0.2)).toBe('0.3');
		expect(formatNumber(1 / 3)).toBe('0.333333333333');
	});

	it('groups thousands for display', () => {
		expect(formatNumber(1234567.891)).toBe('1,234,567.891');
	});

	it('switches to exponential notation for huge values', () => {
		expect(formatNumber(1e13)).toContain('e');
	});

	it('normalises negative zero and uses exponential for tiny values', () => {
		expect(formatNumber(-0)).toBe('0');
		expect(formatNumber(-0.0000000000001)).toBe('-1e-13');
	});

	it('renders plain expression numbers without grouping', () => {
		expect(toExpressionNumber(1234567.891)).toBe('1234567.891');
		expect(toExpressionNumber(1 / 3)).toBe('0.333333333333');
	});

	it('round-trips exponential results back through the parser', () => {
		const text = toExpressionNumber(1e13);
		expect(text).toBe('1e13');
		expect(evaluate(`${text}+1`)).toBe(10000000000001);
	});
});
