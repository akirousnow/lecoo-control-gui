<script lang="ts">
  import Sparkline from './Sparkline.svelte'
  import { rise } from '$lib/motion'

  let {
    label,
    value,
    unit = '',
    values = [],
    tone = 'text-foreground',
    delay = 0
  }: {
    label: string
    value: string
    unit?: string
    values?: number[]
    tone?: string
    delay?: number
  } = $props()
</script>

<div use:rise={{ delay }} class="rounded-xl border border-border/60 bg-card/50 p-4">
  <p class="text-xs text-muted-foreground">{label}</p>
  <div class="mt-1.5 flex items-baseline gap-1">
    <span class="text-2xl font-semibold tabular-nums {tone}">{value}</span>
    {#if unit}
      <span class="text-xs text-muted-foreground">{unit}</span>
    {/if}
  </div>
  <Sparkline {values} className="mt-3 h-7 w-full {tone}" />
</div>
