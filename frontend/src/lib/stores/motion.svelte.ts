class MotionStore {
	reduced = $state(false);

	constructor() {
		if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return;
		const query = window.matchMedia('(prefers-reduced-motion: reduce)');
		this.reduced = query.matches;
		query.addEventListener('change', (event) => {
			this.reduced = event.matches;
		});
	}
}

export const motion = new MotionStore();
