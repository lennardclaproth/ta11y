<script lang="ts">
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import { accountStore } from '$lib/stores/account.svelte';

	let { children } = $props();

	// Admin screens are simply absent for non-admin accounts: the navigation omits them
	// and a direct link lands here. This is a convenience, not the control -- the API
	// registers these routes admin-only and answers 403 regardless of what the client shows.
	$effect(() => {
		void accountStore.ensureLoaded();
	});
</script>

{#if !accountStore.loaded}
	<!-- Nothing is rendered until the account resolves, so the page never flashes a
	     "not available" state at an admin who simply has a slow connection. -->
	<AppShellTemplate>
		{#snippet top()}
			<TopNavbar title="Admin" />
		{/snippet}
		<section class="px-6 py-10" aria-busy="true"></section>
	</AppShellTemplate>
{:else if accountStore.isAdmin}
	{@render children()}
{:else}
	<AppShellTemplate>
		{#snippet top()}
			<TopNavbar title="Admin" />
		{/snippet}
		<section class="mx-auto w-full max-w-2xl px-6 py-10">
			<h2 class="text-2xl">This page isn't available</h2>
			<p class="mt-3 max-w-prose text-sm leading-relaxed text-slate-600">
				{#if accountStore.failed}
					Your account couldn't be loaded, so admin pages are hidden. Check your connection and
					reload.
				{:else}
					Listings, dailies and provider credentials are only available to admin accounts.
				{/if}
			</p>
		</section>
	</AppShellTemplate>
{/if}
