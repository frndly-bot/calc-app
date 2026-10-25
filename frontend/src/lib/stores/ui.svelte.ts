class UiStore {
	historyOpen = $state(false);

	openHistory() {
		this.historyOpen = true;
	}

	closeHistory() {
		this.historyOpen = false;
	}
}

export const ui = new UiStore();
