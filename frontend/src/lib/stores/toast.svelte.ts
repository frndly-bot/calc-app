class ToastStore {
	message = $state<string | null>(null);
	private timer: ReturnType<typeof setTimeout> | null = null;

	show(message: string) {
		this.message = message;
		if (this.timer) clearTimeout(this.timer);
		this.timer = setTimeout(() => {
			this.message = null;
			this.timer = null;
		}, 4200);
	}

	dismiss() {
		if (this.timer) clearTimeout(this.timer);
		this.timer = null;
		this.message = null;
	}
}

export const toast = new ToastStore();
