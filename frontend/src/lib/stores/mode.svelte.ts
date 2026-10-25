import { settings } from '#lib/stores/settings.svelte';

export type Mode = 'normal' | 'scientific' | 'developer' | 'finance';

export interface ModeInfo {
	id: Mode;
	label: string;
	accent: string;
	hint: string;
}

export const modes: ModeInfo[] = [
	{ id: 'normal', label: 'Normal', accent: '#3b82f6', hint: 'Everyday arithmetic' },
	{ id: 'scientific', label: 'Scientific', accent: '#8b5cf6', hint: 'Trig, logs and powers' },
	{ id: 'developer', label: 'Developer', accent: '#10b981', hint: 'Bases and bitwise math' },
	{ id: 'finance', label: 'Finance', accent: '#f59e0b', hint: 'Loans, savings and ROI' }
];

class ModeStore {
	current = $state<Mode>('normal');

	get index() {
		return modes.findIndex((m) => m.id === this.current);
	}

	set(next: Mode, persist = true) {
		if (this.current === next) return;
		this.current = next;
		if (persist) void settings.update({ mode: next });
	}
}

export const mode = new ModeStore();

export function accentFor(id: string): string {
	return modes.find((m) => m.id === id)?.accent ?? modes[0].accent;
}

export function modeLabel(id: string): string {
	return modes.find((m) => m.id === id)?.label ?? id;
}
