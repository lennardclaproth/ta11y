<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import LedgerToolbar from './LedgerToolbar.svelte';
	const { Story } = defineMeta({ title: 'Organisms/LedgerToolbar', component: LedgerToolbar });
</script>

<Story
	name="Transactions"
	args={{ title: 'Transactions', actionLabel: 'Add transaction', onAdd: () => {} }}
/>
<Story
	name="Asset classes"
	args={{ title: 'Asset classes', actionLabel: 'Add asset class', onAdd: () => {} }}
/>
<!-- An action that is not a creation carries its own icon instead of the plus. -->
<Story
	name="Own icon"
	args={{
		title: 'Recurring items',
		actionLabel: 'Review suggestions (3)',
		actionIcon: 'heroicons:light-bulb',
		onAdd: () => {}
	}}
/>
<!-- Nothing to act on: the header keeps its title and reads as a heading alone. -->
<Story name="Without an action" args={{ title: 'Recurring items' }} />
