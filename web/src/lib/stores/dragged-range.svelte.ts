import { periodStore, type PeriodPreset } from './period.svelte';

/**
 * A range dragged on a chart sets the period like any other pick, and the chart that took it says
 * so with a way back to the preset it replaced. Picking another period anywhere else makes `active`
 * false by itself. One per chart that zooms.
 */
export function createDraggedRange() {
	let range = $state<{ from: string; to: string; previous: PeriodPreset } | null>(null);
	const active = $derived(
		range !== null && periodStore.from === range.from && periodStore.to === range.to
	);

	return {
		get active(): boolean {
			return active;
		},
		select(from: string, to: string): void {
			// A drag on top of a drag still goes back to the preset before the first one.
			const previous = active && range ? range.previous : periodStore.preset;
			periodStore.set({ from, to });
			range = { from, to, previous };
		},
		reset(): void {
			const previous = range?.previous ?? 'ytd';
			range = null;
			periodStore.applyPreset(previous === 'custom' ? 'ytd' : previous);
		}
	};
}
