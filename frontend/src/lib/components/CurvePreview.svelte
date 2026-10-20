<script lang="ts">
  import { LineChart } from 'layerchart'
  import type { CurvePoint } from '$lib/domain'

  let { points = [] }: { points?: CurvePoint[] } = $props()

  const data = $derived(
    [...points]
      .sort((a, b) => a.tempC - b.tempC)
      .map((point) => ({ temp: point.tempC, duty: point.duty }))
  )

  const series = [{ key: 'duty', label: '占空比', value: 'duty', color: 'var(--chart-cpu)' }]

  const degrees = (value: unknown) => `${value}°`
  const percent = (value: unknown) => `${value}%`
</script>

{#if data.length > 1}
  <div class="h-full w-full">
    <LineChart
      {data}
      x="temp"
      y="duty"
      {series}
      xDomain={[20, 110]}
      yDomain={[0, 100]}
      highlight={false}
      tooltipContext={false}
      padding={{ top: 8, right: 12, bottom: 20, left: 30 }}
      props={{
        spline: { strokeWidth: 2 },
        xAxis: { ticks: 4, format: degrees },
        yAxis: { ticks: 4, format: percent }
      }}
    />
  </div>
{:else}
  <div class="flex h-full items-center justify-center text-[11px] text-muted-foreground">
    至少需要两个节点
  </div>
{/if}
