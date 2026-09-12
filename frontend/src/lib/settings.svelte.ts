import type { LetterMode, Mode } from './search';

const KEY = 'domain-browser:settings';

interface SettingsData {
	mode: Mode;
	onlyAvailable: boolean;
	query: string;
	letters: string[];
	letterMode: LetterMode;
}

const DEFAULTS: SettingsData = {
	mode: 'prefix',
	onlyAvailable: true,
	query: '',
	letters: [],
	letterMode: 'include'
};

function load(): SettingsData {
	if (typeof localStorage === 'undefined') return { ...DEFAULTS, letters: [] };
	try {
		const raw = JSON.parse(localStorage.getItem(KEY) ?? '{}');
		return {
			mode: raw.mode === 'contains' ? 'contains' : raw.mode === 'postfix' ? 'postfix' : 'prefix',
			onlyAvailable: raw.onlyAvailable !== false,
			query: typeof raw.query === 'string' ? raw.query : '',
			letters: Array.isArray(raw.letters)
				? raw.letters.filter((x: unknown): x is string => typeof x === 'string' && x.length === 1)
				: [],
			letterMode: raw.letterMode === 'exclude' ? 'exclude' : 'include'
		};
	} catch {
		return { ...DEFAULTS, letters: [] };
	}
}

class SettingsStore {
	mode = $state<Mode>(DEFAULTS.mode);
	onlyAvailable = $state<boolean>(DEFAULTS.onlyAvailable);
	query = $state<string>(DEFAULTS.query);
	letters = $state<string[]>([...DEFAULTS.letters]);
	letterMode = $state<LetterMode>(DEFAULTS.letterMode);

	constructor(init: SettingsData) {
		this.mode = init.mode;
		this.onlyAvailable = init.onlyAvailable;
		this.query = init.query;
		this.letters = init.letters;
		this.letterMode = init.letterMode;
	}
}

export const settings = new SettingsStore(load());

$effect.root(() => {
	$effect(() => {
		const data: SettingsData = {
			mode: settings.mode,
			onlyAvailable: settings.onlyAvailable,
			query: settings.query,
			letters: settings.letters,
			letterMode: settings.letterMode
		};
		try {
			localStorage.setItem(KEY, JSON.stringify(data));
		} catch {
			// ignore storage errors
		}
	});
});
