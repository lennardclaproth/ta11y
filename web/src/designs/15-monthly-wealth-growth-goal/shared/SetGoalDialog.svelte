<script lang="ts">
	// Setting or adjusting the one monthly goal. Percentage entry with a slider next to it, and the
	// same number in euros so the choice is readable in money as well as in percent.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Slider from '$lib/components/atoms/slider/Slider.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';

	type Props = {
		open?: boolean;
		percent?: number;
		/** Typical income used for the worked example under the field. */
		exampleIncome?: number;
		title?: string;
	};

	let {
		open = $bindable(false),
		percent = $bindable(30),
		exampleIncome = 5000,
		title = 'Monthly goal'
	}: Props = $props();

	const value = $derived(String(percent));
	const example = $derived(Math.round((exampleIncome * percent) / 100));

	function setPercent(next: number) {
		percent = Math.min(100, Math.max(0, Math.round(next)));
	}
</script>

<Dialog bind:open {title} size="sm">
	<div class="flex flex-col gap-4">
		<Text size="sm" tone="muted">
			Which part of the income you mark each month should go towards wealth: savings, a broker
			deposit or an asset purchase?
		</Text>

		<FormField
			label="Share of income"
			id="goal-share"
			hint="A whole percentage between 0 and 100."
		>
			{#snippet children(field)}
				<div class="flex items-center gap-3">
					<Input
						id={field.id}
						type="number"
						{value}
						ariaDescribedby={field.describedby}
						class="w-24"
						oninput={(e) => setPercent(Number((e.currentTarget as HTMLInputElement).value))}
					/>
					<span class="text-sm whitespace-nowrap text-slate-700">% of income</span>
				</div>
			{/snippet}
		</FormField>

		<Slider
			value={percent}
			min={0}
			max={100}
			step={5}
			ariaLabel="Share of income towards wealth"
			oninput={(e) => setPercent(Number((e.currentTarget as HTMLInputElement).value))}
		/>

		<Text size="sm" tone="muted">
			On a month with <Money amount={exampleIncome} currency="EUR" size="sm" /> of marked income
			that is <Money amount={example} currency="EUR" size="sm" weight="semibold" />.
		</Text>

		<Text size="sm" tone="muted">
			Applies from July 2026 onwards. Months before that keep the goal they had.
		</Text>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Cancel</Button>
		<Button onclick={() => (open = false)}>Save goal</Button>
	{/snippet}
</Dialog>
