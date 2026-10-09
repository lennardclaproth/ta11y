<script lang="ts">
	// Proposal for the Sparkline atom. Today a series whose amounts never changed has min === max,
	// so every point lands on the bottom edge and the line reads as "fell to zero". Here a flat
	// series sits on the middle line instead. Geometry and stroke are otherwise the atom's.
	// Deliberately one neutral slate tone: a rising expense is not an error, a falling one not a
	// success, and this page passes no judgement on either.
	type Props = {
		/** Amounts, oldest to newest. */
		data: number[];
		width?: number;
		height?: number;
		ariaLabel: string;
		class?: string;
	};

	let { data, width = 72, height = 24, ariaLabel, class: className = '' }: Props = $props();

	const stroke = 2;

	const points = $derived.by(() => {
		if (data.length === 0) return [] as Array<{ x: number; y: number }>;
		const min = Math.min(...data);
		const max = Math.max(...data);
		const innerW = width - stroke * 2;
		const innerH = height - stroke * 2;
		const step = data.length > 1 ? innerW / (data.length - 1) : 0;
		return data.map((value, index) => ({
			x: stroke + index * step,
			// A flat series has no shape to show, so it draws through the middle.
			y: max === min ? stroke + innerH / 2 : stroke + innerH - ((value - min) / (max - min)) * innerH
		}));
	});

	const path = $derived(
		points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(' ')
	);
</script>

<svg
	class={['inline-block shrink-0', className].filter(Boolean).join(' ')}
	{width}
	{height}
	viewBox={`0 0 ${width} ${height}`}
	fill="none"
	role="img"
	aria-label={ariaLabel}
>
	<path
		d={path}
		class="text-slate-400"
		stroke="currentColor"
		stroke-width={stroke}
		stroke-linecap="round"
		stroke-linejoin="round"
	/>
</svg>
