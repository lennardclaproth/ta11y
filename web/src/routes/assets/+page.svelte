<script lang="ts">
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import AssetClassDrawer from '$lib/components/organisms/asset-class-drawer/AssetClassDrawer.svelte';
	import AddAssetItemDialog from '$lib/components/organisms/add-asset-item-dialog/AddAssetItemDialog.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import {
		listAssetClasses,
		getAssetSnapshots,
		getAssetClassDetails,
		createAssetClass
	} from '$lib/services/assets';
	import { goto } from '$app/navigation';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore } from '$lib/stores/period.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { decimalStringToNumber } from '$lib/api/money';
	import { donutRamps } from '$lib/charts/theme';
	import type { AssetClass, AssetClassDetails, AssetSnapshotPoint } from '$lib/api/types';

	let classes = $state<AssetClass[]>([]);
	let snapshots = $state<AssetSnapshotPoint[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let drawerOpen = $state(false);
	let details = $state<AssetClassDetails | null>(null);
	let detailsLoading = $state(false);

	// One period for the whole app, chosen in the account overview.
	const from = $derived(periodStore.from);
	const to = $derived(periodStore.to);

	let createOpen = $state(false);
	let className = $state('');
	let creatingClass = $state(false);

	let addItemOpen = $state(false);

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;
	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	const distribution = $derived(
		classes
			.filter((c) => !c.archived)
			.map((c) => ({ label: c.name, value: decimalStringToNumber(c.current_worth) }))
	);

	async function loadAll() {
		loading = true;
		error = null;
		try {
			await accountStore.ensureLoaded();
			// No account yet is an empty state, not a failure.
			if (!accountStore.hasAccount) {
				classes = [];
				snapshots = [];
				return;
			}
			const [cls, snaps] = await Promise.all([
				listAssetClasses({}),
				getAssetSnapshots({
					from: from || undefined,
					to: to || undefined
				})
			]);
			classes = cls;
			snapshots = snaps;
		} catch {
			error = 'Failed to load assets';
		} finally {
			loading = false;
		}
	}

	// Reload (and zoom the worth chart) whenever the date range changes; also the initial load.
	$effect(() => {
		void from;
		void to;
		void loadAll();
	});

	async function createClass() {
		if (className.trim() === '') return;
		creatingClass = true;
		try {
			await accountStore.ensureLoaded();
			await createAssetClass({ name: className.trim() });
			createOpen = false;
			className = '';
			toast.success('Asset class created');
			void loadAll();
		} catch {
			toast.error('Failed to create asset class');
		} finally {
			creatingClass = false;
		}
	}

	async function openClass(row: AssetClass) {
		drawerOpen = true;
		detailsLoading = true;
		details = null;
		try {
			await accountStore.ensureLoaded();
			details = await getAssetClassDetails(row.id);
		} finally {
			detailsLoading = false;
		}
	}

	// A new item changes the class total, so the table behind the drawer is reloaded too.
	async function reloadOpenClass() {
		if (!details) return;
		const classId = details.class.id;
		detailsLoading = true;
		try {
			details = await getAssetClassDetails(classId);
		} catch {
			// The item was saved; only the refresh failed, so say that rather than let the
			// drawer silently keep showing the figures from before it.
			toast.error('Saved, but the class could not be refreshed');
		} finally {
			detailsLoading = false;
		}
		void loadAll();
	}
</script>

{#snippet worthCell(row: AssetClass)}
	<Money amount={decimalStringToNumber(row.current_worth)} currency="EUR" size="sm" />
{/snippet}

{#snippet growthCell(row: AssetClass)}
	{#if row.growth_pct !== null && row.growth_pct !== undefined}
		<Badge intent={row.growth_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.growth_pct >= 0 ? '+' : ''}{row.growth_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<PageContentTemplate title="Assets">
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
				<AnalyticsCard title="Total worth" class="lg:col-span-2">
					<TimeSeriesChart
						height="h-52"
						{loading}
						labels={snapshots.map((s) => s.date)}
						xTickFormat={monthShort}
						datasets={[
							{
								label: 'Total worth',
								data: snapshots.map((s) => decimalStringToNumber(s.total_worth)),
								color: '#059669',
								fill: true
							}
						]}
					/>
				</AnalyticsCard>
				<AnalyticsCard title="Distribution">
					<DonutChart
						data={distribution}
						ramp={donutRamps.incoming}
						{loading}
						formatValue={euro}
						centerLabel="Total"
					/>
				</AnalyticsCard>
			</div>
		{/snippet}

		<!-- The asset-class table has never had a search, and this change does not give it one. -->
		<LedgerToolbar
			title="Asset classes"
			meta={loading ? 'Loading…' : error ? 'Could not load' : `${classes.length} classes`}
		>
			{#snippet actions()}
				<Button shape="default" onclick={() => (createOpen = true)}>
					<Icon icon="heroicons:plus" />
					Add asset class
				</Button>
			{/snippet}
		</LedgerToolbar>
		<DataTable
			rows={classes}
			{loading}
			{error}
			emptyText="No asset classes yet. Add one to start tracking what you own."
			onRowClick={openClass}
			columns={[
				{ key: 'name', header: 'Class', value: (r: AssetClass) => r.name },
				{ key: 'source', header: 'Source', value: (r: AssetClass) => r.source },
				{ key: 'worth', header: 'Current worth', align: 'right', cell: worthCell },
				{ key: 'growth', header: 'Growth', align: 'right', cell: growthCell }
			]}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<AssetClassDrawer
	bind:open={drawerOpen}
	{details}
	loading={detailsLoading}
	onOpenClass={() => details && goto(`/assets/${details.class.id}`)}
	onAddItem={() => (addItemOpen = true)}
	onOpenItem={(assetId) => details && goto(`/assets/${details.class.id}?item=${assetId}`)}
/>

{#if details}
	<AddAssetItemDialog
		bind:open={addItemOpen}
		classId={details.class.id}
		className={details.class.name}
		onSaved={reloadOpenClass}
	/>
{/if}

<Dialog bind:open={createOpen} title="New asset class" size="sm">
	<FormField label="Name" id="class-name">
		{#snippet children(ctx)}
			<Input id={ctx.id} bind:value={className} placeholder="e.g. Real estate" />
		{/snippet}
	</FormField>
	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (createOpen = false)}>Cancel</Button>
		<Button intent="success" onclick={createClass} loading={creatingClass}>Create</Button>
	{/snippet}
</Dialog>
