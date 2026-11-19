<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ActivityTrail from './ActivityTrail.svelte';
	import { trailDirections } from './activity-trail.types';
	import type { TrailStep } from './activity-trail.types';

	const route: TrailStep[] = [
		{ key: 'file', title: 'File', state: 'pending' },
		{ key: 'cashflow', title: 'Cashflow', state: 'pending' },
		{ key: 'portfolio', title: 'Portfolio', state: 'pending' },
		{ key: 'performance', title: 'Performance', state: 'pending' }
	];

	const running: TrailStep[] = [
		{ key: 'file', title: 'File', detail: 'Accepted', state: 'done' },
		{ key: 'cashflow', title: 'Cashflow', detail: 'Imported', state: 'done' },
		{ key: 'portfolio', title: 'Portfolio', detail: 'Reading…', state: 'running' },
		{ key: 'performance', title: 'Performance', detail: 'Waiting', state: 'pending' }
	];

	const finished: TrailStep[] = [
		{ key: 'file', title: 'File', detail: 'Accepted', state: 'done' },
		{ key: 'cashflow', title: 'Cashflow', detail: 'Imported', state: 'done' },
		{ key: 'portfolio', title: 'Portfolio', detail: 'Partly imported', state: 'warning' },
		{ key: 'performance', title: 'Performance', detail: 'Recalculated', state: 'done' }
	];

	const refused: TrailStep[] = [
		{ key: 'file', title: 'File', detail: 'Not recognised', state: 'failed' },
		{ key: 'cashflow', title: 'Cashflow', detail: 'Not started', state: 'pending' },
		{ key: 'portfolio', title: 'Portfolio', detail: 'Not started', state: 'pending' },
		{ key: 'performance', title: 'Performance', detail: 'Not started', state: 'pending' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/ActivityTrail',
		component: ActivityTrail,
		tags: ['autodocs'],
		argTypes: {
			direction: { control: 'select', options: trailDirections }
		}
	});
</script>

<Story name="Playground" args={{ steps: running, direction: 'horizontal' }} />

<!-- Before anything runs the trail is only a route map: titles, no detail line to wrap. -->
<Story name="Not started" args={{ steps: route, direction: 'horizontal' }} />

<Story name="Running" args={{ steps: running, direction: 'horizontal' }} />

<Story name="Finished with a warning" args={{ steps: finished, direction: 'horizontal' }} />

<Story name="Failed at the first step" args={{ steps: refused, direction: 'horizontal' }} />

<Story name="Vertical" args={{ steps: running, direction: 'vertical' }} />

<Story
	name="Mobile"
	args={{ steps: running, direction: 'horizontal' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
