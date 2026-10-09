<script lang="ts">
	/**
	 * Final design for #22 — "Masthead band" (variant C). Every scenario is rendered in its real
	 * surroundings so the notice can be judged in place rather than on a swatch sheet.
	 */
	import Button from '$lib/components/atoms/button/Button.svelte';
	import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';
	import DemoFrame from '../shared/DemoFrame.svelte';
	import MockDialog from '../shared/MockDialog.svelte';
	import MockTransactionFields from '../shared/MockTransactionFields.svelte';
	import SpecimenSection from '../shared/SpecimenSection.svelte';
	import NoticeBand from './NoticeBand.svelte';
	import NoticeBandRegion from './NoticeBandRegion.svelte';
	import { singleNotice, stackedNotices } from './notices';

	type Props = {
		scenario?: 'default' | 'loading' | 'empty' | 'no-matches' | 'error';
	};

	let { scenario = 'default' }: Props = $props();

	const intents: AlertIntent[] = ['success', 'error', 'warning', 'info'];
	const intentMessages: Record<AlertIntent, string> = {
		success: 'Transaction created',
		error: 'Failed to create transaction',
		warning: 'Some rows were skipped during the import',
		info: 'Filtered to 2026-01-01 – 2026-03-31'
	};
</script>

<div class="flex flex-col gap-8 bg-taupe-100 p-4 lg:p-8">
	{#if scenario === 'default'}
		<SpecimenSection label="After an action" hint="One band under the masthead · dismissible">
			<DemoFrame>
				{#snippet banner()}
					<NoticeBandRegion items={singleNotice} />
				{/snippet}
			</DemoFrame>
		</SpecimenSection>

		<SpecimenSection label="Two at once" hint="Stacked bands share one rule">
			<DemoFrame>
				{#snippet banner()}
					<NoticeBandRegion items={stackedNotices} />
				{/snippet}
			</DemoFrame>
		</SpecimenSection>

		<div class="grid gap-8 lg:grid-cols-2">
			<SpecimenSection label="In a modal" hint="Save failed · field error · no toast">
				<MockDialog>
					{#snippet top()}
						<div class="-mx-5">
							<NoticeBand
								intent="error"
								surface="inset"
								gutter="dialog"
								title="This transaction was not saved"
							>
								A transaction with the same amount and description already exists on Mar 2, 2026.
							</NoticeBand>
						</div>
					{/snippet}
					<MockTransactionFields />
				</MockDialog>
			</SpecimenSection>

			<SpecimenSection label="Intents" hint="One shape, a solid chip per intent">
				<div class="flex flex-col border border-slate-300 bg-taupe-50 p-4">
					{#each intents as intent, index (intent)}
						<NoticeBand {intent} gutter="panel" class={index > 0 ? '-mt-px' : ''}>
							{intentMessages[intent]}
						</NoticeBand>
					{/each}
				</div>
			</SpecimenSection>
		</div>
	{:else if scenario === 'loading'}
		<SpecimenSection label="Loading" hint="No band while the records load">
			<DemoFrame state="loading" />
		</SpecimenSection>

		<SpecimenSection label="Saving" hint="Busy action, notice only after the result">
			<MockDialog saving>
				<MockTransactionFields invalid={false} />
			</MockDialog>
		</SpecimenSection>
	{:else if scenario === 'empty'}
		<SpecimenSection label="Empty" hint="Nothing to report, so no band">
			<DemoFrame state="empty" />
		</SpecimenSection>
	{:else if scenario === 'no-matches'}
		<SpecimenSection label="No matches" hint="The active filter is stated as an info band">
			<DemoFrame state="no-matches">
				{#snippet notice()}
					<NoticeBand intent="info" gutter="panel" dismissible>
						Filtered to 2026-01-01 – 2026-03-31
					</NoticeBand>
				{/snippet}
			</DemoFrame>
		</SpecimenSection>
	{:else}
		<SpecimenSection label="Error that stays" hint="At the content, with a recovery action">
			<DemoFrame state="failed">
				{#snippet notice()}
					<NoticeBand intent="error" gutter="panel">
						Failed to load transactions
						{#snippet action()}
							<Button size="sm" variant="outline" intent="secondary">Try again</Button>
						{/snippet}
					</NoticeBand>
				{/snippet}
			</DemoFrame>
		</SpecimenSection>

		<SpecimenSection label="Error in a modal" hint="Stays in the modal, input preserved">
			<MockDialog>
				{#snippet top()}
					<div class="-mx-5">
						<NoticeBand
							intent="error"
							surface="inset"
							gutter="dialog"
							title="This transaction was not saved"
						>
							A transaction with the same amount and description already exists on Mar 2, 2026.
						</NoticeBand>
					</div>
				{/snippet}
				<MockTransactionFields />
			</MockDialog>
		</SpecimenSection>
	{/if}
</div>
