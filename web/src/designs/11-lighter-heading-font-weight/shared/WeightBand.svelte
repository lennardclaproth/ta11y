<script lang="ts">
	// Comparison band for issue #11: today's rendering on the left, the chosen weight on the right.
	// Three rows, because three different things are too heavy today for three different reasons:
	//
	//   1. Heading atom     — asks for semibold (600); only a 400 file ships, so the browser fakes it.
	//   2. Bare <h1>/<h2>   — never reaches the atom and inherits the browser default bold (700).
	//   3. Small <h3> label — Tailwind font-semibold on 14px Garamond, faked the same way.
	//
	// All three land on the same regular (400) after this change.
	import Text from '$lib/components/atoms/typography/Text.svelte';

	type Props = {
		/** One of the d11-weight-* classes from heading-weight.css. */
		weightClass: string;
	};

	let { weightClass }: Props = $props();

	const rows = [
		{
			context: 'Page title, modal and drawer title — Heading atom',
			todayNote: 'semibold 600, faked by the browser',
			sample: 'Cashflow',
			sampleClass: 'font-heading text-4xl leading-tight tracking-tight text-slate-900',
			todayWeightClass: 'font-semibold'
		},
		{
			context: 'Ledger and admin titles — bare <h2>, outside the atom',
			todayNote: 'browser default bold 700, also faked',
			sample: 'Transactions',
			sampleClass: 'font-heading text-2xl leading-tight tracking-tight text-slate-900',
			todayWeightClass: 'font-bold'
		},
		{
			context: 'Asset-class drawer labels — small <h3>',
			todayNote: 'semibold 600 at 14px',
			sample: 'Recent changes',
			sampleClass: 'font-heading text-sm leading-tight text-slate-900',
			todayWeightClass: 'font-semibold'
		}
	];
</script>

<section class="border-b border-slate-300 px-4 pt-6 pb-6 lg:px-8">
	<Text size="xs" tone="muted" class="block tracking-wider uppercase">
		Today versus this design
	</Text>

	<dl class="mt-4 flex flex-col gap-5">
		{#each rows as row (row.context)}
			<div class="grid gap-x-6 gap-y-2 border-t border-slate-200 pt-4 sm:grid-cols-[1fr_1fr]">
				<dt class="sm:col-span-2">
					<Text size="sm" tone="strong" class="block">{row.context}</Text>
				</dt>
				<dd>
					<Text size="xs" tone="muted" class="mb-1 block">Today · {row.todayNote}</Text>
					<p class={[row.sampleClass, row.todayWeightClass].join(' ')}>{row.sample}</p>
				</dd>
				<dd class={weightClass}>
					<Text size="xs" tone="muted" class="mb-1 block">
						This design · regular 400, real font file
					</Text>
					<p class={['d11-sample', row.sampleClass].join(' ')}>{row.sample}</p>
				</dd>
			</div>
		{/each}
	</dl>
</section>
