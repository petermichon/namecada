const KEY = 'domain-browser:starred';

function load(): string[] {
	if (typeof localStorage === 'undefined') return [];
	try {
		const raw = JSON.parse(localStorage.getItem(KEY) ?? '[]');
		return Array.isArray(raw) ? raw.filter((x): x is string => typeof x === 'string') : [];
	} catch {
		return [];
	}
}

function save(items: string[]) {
	try {
		localStorage.setItem(KEY, JSON.stringify(items));
	} catch {
		// localStorage unavailable or quota exceeded: keep in-memory state
	}
}

class StarStore {
	items = $state<string[]>(load());

	has(name: string): boolean {
		return this.items.includes(name);
	}

	toggle(name: string) {
		this.items = this.items.includes(name)
			? this.items.filter((n) => n !== name)
			: [...this.items, name];
		save(this.items);
	}
}

export const stars = new StarStore();
