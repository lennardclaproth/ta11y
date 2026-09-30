<script lang="ts">
	// Proposed molecule `activity-trail` — the import reports step by step instead of
	// filling a progress bar. One component, two orientations, so the dialog can stay
	// horizontal on desktop and the same markup can run vertical.
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { TrailDirection, TrailStep } from './activity-trail.types';
	import { trailDetailClasses, trailIcons, trailMarkerClasses } from './activity-trail.variants';

	type Props = {
		steps: TrailStep[];
		direction?: TrailDirection;
		/** Names the list for assistive technology; the region announcing it lives in the caller. */
		ariaLabel?: string;
		class?: string;
	};

	let {
		steps,
		direction = 'horizontal',
		ariaLabel = 'Import activity',
		class: className = ''
	}: Props = $props();

	const listClasses = $derived(
		[
			direction === 'horizontal' ? 'grid auto-cols-fr grid-flow-col' : 'flex flex-col',
			className
		]
			.filter(Boolean)
			.join(' ')
	);

	// Each item's connector reaches to the centre of the next marker; the marker itself is
	// opaque and sits above the line, so the rule reads as one continuous stroke.
	const itemClasses = $derived(
		direction === 'horizontal'
			? 'relative flex min-w-0 flex-col items-center px-1 text-center'
			: 'relative flex min-w-0 gap-3 pb-4 last:pb-0'
	);

	const connectorClasses = $derived(
		direction === 'horizontal'
			? 'absolute top-3 left-1/2 h-px w-full bg-slate-300'
			: 'absolute top-7 bottom-0 left-3 w-px -translate-x-1/2 bg-slate-300'
	);

	const textClasses = $derived(direction === 'horizontal' ? 'text-xs' : 'text-sm');
</script>

<ol class={listClasses} aria-label={ariaLabel}>
	{#each steps as step, index (step.key)}
		<li class={itemClasses}>
			{#if index < steps.length - 1}
				<span class={connectorClasses} aria-hidden="true"></span>
			{/if}

			<span
				class={[
					'relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full border',
					trailMarkerClasses[step.state]
				].join(' ')}
			>
				<Icon
					icon={trailIcons[step.state]}
					size="sm"
					class={step.state === 'running' ? 'animate-spin motion-reduce:animate-none' : ''}
				/>
			</span>

			<div class={['min-w-0', direction === 'horizontal' ? 'mt-1.5 w-full' : 'flex-1'].join(' ')}>
				<span class={[textClasses, 'block leading-tight font-medium text-slate-900'].join(' ')}>
					{step.title}
				</span>
				{#if step.detail}
					<span
						class={[
							textClasses,
							'block leading-tight tabular-nums',
							trailDetailClasses[step.state]
						].join(' ')}
					>
						{step.detail}
					</span>
				{/if}
			</div>
		</li>
	{/each}
</ol>
