<script lang="ts">
	/**
	 * Static stand-in for the Dialog molecule (which uses the native top layer and would cover the
	 * whole screenshot). Same header / body / footer rhythm, so the notice can be judged in place.
	 */
	import type { Snippet } from 'svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	type Props = {
		title?: string;
		/** Notice directly under the header, above the fields. */
		top?: Snippet;
		/** Notice directly above the footer, next to the failing action. */
		bottom?: Snippet;
		/** Save in flight: the action is busy and no notice has been produced yet. */
		saving?: boolean;
		children: Snippet;
		class?: string;
	};

	let {
		title = 'New transaction',
		top,
		bottom,
		saving = false,
		children,
		class: className = ''
	}: Props = $props();
</script>

<div class={['bg-slate-900/40 p-6', className].filter(Boolean).join(' ')}>
	<div class="mx-auto w-full max-w-lg rounded-2xl bg-white text-slate-800 shadow-2xl">
		<header class="flex items-start justify-between gap-4 px-5 pt-5">
			<h2 class="font-heading text-2xl leading-tight tracking-tight text-slate-900">{title}</h2>
			<span class="inline-flex size-8 items-center justify-center rounded-full text-slate-500">
				<Icon icon="heroicons:x-mark" size="md" />
			</span>
		</header>

		{#if top}
			<div class="px-5 pt-4">{@render top()}</div>
		{/if}

		<div class="px-5 py-4 text-sm">
			{@render children()}
		</div>

		{#if bottom}
			<div class="px-5 pb-4">{@render bottom()}</div>
		{/if}

		<footer class="flex items-center justify-end gap-2 border-t border-slate-200 px-5 py-4">
			<Button variant="ghost" intent="secondary" disabled={saving}>Cancel</Button>
			<Button intent="success" loading={saving}>Save</Button>
		</footer>
	</div>
</div>
