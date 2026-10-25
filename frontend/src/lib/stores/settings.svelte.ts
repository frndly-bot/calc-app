import { settingsService, type Settings } from '#lib/bindings';

export const defaultSettings: Settings = {
	theme: 'dark',
	mode: 'normal',
	angleUnit: 'deg',
	devBits: 32,
	devSigned: false,
	currency: 'USD'
};

function applyTheme(theme: string) {
	if (typeof document !== 'undefined') {
		document.documentElement.dataset.theme = theme;
	}
}

class SettingsStore {
	current = $state<Settings>({ ...defaultSettings });

	async load() {
		try {
			this.current = { ...defaultSettings, ...(await settingsService.get()) };
		} catch {
			// Browser preview without the Go backend keeps the defaults.
		}
		applyTheme(this.current.theme);
	}

	async update(patch: Partial<Settings>) {
		this.current = { ...this.current, ...patch };
		applyTheme(this.current.theme);
		try {
			await settingsService.save(this.current);
		} catch {
			// Preferences stay in memory when the backend is unavailable.
		}
	}

	toggleTheme() {
		return this.update({ theme: this.current.theme === 'dark' ? 'light' : 'dark' });
	}
}

export const settings = new SettingsStore();
