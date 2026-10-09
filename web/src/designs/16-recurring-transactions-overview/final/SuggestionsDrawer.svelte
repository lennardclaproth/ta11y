<script lang="ts">
	// "Voorstellen uit je historie". Nothing here counts until it is confirmed, so the drawer sits
	// beside the list it would add to rather than replacing it.
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { recurringSuggestions, rhythmLabels } from '../recurring.fixture';

	let { open = $bindable(false) }: { open?: boolean } = $props();
</script>

<Drawer bind:open title="Suggestions from your history" width="max-w-lg">
	<Text size="sm" tone="muted" class="mb-4">
		{recurringSuggestions.length} patterns found in transactions you have not ignored. Nothing is added
		until you confirm it, and a dismissed suggestion does not come back.
	</Text>

	<div class="flex flex-col gap-3">
		{#each recurringSuggestions as suggestion (suggestion.id)}
			<Panel
				variant="muted"
				shape="sm"
				padding="md"
				bordered
				class="flex flex-col gap-3 border-dashed"
			>
				<div class="flex items-start justify-between gap-3">
					<div class="min-w-0">
						<Heading level="h3" size="sm" class="truncate text-slate-900">{suggestion.name}</Heading>
						<Text as="span" size="xs" tone="muted">
							{suggestion.direction === 'in' ? 'Income' : 'Expense'} · {suggestion.matches} transactions
							since {formatDisplayDate(suggestion.since)}
						</Text>
					</div>
					<Badge intent="neutral" variant="outline" size="sm">
						{rhythmLabels[suggestion.rhythm]}
					</Badge>
				</div>

				<div class="flex items-end justify-between gap-3">
					<Money amount={suggestion.amount} currency="EUR" size="lg" weight="semibold" />
					<!-- The name is a guess read off the statement text; showing the text is what makes the
					     guess checkable instead of something ta11y decided quietly. -->
					<span class="min-w-0 truncate text-xs text-slate-500">
						Named after “{suggestion.sample}”
					</span>
				</div>

				<div class="flex flex-wrap items-center gap-2 border-t border-slate-200 pt-3">
					<Button size="sm" shape="default">Confirm</Button>
					<Button size="sm" shape="default" variant="outline">Adjust name or rhythm</Button>
					<Button size="sm" shape="default" variant="ghost" intent="secondary">Dismiss</Button>
				</div>
			</Panel>
		{/each}
	</div>

	{#snippet footer()}
		<Button variant="outline" shape="default" onclick={() => (open = false)}>Close</Button>
	{/snippet}
</Drawer>
