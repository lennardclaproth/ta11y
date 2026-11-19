import { periodStore, type PeriodPreset } from './period.svelte';

type Period = { from: string; to: string; preset: PeriodPreset };

/**
 * A range dragged on a chart sets the period like any other pick, and the chart that took it says
 * so with a way back to the period it replaced. Picking another period anywhere else makes `active`
 * false by itself. One per chart that zooms.
 */
export function createDraggedRange() {
	let range = $state<{ from: string; to: string; previous: Period } | null>(null);
	const active = $derived(
		range !== null && periodStore.from === range.from && periodStore.to === range.to
	);

	return {
		get active(): boolean {
			return active;
		},
		select(from: string, to: string): void {
			// A drag on top of a drag still goes back to the period before the first one.
			const previous: Period =
				active && range
					? range.previous
					: { from: periodStore.from, to: periodStore.to, preset: periodStore.preset };
			periodStore.set({ from, to });
			range = { from, to, previous };
		},
		reset(): void {
			const previous = range?.previous;
			range = null;
			if (!previous) periodStore.applyPreset('ytd');
			else if (previous.preset === 'custom') periodStore.set(previous);
			else periodStore.applyPreset(previous.preset);
		}
	};
}
