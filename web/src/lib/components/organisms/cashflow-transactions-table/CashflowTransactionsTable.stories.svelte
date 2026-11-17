<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import CashflowTransactionsTable from './CashflowTransactionsTable.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import { ignoreRules } from '$lib/data/fixtures/ignoreRules';

	// With ignored rows shown, each one names the rule that put it there — or says "By
	// hand" when nobody's rule did, rather than being the one row without a badge.
	const withIgnored = cashflowTransactions.slice(0, 10).map((tx, index) => {
		if (index === 1) return { ...tx, ignored: true, ignored_by_rule_id: ignoreRules[1].id };
		if (index === 3) return { ...tx, ignored: true, ignored_by_rule_id: ignoreRules[0].id };
		if (index === 5) return { ...tx, ignored: true, ignored_by_rule_id: null };
		return tx;
	});

	const tagOptions = [
		{ value: 'salary', label: 'salary' },
		{ value: 'rent', label: 'rent' },
		{ value: 'groceries', label: 'groceries' },
		{ value: 'utilities', label: 'utilities' },
		{ value: 'dining', label: 'dining' },
		{ value: 'transport', label: 'transport' }
	];

	const { Story } = defineMeta({
		title: 'Organisms/CashflowTransactionsTable',
		component: CashflowTransactionsTable,
		tags: ['autodocs']
	});
</script>

<Story name="Default" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable
			rows={cashflowTransactions.slice(0, 10)}
			total={cashflowTransactions.length}
			{tagOptions}
		/>
	</div>
</Story>

<Story name="Selected (with bulk actions)" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable
			rows={cashflowTransactions.slice(0, 10)}
			total={cashflowTransactions.length}
			selectedIds={['cf-0002', 'cf-0003']}
			{tagOptions}
		>
			{#snippet bulkActions()}
				<Button size="sm" variant="outline">Tag</Button>
				<Button size="sm" variant="outline" intent="error">Ignore</Button>
			{/snippet}
		</CashflowTransactionsTable>
	</div>
</Story>

<Story name="Showing ignored rows" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable
			rows={withIgnored}
			total={cashflowTransactions.length}
			showIgnored
			{ignoreRules}
			{tagOptions}
		/>
	</div>
</Story>

<!-- A rule that has since been deleted leaves its rows ignored; they read as "By hand"
     because there is no rule left to name. -->
<Story name="Ignored by a rule that is gone" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable
			rows={withIgnored}
			total={cashflowTransactions.length}
			showIgnored
			ignoreRules={[]}
			{tagOptions}
		/>
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable rows={[]} loading {tagOptions} />
	</div>
</Story>

<Story name="Empty" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable rows={[]} {tagOptions} />
	</div>
</Story>

<Story name="Error" asChild>
	<div class="h-[28rem] rounded-2xl border border-slate-200 bg-white">
		<CashflowTransactionsTable rows={[]} error="Failed to load transactions" {tagOptions} />
	</div>
</Story>
