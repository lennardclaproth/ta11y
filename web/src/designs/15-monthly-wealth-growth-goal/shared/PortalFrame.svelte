<script lang="ts">
	// The portal shell every variant sits in: masthead plus the main region. Nothing new here,
	// it only keeps the three prototypes from repeating the same wiring.
	import type { Snippet } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	type Props = {
		title?: string;
		children: Snippet;
	};

	let { title = 'Cashflow', children }: Props = $props();

	const navActions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];
</script>

<!-- The shell is viewport-tall in the app; in the prototype it grows with its content so a
     screenshot shows the whole screen instead of the first fold. -->
<AppShellTemplate class="h-auto! min-h-dvh">
	{#snippet top()}
		<TopNavbar
			{title}
			showSearch
			showDateRange
			actions={navActions}
			accountName="Design preview"
			accountEmail="preview@example.com"
		/>
	{/snippet}

	{@render children()}
</AppShellTemplate>
