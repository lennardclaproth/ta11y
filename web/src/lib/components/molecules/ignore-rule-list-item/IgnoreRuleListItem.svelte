<script lang="ts">
	// One rule in the index on the left of the rules page: its name, what it matches on,
	// and how many transactions it has ignored so far. A rule that is switched off says so
	// in a word rather than by being a shade lighter.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { ruleDirectionLabel, ruleFieldLabel, allBanksLabel } from '$lib/api/ignoreRules';
	import type { IgnoreRule } from '$lib/api/types';

	type Props = {
		rule: IgnoreRule;
		/** This rule is the one open on the right. */
		active?: boolean;
		onSelect?: (rule: IgnoreRule) => void;
		class?: string;
	};

	let { rule, active = false, onSelect, class: className = '' }: Props = $props();

	const classes = $derived(
		[
			'flex w-full items-start gap-3 px-4 py-3 text-left transition-colors',
			'focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-500',
			active ? 'bg-amber-50' : 'hover:bg-slate-50',
			className
		]
			.filter(Boolean)
			.join(' ')
	);

	const scope = $derived(
		`${ruleDirectionLabel(rule.direction)} · ${rule.source ? rule.source.toUpperCase() : allBanksLabel}`
	);
</script>

<button
	type="button"
	class={classes}
	aria-current={active ? 'true' : undefined}
	onclick={() => onSelect?.(rule)}
>
	<span class="min-w-0 flex-1">
		<span class="flex items-center gap-2">
			<span class="truncate font-heading text-lg text-slate-900">{rule.name}</span>
			{#if !rule.enabled}
				<Badge intent="neutral" variant="outline" size="sm">Off</Badge>
			{/if}
		</span>
		<span class="mt-0.5 block truncate text-sm text-slate-600">
			{ruleFieldLabel(rule.match_field)} contains “{rule.contains}”
		</span>
		<span class="block truncate text-xs text-slate-500">{scope}</span>
	</span>
	<span class="shrink-0 text-right">
		<span class="block text-sm text-slate-900 tabular-nums">{rule.ignored_total}</span>
		<span class="block text-xs text-slate-500">ignored</span>
	</span>
</button>
