<script lang="ts">
	import { onMount } from 'svelte';
	import { adminMode } from '$lib/stores/admin.svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	let { children } = $props();
	let ready = $state(false);
	onMount(() => {
		ready = true;
	});

	// Keep admin tools gated while explaining how to enter, including on direct links.
</script>

{#if adminMode.enabled}
	{@render children()}
{:else}
	<AppShellTemplate>
		{#snippet top()}
			<TopNavbar
				title="Admin tools"
				accountName="Account"
				adminMode={false}
				onAdminToggle={(v) => adminMode.set(v)}
			/>
		{/snippet}
		<section class="mx-auto w-full max-w-2xl px-6 py-10">
			<h2 class="text-3xl">Enable admin mode to continue</h2>
			<p class="mt-3 mb-6 max-w-prose text-sm leading-relaxed text-slate-600">
				Listings and daily price uploads are managed in admin mode. Enable it here to open the page
				you selected. You can turn it off again in the account menu.
			</p>
			<Button disabled={!ready} onclick={() => adminMode.set(true)}>Enable admin mode</Button>
		</section>
	</AppShellTemplate>
{/if}
