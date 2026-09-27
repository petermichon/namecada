import type { Attachment } from 'svelte/attachments';

export interface FocusEnv {
	connected: boolean;
	activeIsBody: boolean;
	pointerFine: boolean;
}

export function shouldAutofocus(env: FocusEnv): boolean {
	return env.connected && env.pointerFine && env.activeIsBody;
}

export const autofocus: Attachment<HTMLElement> = (node) => {
	const env: FocusEnv = {
		connected: node.isConnected,
		activeIsBody: document.activeElement === null || document.activeElement === document.body,
		pointerFine: window.matchMedia('(pointer: fine)').matches
	};
	if (shouldAutofocus(env)) node.focus({ preventScroll: true });
};
