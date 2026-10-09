<script lang="ts">
	import Button from '$lib/components/atoms/button/Button.svelte';
	import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';
	import DemoFrame from '../shared/DemoFrame.svelte';
	import MockDialog from '../shared/MockDialog.svelte';
	import MockTransactionFields from '../shared/MockTransactionFields.svelte';
	import SpecimenSection from '../shared/SpecimenSection.svelte';
	import RuledNotice from './RuledNotice.svelte';

	const intents: AlertIntent[] = ['success', 'error', 'warning', 'info'];
	const intentMessages: Record<AlertIntent, string> = {
		success: 'Transaction created',
		error: 'Failed to create transaction',
		warning: 'Some rows were skipped during the import',
		info: 'Filtered to 2026-01-01 – 2026-03-31'
	};
</script>

<div class="flex flex-col gap-8 bg-taupe-100 p-4 lg:p-8">
	<SpecimenSection label="In context" hint="Toast after an action · notice that stays">
		<DemoFrame>
			{#snippet notice()}
				<RuledNotice intent="error">
					Failed to load transactions
					{#snippet action()}
						<Button size="sm" variant="outline" intent="secondary">Try again</Button>
					{/snippet}
				</RuledNotice>
			{/snippet}

			{#snippet overlay()}
				<div
					class="absolute inset-x-4 top-4 flex flex-col gap-2 sm:left-auto sm:w-full sm:max-w-sm"
				>
					<RuledNotice intent="success" surface="paper" dismissible>
						Transaction created
					</RuledNotice>
					<RuledNotice intent="info" surface="paper" dismissible>
						Filtered to 2026-01-01 – 2026-03-31
					</RuledNotice>
				</div>
			{/snippet}
		</DemoFrame>
	</SpecimenSection>

	<div class="grid gap-8 lg:grid-cols-2">
		<SpecimenSection label="In a modal" hint="Save failed · field error">
			<MockDialog>
				{#snippet bottom()}
					<RuledNotice intent="error" title="This transaction was not saved">
						A transaction with the same amount and description already exists on Mar 2, 2026. Pick
						another day.
					</RuledNotice>
				{/snippet}
				<MockTransactionFields />
			</MockDialog>
		</SpecimenSection>

		<SpecimenSection label="Intents" hint="Rule and kicker carry the meaning">
			<div class="flex flex-col gap-3 bg-taupe-50 p-4">
				{#each intents as intent (intent)}
					<RuledNotice {intent}>{intentMessages[intent]}</RuledNotice>
				{/each}
			</div>
		</SpecimenSection>
	</div>
</div>
