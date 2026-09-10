<script lang="ts">
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { InputShape, InputSize } from '$lib/components/atoms/input/input.types';
	import {
		inputIconContainerSizeClasses,
		inputIconPaddingClasses,
		inputIconSizeClasses
	} from '$lib/components/molecules/icon-input/icon-input.variants';

	type Props = {
		/** Two-way bindable query value. */
		value?: string;
		placeholder?: string;
		size?: InputSize;
		shape?: InputShape;
		disabled?: boolean;
		/** Debounce window (ms) before `onSearch` fires. */
		debounceMs?: number;
		ariaLabel?: string;
		/** Debounced query callback (fires after `debounceMs` of inactivity; immediate on clear). */
		onSearch?: (value: string) => void;
		/** Raw, non-debounced input callback. */
		oninput?: (event: Event) => void;
		class?: string;
	};

	let {
		value = $bindable(''),
		placeholder = 'Search…',
		size = 'md',
		shape = 'rounded',
		disabled = false,
		debounceMs = 300,
		ariaLabel = 'Search',
		onSearch,
		oninput,
		class: className = ''
	}: Props = $props();

	let mounted = false;

	// Debounce `onSearch` on value changes; skip the initial mount so it doesn't fire for the seed value.
	$effect(() => {
		const current = value;
		if (!mounted) {
			mounted = true;
			return;
		}
		if (!onSearch) return;
		const timer = setTimeout(() => onSearch(current), debounceMs);
		return () => clearTimeout(timer);
	});

	function clear() {
		value = '';
		onSearch?.('');
	}

	const clearable = $derived(value.length > 0 && !disabled);

	// Only the trailing clear button reserves space; the placeholder and the field's
	// own affordances already say this is a search box, so a leading icon would be a
	// second thing to look at for no extra meaning.
	//
	// `type="search"` makes Chromium paint its own clear button inside the field, which lands
	// beside the one below and reads as two X's. The type is worth keeping for the semantics,
	// so the native affordance is suppressed and this component owns clearing outright.
	const inputClasses = $derived(
		[
			'[&::-webkit-search-cancel-button]:appearance-none',
			clearable ? inputIconPaddingClasses[size].right : '',
			className
		]
			.filter(Boolean)
			.join(' ')
	);
</script>

<div class="group relative w-full">
	<Input
		type="search"
		bind:value
		{placeholder}
		{size}
		{shape}
		{disabled}
		{ariaLabel}
		class={inputClasses}
		{oninput}
	/>

	{#if clearable}
		<button
			type="button"
			aria-label="Clear search"
			class={[
				'absolute inset-y-0 right-0 flex items-center justify-center text-slate-500',
				'transition-colors hover:text-slate-700',
				'focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none',
				inputIconContainerSizeClasses[size]
			].join(' ')}
			onclick={clear}
		>
			<Icon icon="heroicons:x-mark" size={inputIconSizeClasses[size]} />
		</button>
	{/if}
</div>
