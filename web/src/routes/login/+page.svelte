<script lang="ts">
  import { page } from '$app/state';

  import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
  import Heading from '$lib/components/atoms/typography/Heading.svelte';
  import Panel from '$lib/components/atoms/panel/Panel.svelte';
  import Spinner from '$lib/components/atoms/spinner/Spinner.svelte';
  import Text from '$lib/components/atoms/typography/Text.svelte';
  import { listProviders, loginUrl, providerLabel, signInErrorMessage } from '$lib/services/auth';

  // Sign-in is a full page navigation to the API, which redirects on to the identity
  // provider and back. It is deliberately an <a>, not a fetch: XHR cannot follow the
  // provider's redirect, and the session cookie is set on the way back.
  let providers = $state<string[] | null>(null);
  let loadFailed = $state(false);

  // The callback redirects here with ?error=<reason> when a sign-in could not complete.
  const signInError = $derived(signInErrorMessage(page.url.searchParams.get('error')));

  $effect(() => {
    listProviders()
      .then((resolved) => {
        providers = resolved;
      })
      .catch(() => {
        loadFailed = true;
      });
  });
</script>

<svelte:head><title>Sign in · ta11y</title></svelte:head>

<main class="flex min-h-screen items-center justify-center px-6 py-16">
  <Panel class="w-full max-w-md" padding="lg">
    <Heading level="h1" size="xl">Sign in</Heading>
    <Text class="mt-2 block text-sm leading-relaxed text-slate-600">
      ta11y uses your existing account with an identity provider. It never sees or
      stores a password.
    </Text>

    {#if signInError}
      <!-- The card already pads its own content, so the band keeps its text on the card's own
           text column rather than adding a gutter of its own. -->
      <NoticeBand intent="error" surface="inset" gutter="none" class="mt-6" title="Couldn't sign you in">
        {signInError}
      </NoticeBand>
    {/if}

    <div class="mt-8 flex flex-col gap-3">
      {#if providers === null && !loadFailed}
        <div class="flex items-center gap-2 text-sm text-slate-600">
          <Spinner size="sm" />
          Loading sign-in options…
        </div>
      {:else if loadFailed}
        <NoticeBand intent="error" surface="inset" gutter="none" title="The service is unreachable">
          Sign-in options couldn't be loaded. Check your connection and reload the page.
        </NoticeBand>
      {:else if providers && providers.length === 0}
        <NoticeBand
          intent="warning"
          surface="inset"
          gutter="none"
          title="No sign-in methods are configured"
        >
          The server has authentication switched on but no identity provider set up.
        </NoticeBand>
      {:else if providers}
        {#each providers as provider (provider)}
          <a
            href={loginUrl(provider)}
            data-sveltekit-reload
            class="inline-flex w-full items-center justify-center rounded-md border border-slate-300 bg-white px-4 py-2.5 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus-visible:ring-2 focus-visible:ring-amber-400 focus-visible:outline-none"
          >
            Continue with {providerLabel(provider)}
          </a>
        {/each}
      {/if}
    </div>
  </Panel>
</main>
