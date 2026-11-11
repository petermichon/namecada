import type { CheckClient, CheckEvent } from './api';

// inTauri reports whether the page runs inside the packaged app. The desktop
// shell injects __TAURI_INTERNALS__ before the bundle loads.
export function inTauri(): boolean {
	return typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window;
}

// tauriCheckClient streams the same manifest → result* → done events as the
// HTTP client, but over IPC: the command pushes each event onto a channel.
export const tauriCheckClient: CheckClient = {
	async stream(label, onEvent, signal) {
		const { Channel, invoke } = await import('@tauri-apps/api/core');
		if (signal?.aborted) return;

		const channel = new Channel<CheckEvent>();
		channel.onmessage = (event) => {
			if (!signal?.aborted) onEvent(event);
		};
		await invoke('check', { label, onEvent: channel });
	}
};
