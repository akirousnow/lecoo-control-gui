<script lang="ts">
  import StatCard from '$lib/components/StatCard.svelte'
  import TempChart from '$lib/components/TempChart.svelte'
  import { describeCharge, describeFan, powerLabel } from '$lib/labels'
  import { rise } from '$lib/motion'
  import { store } from '$lib/store.svelte'

  const blank = (value: number | null | undefined) => (value == null ? '—' : String(value))

  const cpuTemps = $derived(store.history.map((point) => point.cpu ?? 0))
  const sysTemps = $derived(store.history.map((point) => point.sys ?? 0))
  const cpuRpms = $derived(store.history.map((point) => point.cpuRpm ?? 0))
  const gpuRpms = $derived(store.history.map((point) => point.gpuRpm ?? 0))

  const updated = $derived(
    store.lastSampleAt
      ? `更新于 ${new Date(store.lastSampleAt).toLocaleTimeString('zh-CN', { hour12: false })}`
      : '等待数据'
  )

  const percent = $derived(store.battery?.percent ?? 0)

  const pending = $derived(
    store.battery != null && store.battery.target !== store.battery.effective
  )
</script>

<div class="flex flex-col gap-5">
  <header class="flex items-end justify-between">
    <div>
      <h1 class="text-lg font-semibold">总览</h1>
      <p class="mt-0.5 text-xs text-muted-foreground">实时温度、转速与当前配置</p>
    </div>
    <span class="text-[11px] text-muted-foreground">{updated}</span>
  </header>

  <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
    <StatCard
      label="CPU 温度"
      value={blank(store.sample?.cpuTempC)}
      unit="°C"
      values={cpuTemps}
      tone="text-orange-500 dark:text-orange-300"
    />
    <StatCard
      label="系统温度"
      value={blank(store.sample?.sysTempC)}
      unit="°C"
      values={sysTemps}
      tone="text-amber-600 dark:text-amber-200"
      delay={0.04}
    />
    <StatCard
      label="CPU 风扇"
      value={blank(store.sample?.cpuRpm)}
      unit="RPM"
      values={cpuRpms}
      tone="text-sky-600 dark:text-sky-300"
      delay={0.08}
    />
    <StatCard
      label="GPU 风扇"
      value={blank(store.sample?.gpuRpm)}
      unit="RPM"
      values={gpuRpms}
      tone="text-indigo-500 dark:text-indigo-300"
      delay={0.12}
    />
  </div>

  <div class="grid gap-3 lg:grid-cols-3">
    <div use:rise={{ delay: 0.16 }} class="rounded-xl border border-border/60 bg-card/50 p-4 lg:col-span-2">
      <div class="flex items-center justify-between">
        <p class="text-xs text-muted-foreground">温度走势</p>
        <div class="flex items-center gap-3 text-[11px] text-muted-foreground">
          <span class="flex items-center gap-1.5">
            <span class="size-1.5 rounded-full bg-[var(--chart-cpu)]"></span>CPU
          </span>
          <span class="flex items-center gap-1.5">
            <span class="size-1.5 rounded-full bg-[var(--chart-sys)]"></span>系统
          </span>
        </div>
      </div>
      <div class="mt-2 h-52">
        <TempChart points={store.history} />
      </div>
    </div>

    <div use:rise={{ delay: 0.2 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
      <div class="flex items-center justify-between">
        <p class="text-xs text-muted-foreground">电池</p>
        <span class="text-[11px] text-muted-foreground">
          {describeCharge(store.settings?.charge ?? null)}
        </span>
      </div>
      <div class="mt-2 flex items-baseline gap-1">
        <span class="text-2xl font-semibold tabular-nums">{blank(store.battery?.percent)}</span>
        <span class="text-xs text-muted-foreground">%</span>
      </div>
      <div class="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
        <div
          class="h-full rounded-full bg-primary transition-[width] duration-500"
          style={`width:${percent}%`}
        ></div>
      </div>
      <dl class="mt-4 space-y-1.5 text-[11px]">
        <div class="flex justify-between gap-2">
          <dt class="text-muted-foreground">目标</dt>
          <dd class="truncate text-foreground/90">{store.battery?.target ?? '—'}</dd>
        </div>
        <div class="flex justify-between gap-2">
          <dt class="text-muted-foreground">生效</dt>
          <dd class="truncate text-foreground/90">{store.battery?.effective ?? '—'}</dd>
        </div>
      </dl>
      {#if pending}
        <p class="mt-3 rounded-md bg-amber-500/10 px-2 py-1.5 text-[11px] text-amber-700 dark:text-amber-300">
          已下发，等待生效
        </p>
      {/if}
    </div>
  </div>

  <div use:rise={{ delay: 0.24 }} class="grid grid-cols-2 gap-3 xl:grid-cols-4">
    <div class="rounded-xl border border-border/60 bg-card/50 px-4 py-3">
      <p class="text-[11px] text-muted-foreground">电源模式</p>
      <p class="mt-1 text-sm font-medium">{powerLabel(store.settings?.power.kind ?? '')}</p>
    </div>
    <div class="rounded-xl border border-border/60 bg-card/50 px-4 py-3">
      <p class="text-[11px] text-muted-foreground">CPU 风扇</p>
      <p class="mt-1 text-sm font-medium">{describeFan(store.settings?.fanCpu ?? null)}</p>
    </div>
    <div class="rounded-xl border border-border/60 bg-card/50 px-4 py-3">
      <p class="text-[11px] text-muted-foreground">GPU 风扇</p>
      <p class="mt-1 text-sm font-medium">{describeFan(store.settings?.fanGpu ?? null)}</p>
    </div>
    <div class="rounded-xl border border-border/60 bg-card/50 px-4 py-3">
      <p class="text-[11px] text-muted-foreground">充电限制</p>
      <p class="mt-1 text-sm font-medium">{describeCharge(store.settings?.charge ?? null)}</p>
    </div>
  </div>
</div>
