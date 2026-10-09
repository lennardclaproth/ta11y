<script lang="ts">
	/**
	 * Proposed field-level error line: the existing FormField error text plus the same icon the
	 * notices use, so a per-field error reads as part of one family. Shared by all variants.
	 */
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	type Props = {
		id?: string;
		message: string;
	};

	let { id, message }: Props = $props();
</script>

<div {id} role="alert" class="flex items-center gap-1.5 text-red-700">
	<Icon icon="heroicons:exclamation-circle" size="sm" />
	<Text as="span" size="sm" tone="danger">{message}</Text>
</div>
