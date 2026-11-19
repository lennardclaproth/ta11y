import { decimalStringToNumber } from '$lib/api/money';
import type { AssetSnapshotPoint } from '$lib/api/types';
import { getAssetSnapshots } from '$lib/services/assets';
import { accountStore } from '$lib/stores/account.svelte';
import { periodStore } from '$lib/stores/period.svelte';

/**
 * Net worth for the account overview, read from the total-worth snapshots the API already
 * keeps. No new calculation: the displayed worth is the latest snapshot, and the change is the
 * first snapshot in the selected period against the last one. Fewer than two snapshots inside
 * the period means there is no change to report, which the overview says in words rather than
 * showing a zero that would read as "no movement".
 *
 * The whole series is fetched once rather than per period, because the period is a client-side
 * window over it and the oldest snapshot is also what gives the `Max` preset a real start date.
 */

let snapshots = $state<AssetSnapshotPoint[]>([]);
let loading = $state(false);
let failed = $state(false);
let loadPromise: Promise<void> | null = null;

const inPeriod = $derived(
	snapshots.filter((point) => point.date >= periodStore.from && point.date <= periodStore.to)
);

export const netWorthStore = {
	/** The latest known total worth, or null while there is no snapshot at all. */
	get worth(): number | null {
		const latest = snapshots.at(-1);
		return latest ? decimalStringToNumber(latest.total_worth) : null;
	},
	/** Percentage change across the selected period, or null when the period has no movement to report. */
	get changePct(): number | null {
		const first = inPeriod.at(0);
		const last = inPeriod.at(-1);
		if (!first || !last || first === last) return null;
		const start = decimalStringToNumber(first.total_worth);
		const end = decimalStringToNumber(last.total_worth);
		if (start === 0) return null;
		return ((end - start) / Math.abs(start)) * 100;
	},
	get loading(): boolean {
		return loading;
	},
	get failed(): boolean {
		return failed;
	},
	/** Fetch the series once. Idempotent; safe to call from every page. */
	ensureLoaded(): Promise<void> {
		if (!loadPromise) loadPromise = load();
		return loadPromise;
	},
	/** Drop the cached series so the next `ensureLoaded` refetches it. */
	reset(): void {
		snapshots = [];
		failed = false;
		loadPromise = null;
	}
};

async function load(): Promise<void> {
	loading = true;
	failed = false;
	try {
		await accountStore.ensureLoaded();
		// No session yet is an empty overview, not a failure: an account-scoped request
		// would be refused and read as one.
		if (!accountStore.hasAccount) {
			snapshots = [];
			return;
		}
		const points = await getAssetSnapshots({});
		snapshots = [...points].sort((a, b) => a.date.localeCompare(b.date));
		periodStore.setEarliest(snapshots.at(0)?.date ?? null);
	} catch {
		// A missing net worth is supplementary information; the page it sits above still works.
		failed = true;
		snapshots = [];
	} finally {
		loading = false;
	}
}
