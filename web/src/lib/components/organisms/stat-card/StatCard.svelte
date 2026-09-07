<script lang="ts">
  // Composes atoms (Text, Money, Sparkline) plus the
  // TrendIndicator molecule, which is why it lives in organisms/ rather than molecules/.
  import Money from '$lib/components/atoms/money/Money.svelte';
  import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
  import Text from '$lib/components/atoms/typography/Text.svelte';
  import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
  import type { TrendIndicatorFormat } from '$lib/components/molecules/trend-indicator/trend-indicator.types';

  type Props = {
    label: string;
    amount: number;
    currency?: string;
    locale?: string;
    /** Optional change shown as a TrendIndicator. */
    change?: number;
    changeFormat?: TrendIndicatorFormat;
    /** Optional series rendered as a Sparkline. */
    trend?: number[];
    class?: string;
  };

  let {
    label,
    amount,
    currency = 'EUR',
    locale = 'en-US',
    change,
    changeFormat = 'percent',
    trend,
    class: className = ''
  }: Props = $props();
</script>

<section class={['flex min-w-0 flex-col gap-2 border-t border-slate-400 py-3', className].filter(Boolean).join(' ')} aria-label={label}>
  <Text as="span" size="sm" tone="muted">{label}</Text>

  <div class="flex items-end justify-between gap-3">
    <Money {amount} {currency} {locale} size="xl" weight="semibold" />

    {#if trend && trend.length > 1}
      <Sparkline data={trend} width={88} height={32} />
    {/if}
  </div>

  {#if change !== undefined}
    <TrendIndicator value={change} format={changeFormat} {currency} {locale} size="sm" />
  {/if}
</section>
