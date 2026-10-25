import { Call } from '@wailsio/runtime';

const PACKAGE = 'calc-app/backend/services';

function invoke<T>(service: string, method: string, args: unknown[]): Promise<T> {
	return Call.ByName(`${PACKAGE}.${service}.${method}`, ...args) as Promise<T>;
}

export interface Formatted {
	Hex: string;
	Dec: string;
	Oct: string;
	Bin: string;
}

export interface DevResult {
	raw: string;
	value: number;
	formatted: Formatted;
}

export interface HistoryEntry {
	id: number;
	mode: string;
	expr: string;
	result: string;
	createdAt: string;
}

export interface Settings {
	theme: string;
	mode: string;
	angleUnit: string;
	devBits: number;
	devSigned: boolean;
	currency: string;
}

export interface Payment {
	number: number;
	payment: number;
	principal: number;
	interest: number;
	balance: number;
}

export interface LoanSummary {
	paymentPerPeriod: number;
	numPeriods: number;
	totalInterest: number;
	totalPaid: number;
	schedule: Payment[];
}

export interface SavingsResult {
	futureValue: number;
	totalDeposited: number;
	totalInterest: number;
}

export const calcService = {
	scientificEvaluate(expr: string, angleMode: string) {
		return invoke<number>('CalcService', 'ScientificEvaluate', [expr, angleMode]);
	},
	devParse(text: string, base: number, bits: number, signed: boolean) {
		return invoke<DevResult>('CalcService', 'DevParse', [text, base, bits, signed]);
	},
	devBinary(op: string, a: string, b: string, bits: number, signed: boolean) {
		return invoke<DevResult>('CalcService', 'DevBinary', [op, a, b, bits, signed]);
	},
	devUnary(op: string, a: string, bits: number, signed: boolean) {
		return invoke<DevResult>('CalcService', 'DevUnary', [op, a, bits, signed]);
	},
	devFormat(raw: string, bits: number, signed: boolean) {
		return invoke<DevResult>('CalcService', 'DevFormat', [raw, bits, signed]);
	},
	calcLoan(principal: number, annualRate: number, years: number, paymentsPerYear: number) {
		return invoke<LoanSummary>('CalcService', 'CalcLoan', [
			principal,
			annualRate,
			years,
			paymentsPerYear
		]);
	},
	calcSavings(monthlyDeposit: number, annualRate: number, years: number) {
		return invoke<SavingsResult>('CalcService', 'CalcSavings', [
			monthlyDeposit,
			annualRate,
			years
		]);
	},
	calcROI(initial: number, final: number) {
		return invoke<number>('CalcService', 'CalcROI', [initial, final]);
	},
	calcCAGR(initial: number, final: number, years: number) {
		return invoke<number>('CalcService', 'CalcCAGR', [initial, final, years]);
	},
	calcPercentChange(from: number, to: number) {
		return invoke<number>('CalcService', 'CalcPercentChange', [from, to]);
	}
};

export const historyService = {
	list() {
		return invoke<HistoryEntry[]>('HistoryService', 'List', []);
	},
	add(mode: string, expr: string, result: string) {
		return invoke<HistoryEntry>('HistoryService', 'Add', [mode, expr, result]);
	},
	clear() {
		return invoke<void>('HistoryService', 'Clear', []);
	},
	remove(id: number) {
		return invoke<void>('HistoryService', 'Delete', [id]);
	}
};

export const settingsService = {
	get() {
		return invoke<Settings>('SettingsService', 'Get', []);
	},
	save(settings: Settings) {
		return invoke<void>('SettingsService', 'Save', [settings]);
	}
};

export function errorMessage(error: unknown): string {
	const message = error instanceof Error ? error.message : String(error);
	return message.replace(/^RuntimeError:\s*/, '').replace(/^Error:\s*/, '');
}
