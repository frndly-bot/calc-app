class MemoryStore {
	value = $state(0);

	get hasValue() {
		return this.value !== 0;
	}

	clear() {
		this.value = 0;
	}

	recall(): number {
		return this.value;
	}

	add(amount: number) {
		if (Number.isFinite(amount)) this.value += amount;
	}

	subtract(amount: number) {
		if (Number.isFinite(amount)) this.value -= amount;
	}
}

export const memory = new MemoryStore();
