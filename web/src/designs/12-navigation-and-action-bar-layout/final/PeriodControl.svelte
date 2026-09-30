<script lang="ts">
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { period } from '../shared/mock-data';

	/**
	 * The one period of the app: presets stay visible, and the resolved range stays visible
	 * next to them. Clicking the range opens the existing date picker.
	 *
	 * Build note: picking "Custom" should open that same picker, which needs `open` to become
	 * a bindable prop on DateRangePicker. Presets therefore stay a selection here.
	 */
	type Props = {
		/** `block` stacks presets above the range for the narrow overview panel. */
		layout?: 'inline' | 'block';
		showCaption?: boolean;
	};

	let { layout = 'inline', showCaption = false }: Props = $props();

	let preset = $state('6m');
	let from = $state<string | null>(period.from);
	let to = $state<string | null>(period.to);

	const presets = [
		{ value: '1m', label: '1M' },
		{ value: '3m', label: '3M' },
		{ value: '6m', label: '6M' },
		{ value: 'ytd', label: 'YTD' },
		{ value: 'custom', label: 'Custom' }
	];
</script>

<div
	class={layout === 'inline'
		? 'flex flex-col items-end gap-1.5'
		: 'flex flex-col items-stretch gap-2'}
>
	<Tabs
		tabs={presets}
		bind:value={preset}
		size="sm"
		ariaLabel="Period"
		class={layout === 'block' ? 'w-full justify-between' : ''}
	/>
	<DateRangePicker
		bind:from
		bind:to
		size="sm"
		showPresets={false}
		ariaLabel="Selected period, opens the date picker"
		class={layout === 'block' ? 'w-full justify-center' : ''}
	/>
	{#if showCaption}
		<Text as="p" size="xs" tone="subtle">
			Applies to the charts and the table on every page.
		</Text>
	{/if}
</div>
