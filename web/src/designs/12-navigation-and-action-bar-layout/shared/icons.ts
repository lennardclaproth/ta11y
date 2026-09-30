import type { IconifyIcon } from '@iconify/svelte';

/**
 * Inline copies of the heroicons used by these prototypes, following the precedent in
 * LedgerToolbar: bundling the control glyphs keeps the screenshots deterministic when the
 * Iconify API is unreachable. Production code keeps using the string ids.
 */
function outline(path: string): IconifyIcon {
	return {
		width: 24,
		height: 24,
		body: `<path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="${path}"/>`
	};
}

export const icons = {
	plus: outline('M12 4.5v15m7.5-7.5h-15'),
	upload: outline(
		'M12 16.5V9.75m0 0 3 3m-3-3-3 3M6.75 19.5a4.5 4.5 0 0 1-.41-8.98 4.5 4.5 0 0 1 8.08-3.015 4.5 4.5 0 0 1 5.58 5.995'
	),
	rebuild: outline(
		'M16.023 9.348h4.992V4.356m0 4.992-3.181-3.183a8.25 8.25 0 0 0-13.803 3.7M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7'
	),
	chevronDown: outline('m19.5 8.25-7.5 7.5-7.5-7.5'),
	search: outline('m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z'),
	tag: outline(
		'M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.878.548 2.374-.315l3.548-6.187c.343-.598.244-1.35-.243-1.837L11.16 3.66A2.25 2.25 0 0 0 9.568 3ZM6 6h.008v.008H6V6Z'
	),
	signOut: outline(
		'M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15M12 9l-3 3m0 0 3 3m-3-3h12.75'
	),
	calendar: outline(
		'M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5'
	),
	menu: outline('M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5')
};
