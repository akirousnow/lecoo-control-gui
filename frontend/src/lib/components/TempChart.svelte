<script lang="ts">
  import { LineChart } from 'layerchart'
  import type { HistoryPoint } from '$lib/store.svelte'

  let { points = [] }: { points?: HistoryPoint[] } = $props()

  const TARGET_POINTS = 56

  function aggregate(source: HistoryPoint[]): HistoryPoint[] {
    const usable = source.filter((point) => point.cpu != null && point.sys != null)
    if (usable.length <= TARGET_POINTS) {
      return usable
    }
    const size = Math.ceil(usable.length / TARGET_POINTS)
    const merged: HistoryPoint[] = []
    for (let index = 0; index < usable.length; index += size) {
      const slice = usable.slice(index, index + size)
      const count = slice.length
      if (count === 0) {
        continue
      }
      let cpu = 0
      let sys = 0
      for (const point of slice) {
        cpu += point.cpu ?? 0
        sys += point.sys ?? 0
      }
      const last = slice[count - 1]
      merged.push({
        t: last.t,
        cpu: Math.round(cpu / count),
        sys: Math.round(sys / count),
        cpuRpm: last.cpuRpm,
        gpuRpm: last.gpuRpm
      })
    }
    return merged
  }

  function clock(value: unknown): string {
    if (!(value instanceof Date)) {
      return String(value)
    }
    const pad = (input: number) => String(input).padStart(2, '0')
    return `${pad(value.getHours())}:${pad(value.getMinutes())}:${pad(value.getSeconds())}`
  }

  const data = $derived(
    aggregate(points).map((point) => ({
      date: new Date(point.t),
      cpu: point.cpu ?? 0,
      sys: point.sys ?? 0
    }))
  )

  const series = [
    { key: 'cpu', label: 'CPU', value: 'cpu', color: 'var(--chart-cpu)' },
    { key: 'sys', label: '系统', value: 'sys', color: 'var(--chart-sys)' }
  ]
</script>

{#if data.length > 1}
  <div class="h-full w-full">
    <LineChart
      {data}
      x="date"
      y="cpu"
      {series}
      seriesLayout="overlap"
      yDomain={[0, 100]}
      highlight={false}
      tooltipContext={false}
      padding={{ top: 10, right: 16, bottom: 22, left: 34 }}
      props={{
        spline: { strokeWidth: 1.8 },
        xAxis: { ticks: 5, format: clock, tickSpacing: 72 },
        yAxis: { ticks: 5 }
      }}
    />
  </div>
{:else}
  <div class="flex h-full items-center justify-center text-xs text-muted-foreground">
    等待采样数据
  </div>
{/if}
