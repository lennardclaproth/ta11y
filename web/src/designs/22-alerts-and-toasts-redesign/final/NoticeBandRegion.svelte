<script lang="ts">
	/**
	 * Final design for #22 — proposed replacement for the ToastHost organism.
	 *
	 * The single notification place, but in normal document flow directly under the masthead instead
	 * of floating at the top right. Stacked bands share one hairline (`-mt-px`) so several notices
	 * read as one ruled block rather than a pile of cards. The store, its timers and its API are
	 * unchanged — only where the host renders and what it renders change.
	 */
	import NoticeBand from './NoticeBand.svelte';
	import type { NoticeItem } from './notices';

	type Props = {
		items: NoticeItem[];
		onDismiss?: (id: string) => void;
		class?: string;
	};

	let { items, onDismiss, class: className = '' }: Props = $props();
</script>

{#if items.length}
	<div
		class={['flex flex-col', className].filter(Boolean).join(' ')}
		role="region"
		aria-label="Notifications"
	>
		{#each items as item, index (item.id)}
			<NoticeBand
				intent={item.intent}
				title={item.title}
				gutter="page"
				dismissible={item.dismissible ?? true}
				onDismiss={() => onDismiss?.(item.id)}
				class={index > 0 ? '-mt-px' : ''}
			>
				{item.message}
			</NoticeBand>
		{/each}
	</div>
{/if}
