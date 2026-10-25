import { errorMessage, historyService, type HistoryEntry } from '#lib/bindings';
import { toast } from '#lib/stores/toast.svelte';

class HistoryStore {
	entries = $state<HistoryEntry[]>([]);
	private pending: Promise<void> | null = null;

	async load() {
		if (this.pending) return this.pending;
		this.pending = (async () => {
			try {
				this.entries = await historyService.list();
			} catch {
				// The backend is unavailable in a plain browser preview.
			} finally {
				this.pending = null;
			}
		})();
		return this.pending;
	}

	async add(mode: string, expr: string, result: string) {
		try {
			const entry = await historyService.add(mode, expr, result);
			this.entries = [entry, ...this.entries].slice(0, 200);
		} catch (error) {
			toast.show(errorMessage(error));
		}
	}

	async remove(id: number) {
		const previous = this.entries;
		this.entries = this.entries.filter((entry) => entry.id !== id);
		try {
			await historyService.remove(id);
		} catch (error) {
			this.entries = previous;
			toast.show(errorMessage(error));
		}
	}

	async clear() {
		this.entries = [];
		try {
			await historyService.clear();
		} catch (error) {
			toast.show(errorMessage(error));
			await this.load();
		}
	}
}

export const history = new HistoryStore();
