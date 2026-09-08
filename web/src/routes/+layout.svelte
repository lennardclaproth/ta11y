<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import favicon from '$lib/assets/favicon.svg';

	import Spinner from '$lib/components/atoms/spinner/Spinner.svelte';
	import { accountStore } from '$lib/stores/account.svelte';

	let { children } = $props();

	// The sign-in page is the one route that must render without a session.
	const isSignInRoute = $derived(page.url.pathname === '/login');

	$effect(() => {
		if (isSignInRoute) return;
		void accountStore.ensureLoaded().then(() => {
			// A signed-out visitor is sent to sign in. `accountStore` also redirects on a
			// 401 from any later request, which covers a session lapsing mid-visit.
			if (accountStore.isEmpty) {
				window.location.assign('/login');
			}
		});
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

{#if isSignInRoute}
	{@render children()}
{:else if accountStore.hasAccount}
	{@render children()}
{:else if accountStore.failed}
	<!-- An unreachable API is not the same as being signed out, so it says so rather than
	     bouncing the visitor to a sign-in page that will not load either. -->
	<main class="mx-auto w-full max-w-2xl px-6 py-16">
		<h1 class="text-2xl">Can't reach the server</h1>
		<p class="mt-3 text-sm leading-relaxed text-slate-600">
			ta11y couldn't load your session. Check your connection and reload the page.
		</p>
	</main>
{:else}
	<!-- Resolving the session, or redirecting to sign in. Nothing is rendered so the app
	     never flashes signed-in chrome at someone who is signed out. -->
	<main class="flex min-h-screen items-center justify-center" aria-busy="true">
		<Spinner size="lg" class="text-slate-400" />
	</main>
{/if}
