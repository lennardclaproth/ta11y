import { browser } from '$app/environment';

/**
 * Admin-mode toggle (DESIGN_PLAN §3.5). A client-side flag (the reference guarded /admin routes on an
 * admin-mode store), persisted to localStorage. The account menu binds to it; pages guard /admin routes
 * client-side with an explicit enable-admin screen in the admin layout.
 */
const STORAGE_KEY = 'mft.adminMode';

let enabled = $state(browser ? localStorage.getItem(STORAGE_KEY) === 'true' : false);

function persist() {
	if (browser) localStorage.setItem(STORAGE_KEY, String(enabled));
}

export const adminMode = {
	get enabled(): boolean {
		return enabled;
	},
	set(value: boolean) {
		enabled = value;
		persist();
	},
	toggle() {
		enabled = !enabled;
		persist();
	}
};
