<script lang="ts">
	// Variant E — "Card rail": the analytics band becomes one horizontally scrolling row of cards.
	// The three Cashflow charts keep their place at the start; the goal, the streak and the monthly
	// standing are three more cards you scroll (or swipe) to. The ledger below never changes.
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import PortalFrame from '../shared/PortalFrame.svelte';
	import NetTrendChart from '../shared/NetTrendChart.svelte';
	import TagDonut from '../shared/TagDonut.svelte';
	import TransactionsPanel from '../shared/TransactionsPanel.svelte';
	import GoalSettingPanel from '../shared/GoalSettingPanel.svelte';
	import RunningMonthPanel from '../shared/RunningMonthPanel.svelte';
	import StreakPanel from '../shared/StreakPanel.svelte';
	import NoGoalPanel from '../shared/NoGoalPanel.svelte';
	import MonthStandingRow from '../shared/MonthStandingRow.svelte';
	import SetGoalDialog from '../shared/SetGoalDialog.svelte';
	import { currentGoalPercent, goalAmount, goalMonths } from '../goal-data';

	type Props = {
		/** Which card the rail is scrolled to when the page opens. */
		startAt?: 'trend' | 'month' | 'goal';
		goalSet?: boolean;
	};

	let { startAt = 'month', goalSet = true }: Props = $props();

	let rail = $state<HTMLDivElement | null>(null);
	let goalDialog = $state(false);
	let goalPercent = $state(currentGoalPercent);
	let selectedIds = $state<string[]>([]);

	const running = goalMonths[0];
	const past = goalMonths.slice(1);

	// The prototype opens on a given card so a screenshot shows the goal half of the rail; in the
	// app the rail simply starts at the left. The second pass runs after the charts have taken
	// their final width, which otherwise moves the card out from under the scroll position.
	$effect(() => {
		const track = rail;
		if (!track) return;
		const align = () => {
			const card = track.querySelector(`[data-card="${startAt}"]`);
			if (card instanceof HTMLElement) track.scrollLeft = card.offsetLeft - track.offsetLeft;
		};
		align();
		const timer = setTimeout(align, 400);
		// A viewport change re-snaps the rail; the screenshot run resizes, so hold the position.
		window.addEventListener('resize', align);
		return () => {
			clearTimeout(timer);
			window.removeEventListener('resize', align);
		};
	});

	function nudge(direction: 1 | -1) {
		rail?.scrollBy({ left: direction * 360, behavior: 'smooth' });
	}

	// One width vocabulary for the rail: a full-width card on a phone, a fixed column on desktop.
	const cardClass = 'w-[82vw] shrink-0 snap-start sm:w-[20rem]';
	const wideCardClass = 'w-[82vw] shrink-0 snap-start sm:w-[20rem] lg:w-[32rem]';
</script>

<PortalFrame>
	<PageContentTemplate>
		{#snippet analytics()}
			<div class="flex flex-col gap-2">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<Text size="xs" tone="muted">
						Scroll sideways for the rest of the cards — the order becomes a setting later.
					</Text>
					<!-- The arrows are the keyboard (and non-touch) way through the rail, so they stay
					     visible at every width. -->
					<div class="flex gap-1">
						<IconButton
							icon="heroicons:chevron-left"
							ariaLabel="Scroll cards left"
							size="sm"
							variant="outline"
							onclick={() => nudge(-1)}
						/>
						<IconButton
							icon="heroicons:chevron-right"
							ariaLabel="Scroll cards right"
							size="sm"
							variant="outline"
							onclick={() => nudge(1)}
						/>
					</div>
				</div>

				<div
					bind:this={rail}
					role="group"
					aria-label="Cashflow and wealth-goal cards"
					class="flex snap-x snap-proximity gap-4 overflow-x-auto pb-2"
				>
					<div data-card="trend" class={wideCardClass}>
						<AnalyticsCard title="Net trend"><NetTrendChart /></AnalyticsCard>
					</div>
					<div data-card="incoming" class={cardClass}>
						<AnalyticsCard title="Incoming"><TagDonut direction="in" /></AnalyticsCard>
					</div>
					<div data-card="outgoing" class={cardClass}>
						<AnalyticsCard title="Outgoing"><TagDonut direction="out" /></AnalyticsCard>
					</div>

					{#if goalSet}
						<div data-card="month" class={wideCardClass}>
							<AnalyticsCard title="July 2026 against your goal">
								<RunningMonthPanel month={running} stacked />
							</AnalyticsCard>
						</div>
						<div data-card="streak" class={cardClass}>
							<AnalyticsCard title="Streak"><StreakPanel /></AnalyticsCard>
						</div>
						<div data-card="standing" class={wideCardClass}>
							<AnalyticsCard title="Monthly standing">
								<div class="max-h-44 overflow-y-auto pr-1">
									{#each past as month (month.month)}
										<MonthStandingRow {month} />
									{/each}
								</div>
							</AnalyticsCard>
						</div>
						<div data-card="goal" class={cardClass}>
							<AnalyticsCard title="Monthly goal">
								<GoalSettingPanel
									percent={goalPercent}
									amount={goalAmount(running)}
									onAdjust={() => (goalDialog = true)}
								/>
							</AnalyticsCard>
						</div>
					{:else}
						<div data-card="goal" class={wideCardClass}>
							<AnalyticsCard title="Wealth goal">
								<NoGoalPanel onSet={() => (goalDialog = true)} />
							</AnalyticsCard>
						</div>
					{/if}
				</div>
			</div>
		{/snippet}

		<TransactionsPanel bind:selectedIds />
	</PageContentTemplate>
</PortalFrame>

<SetGoalDialog bind:open={goalDialog} bind:percent={goalPercent} />
