import type { NoticeIntent } from '$lib/components/molecules/notice-band/notice-band.types';

/**
 * The single toast system (DESIGN_PLAN §5.3). One store, consumed by exactly one `NoticeBandRegion`
 * organism that renders the `notice-band` molecule. The reference re-implemented toasts on every page;
 * here pages just call `toast.success(...)` etc.
 */
export interface Toast {
	id: number;
	intent: NoticeIntent;
	title?: string;
	message: string;
	dismissible: boolean;
	/** Auto-dismiss after this many ms; `0` keeps the toast until dismissed. */
	duration: number;
}

export interface ToastOptions {
	title?: string;
	duration?: number;
	dismissible?: boolean;
}

const DEFAULT_DURATION = 4500;
/**
 * How many auto-dismissing notices exist at once. A band sits in the page flow, so every extra one
 * pushes the content down; beyond a few at a time the jump costs more than the notice is worth. The
 * cap lives here rather than in the region so a notice can never count down unseen: past this many,
 * the oldest is dismissed — timer and all — the moment a newer one arrives. A notice with
 * `duration: 0` is outside the cap: it was asked to stay until someone dismisses it, and that is
 * exactly the kind that has to be read before it goes.
 */
const MAX_ITEMS = 3;

let items = $state<Toast[]>([]);
let nextId = 0;
// Plain object (not a Map) since these timer handles are bookkeeping, not reactive state.
const timers: Record<number, ReturnType<typeof setTimeout>> = {};

function dismiss(id: number): void {
	const timer = timers[id];
	if (timer) {
		clearTimeout(timer);
		delete timers[id];
	}
	items = items.filter((toastItem) => toastItem.id !== id);
}

function push(intent: NoticeIntent, message: string, options: ToastOptions = {}): number {
	const id = (nextId += 1);
	const duration = options.duration ?? DEFAULT_DURATION;
	const next: Toast = {
		id,
		intent,
		title: options.title,
		message,
		dismissible: options.dismissible ?? true,
		duration
	};
	// Newest on top, and the oldest beyond the cap goes rather than waiting for a turn it would
	// have spent its timer on. Only the ones that count down are counted and dropped — a sticky
	// notice stays until it is dismissed, as its `duration` promises.
	items = [next, ...items];
	const expiring = items.filter((item) => item.duration > 0);
	for (const stale of expiring.slice(MAX_ITEMS)) dismiss(stale.id);
	if (duration > 0) {
		timers[id] = setTimeout(() => dismiss(id), duration);
	}
	return id;
}

function clear(): void {
	for (const key of Object.keys(timers)) {
		clearTimeout(timers[Number(key)]);
		delete timers[Number(key)];
	}
	items = [];
}

const BACKGROUND_HINTS = ['schedul', 'queue', 'pending', 'process', 'background', 'running'];

/**
 * Map a backend/realtime status string to an intent (the reference's "scheduled/queued/background →
 * info" heuristic), then push. Useful for realtime events (e.g. `import.completed`, `bulk_tag.queued`).
 */
function fromStatus(status: string, message: string, options: ToastOptions = {}): number {
	const normalized = status.toLowerCase();
	let intent: NoticeIntent = 'info';
	if (BACKGROUND_HINTS.some((hint) => normalized.includes(hint))) intent = 'info';
	else if (normalized.includes('fail') || normalized.includes('error')) intent = 'error';
	else if (
		normalized.includes('complete') ||
		normalized.includes('success') ||
		normalized.includes('done')
	)
		intent = 'success';
	else if (normalized.includes('warn')) intent = 'warning';
	return push(intent, message, options);
}

/** The toast API. `toast.items` is reactive when read in a component (Svelte 5 cross-module pattern). */
export const toast = {
	get items(): Toast[] {
		return items;
	},
	info: (message: string, options?: ToastOptions) => push('info', message, options),
	success: (message: string, options?: ToastOptions) => push('success', message, options),
	warning: (message: string, options?: ToastOptions) => push('warning', message, options),
	error: (message: string, options?: ToastOptions) => push('error', message, options),
	push,
	dismiss,
	clear,
	fromStatus
};
