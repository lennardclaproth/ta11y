<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import LedgerToolbar from './LedgerToolbar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/LedgerToolbar',
		component: LedgerToolbar,
		tags: ['autodocs'],
		argTypes: {
			showSearch: { control: 'boolean' }
		}
	});
</script>

<!--
  Everything that acts on the rows sits here: filters and actions on the rule, searching on the
  line underneath. Supporting actions are ruled so the one filled button stays the dominant one.
-->

<Story name="Transactions" asChild>
	<div class="bg-taupe-50">
		<LedgerToolbar title="Transactions" meta="128 rows" showSearch searchPlaceholder="Search description…">
			{#snippet actions()}
				<Button variant="ruled"><Icon icon="heroicons:cloud-arrow-up" />Import CSV</Button>
				<Button shape="default"><Icon icon="heroicons:plus" />Add transaction</Button>
			{/snippet}
		</LedgerToolbar>
	</div>
</Story>

<Story name="With a selection" asChild>
	<div class="bg-taupe-50">
		<LedgerToolbar title="Transactions" meta="128 rows · 3 selected" showSearch>
			{#snippet actions()}
				<Button variant="ruled"><Icon icon="heroicons:tag" />Tag 3 selected</Button>
				<Button shape="default"><Icon icon="heroicons:plus" />Add transaction</Button>
			{/snippet}
		</LedgerToolbar>
	</div>
</Story>

<Story name="With tabs and a row filter" asChild>
	<div class="bg-taupe-50">
		<LedgerToolbar
			title="Positions"
			meta="12 positions"
			showSearch
			searchPlaceholder="Search symbol or name…"
		>
			{#snippet before()}
				<Tabs
					tabs={[
						{ value: 'positions', label: 'Positions' },
						{ value: 'transactions', label: 'Transactions' }
					]}
					size="sm"
					ariaLabel="Portfolio view"
				/>
			{/snippet}
			{#snippet filters()}
				<Tabs
					tabs={[
						{ value: 'open', label: 'Open' },
						{ value: 'closed', label: 'Closed' },
						{ value: 'all', label: 'All' }
					]}
					size="sm"
					ariaLabel="Position status"
				/>
			{/snippet}
			{#snippet actions()}
				<Button shape="default"><Icon icon="heroicons:plus" />Add transaction</Button>
			{/snippet}
		</LedgerToolbar>
	</div>
</Story>

<Story name="Search in use" asChild>
	<div class="bg-taupe-50">
		<LedgerToolbar title="Transactions" meta="0 of 128 rows" showSearch searchValue="grocerys">
			{#snippet actions()}
				<Button shape="default"><Icon icon="heroicons:plus" />Add transaction</Button>
			{/snippet}
		</LedgerToolbar>
	</div>
</Story>

<!-- Admin screens have no charts, and one sentence of explanation under the title. -->
<Story name="Admin listings" asChild>
	<div class="bg-taupe-50">
		<LedgerToolbar
			title="Market listings"
			meta="42 listings"
			description="Manage instruments used in your portfolio and price history."
			showSearch
			searchPlaceholder="Search listings…"
		>
			{#snippet actions()}
				<Button variant="ruled"><Icon icon="heroicons:book-open" />Browse catalogue</Button>
				<Button shape="default"><Icon icon="heroicons:plus" />Add listing</Button>
			{/snippet}
		</LedgerToolbar>
	</div>
</Story>
