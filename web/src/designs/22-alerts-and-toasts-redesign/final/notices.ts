import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';

/**
 * Shape of one queued notification. Mirrors the existing toast store item
 * (`web/src/lib/stores/toast.svelte.ts`) — this design does not change it.
 */
export interface NoticeItem {
	id: string;
	intent: AlertIntent;
	title?: string;
	message: string;
	dismissible?: boolean;
}

/** Obviously fake notifications, reusing the wording the app already produces. */
export const singleNotice: NoticeItem[] = [
	{ id: 'n1', intent: 'success', message: 'Transaction created' }
];

export const stackedNotices: NoticeItem[] = [
	{ id: 'n1', intent: 'success', message: 'Transaction created' },
	{ id: 'n2', intent: 'info', message: 'Portfolio rebuild started' }
];
