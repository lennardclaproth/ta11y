import { browser } from '$app/environment';
import { apiBaseUrl, useMocks } from '$lib/api/config';

/**
 * Realtime refresh (DESIGN_PLAN §5.7). Subscribes to the account WebSocket (`/ws/accounts/{id}`) and
 * debounces an `onRefresh` callback (~250ms) when relevant backend events arrive. No-ops on the server
 * and in mock mode, so the portal runs standalone without a WS backend.
 *
 * The socket only stays open while someone is looking: it closes when the tab is hidden, and the
 * server closes it after a stretch without updates. It reopens when the tab becomes visible again
 * or on the next pointer/key interaction — never on a timer — so an idle tab holds no connection and
 * the API can scale to zero. Every reopen triggers one refresh to pick up events missed meanwhile.
 */
/** Known events: `assets.rebuilt`, `import.completed`, `bulk_tag.completed`, `portfolio.rebuilt`. */
export type RealtimeEvent = string;

export interface RealtimeOptions {
	accountId: string;
	onRefresh: () => void;
	onEvent?: (event: string, payload: unknown) => void;
	/** Only react to these event types (defaults to all). */
	events?: RealtimeEvent[];
	debounceMs?: number;
}

export interface RealtimeConnection {
	disconnect: () => void;
}

/** Minimum gap between reconnect attempts, so a refused socket isn't retried on every keystroke. */
const RECONNECT_MIN_INTERVAL_MS = 5_000;

const INTERACTION_EVENTS = ['pointerdown', 'keydown'] as const;

/** The API base as a ws(s) URL; a relative base (`/api`) resolves against the page's origin. */
function wsBaseUrl(): string {
	const url = new URL(apiBaseUrl, window.location.href);
	url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
	return url.toString().replace(/\/+$/, '');
}

export function connectRealtime(options: RealtimeOptions): RealtimeConnection {
	// Standalone / SSR: nothing to connect to.
	if (!browser || useMocks || !apiBaseUrl) {
		return { disconnect() {} };
	}

	const debounceMs = options.debounceMs ?? 250;
	const url = `${wsBaseUrl()}/ws/accounts/${options.accountId}`;
	let timer: ReturnType<typeof setTimeout> | undefined;
	let socket: WebSocket | undefined;
	let lastAttemptAt = 0;
	let stopped = false;

	function scheduleRefresh() {
		if (timer) clearTimeout(timer);
		timer = setTimeout(() => options.onRefresh(), debounceMs);
	}

	function onMessage(message: MessageEvent) {
		let parsed: { type?: string; event?: string };
		try {
			parsed = JSON.parse(message.data) as { type?: string; event?: string };
		} catch {
			return;
		}
		const event = parsed.type ?? parsed.event ?? '';
		if (options.events && !options.events.includes(event)) return;
		options.onEvent?.(event, parsed);
		scheduleRefresh();
	}

	function open(isReconnect: boolean) {
		if (stopped || socket || document.visibilityState !== 'visible') return;
		const now = Date.now();
		if (isReconnect && now - lastAttemptAt < RECONNECT_MIN_INTERVAL_MS) return;
		lastAttemptAt = now;

		const current = new WebSocket(url);
		socket = current;
		current.addEventListener('message', onMessage);
		current.addEventListener('close', () => {
			if (socket === current) socket = undefined;
		});
		if (isReconnect) {
			current.addEventListener('open', scheduleRefresh, { once: true });
		}
	}

	function close() {
		const current = socket;
		socket = undefined;
		try {
			current?.close();
		} catch {
			// already closing/closed
		}
	}

	function onVisibilityChange() {
		if (document.visibilityState === 'hidden') close();
		else open(true);
	}

	function onInteraction() {
		open(true);
	}

	document.addEventListener('visibilitychange', onVisibilityChange);
	for (const type of INTERACTION_EVENTS) {
		window.addEventListener(type, onInteraction, { passive: true, capture: true });
	}
	open(false);

	return {
		disconnect() {
			stopped = true;
			if (timer) clearTimeout(timer);
			document.removeEventListener('visibilitychange', onVisibilityChange);
			for (const type of INTERACTION_EVENTS) {
				window.removeEventListener(type, onInteraction, { capture: true });
			}
			close();
		}
	};
}
