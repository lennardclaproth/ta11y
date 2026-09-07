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
