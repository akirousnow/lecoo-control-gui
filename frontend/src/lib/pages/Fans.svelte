<script lang="ts">
  import { Check } from '@lucide/svelte'
  import CurvePreview from '$lib/components/CurvePreview.svelte'
  import Range from '$lib/components/Range.svelte'
  import { Button } from '$lib/components/ui/button'
  import { Switch } from '$lib/components/ui/switch'
  import type { CurvePoint, FanCurve, FanTarget } from '$lib/domain'
  import { describeFan, fanModes, fanTargets } from '$lib/labels'
  import { rise } from '$lib/motion'
  import { store } from '$lib/store.svelte'
  import { cn } from '$lib/utils'

  let target = $state<FanTarget>('both')
  let mode = $state('auto')
  let pwm = $state(128)
  let turboOpen = $state(false)
  let turboConfirmed = $state(false)

  const fallback: FanCurve = {
    enabled: false,
    target: 'both',
    points: [
      { tempC: 40, duty: 30 },
      { tempC: 55, duty: 40 },
      { tempC: 65, duty: 55 },
      { tempC: 75, duty: 75 },
      { tempC: 85, duty: 95 }
    ],
    minDuty: 25,
    maxDuty: 100,
    hysteresis: 8,
    panicTemp: 95
  }

  const numberField =
    'h-7 rounded-md border border-border/60 bg-background px-2 text-right font-mono text-[11px] tabular-nums outline-none focus:border-ring'

  function cloneCurve(source: FanCurve): FanCurve {
    return {
      enabled: source.enabled,
      target: source.target,
      points: source.points.map((point) => ({ tempC: point.tempC, duty: point.duty })),
      minDuty: source.minDuty,
      maxDuty: source.maxDuty,
      hysteresis: source.hysteresis,
      panicTemp: source.panicTemp
    }
  }

  let draft = $state<FanCurve>(cloneCurve(fallback))
  let ready = $state(false)

  $effect(() => {
    const state = store.curve
    if (!state || ready) {
      return
    }
    draft = cloneCurve(state.curve)
    ready = true
  })

  const activeKind = $derived(
    target === 'gpu' ? store.settings?.fanGpu.kind : store.settings?.fanCpu.kind
  )

  $effect(() => {
    const kind = activeKind
    if (!kind) {
      return
    }
    if (kind === 'custom') {
      const value = target === 'gpu' ? store.settings?.fanGpu.pwm : store.settings?.fanCpu.pwm
      if (value != null) {
        pwm = value
      }
    }
    if (kind !== 'unknown' && kind !== 'turbo') {
      mode = kind
    }
  })

  const curveEnabled = $derived(store.curve?.curve.enabled ?? false)
  const liveTemp = $derived(store.curve?.tempC ?? null)
  const liveDuty = $derived(store.curve?.duty ?? 0)
  const livePwm = $derived(store.curve?.appliedPwm ?? -1)
  const curveError = $derived(store.curve?.lastError ?? '')

  const sortedPoints = $derived([...draft.points].sort((a, b) => a.tempC - b.tempC))

  async function apply() {
    await store.setFan(target, mode, pwm)
  }

  async function applyTurbo() {
    await store.setFan(target, 'turbo', 0)
    turboConfirmed = false
  }

  async function toggleCurve(next: boolean): Promise<void> {
    await store.setCurve({ ...cloneCurve(draft), enabled: next })
    if (store.curve) {
      draft = cloneCurve(store.curve.curve)
    }
  }

  async function saveCurve(): Promise<void> {
    await store.setCurve(cloneCurve(draft))
    if (store.curve) {
      draft = cloneCurve(store.curve.curve)
    }
  }

  function resetDraft(): void {
    if (store.curve) {
      draft = cloneCurve(store.curve.curve)
    }
  }

  function addPoint(): void {
    const last = sortedPoints[sortedPoints.length - 1]
    const point: CurvePoint = last
      ? { tempC: Math.min(110, last.tempC + 10), duty: Math.min(100, last.duty + 10) }
      : { tempC: 50, duty: 40 }
    draft.points = [...draft.points, point]
  }

  function removePoint(index: number): void {
    if (draft.points.length <= 2) {
      return
    }
    draft.points = draft.points.filter((_, position) => position !== index)
  }
</script>

<div class="flex flex-col gap-5">
  <header>
    <h1 class="text-lg font-semibold">风扇</h1>
    <p class="mt-0.5 text-xs text-muted-foreground">手动设定目标与模式，或者交给温度曲线自动调节</p>
  </header>

  <section use:rise class="rounded-xl border border-border/60 bg-card/50 p-4">
    <div class="flex items-center justify-between">
      <p class="text-xs text-muted-foreground">目标</p>
      {#if curveEnabled}
        <span class="text-[10px] text-amber-600 dark:text-amber-300">曲线启用中，手动设置已暂停</span>
      {/if}
    </div>
    <div class="mt-2 grid grid-cols-3 gap-1 rounded-lg bg-muted/40 p-1">
      {#each fanTargets as item (item.id)}
        <button
          type="button"
          class={cn(
            'rounded-md px-3 py-1.5 text-xs transition-colors',
            target === item.id
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          )}
          onclick={() => (target = item.id)}
        >
          {item.label}
        </button>
      {/each}
    </div>
  </section>

  <section use:rise={{ delay: 0.06 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
    <p class="text-xs text-muted-foreground">模式</p>
    <div class="mt-2 grid grid-cols-3 gap-1 rounded-lg bg-muted/40 p-1">
      {#each fanModes as item (item.id)}
        <button
          type="button"
          class={cn(
            'rounded-md px-3 py-1.5 text-xs transition-colors',
            mode === item.id
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          )}
          onclick={() => (mode = item.id)}
        >
          {item.label}
        </button>
      {/each}
    </div>
    <p class="mt-2 text-[11px] text-muted-foreground">
      {fanModes.find((item) => item.id === mode)?.hint ?? ''}
    </p>

    {#if mode === 'custom'}
      <div class="mt-4">
        <div class="flex items-center justify-between">
          <span class="text-[11px] text-muted-foreground">PWM 占空比</span>
          <span class="text-xs tabular-nums">{pwm} / {store.dutyMax}</span>
        </div>
        <div class="mt-2">
          <Range bind:value={pwm} min={0} max={store.dutyMax} />
        </div>
      </div>
    {/if}

    {#if mode === 'full'}
      <p class="mt-3 rounded-md bg-amber-500/10 px-2.5 py-2 text-[11px] text-amber-700 dark:text-amber-300">
        全速会把风扇拉到 100% 占空比，噪音和功耗都会明显上升。
      </p>
    {/if}

    <div class="mt-4 flex items-center justify-between">
      <div class="text-[11px] text-muted-foreground">
        当前 {describeFan(store.settings?.fanCpu ?? null, store.dutyMax)}
        {#if target !== 'cpu'}
          <span class="mx-1 text-border">·</span>
          GPU {describeFan(store.settings?.fanGpu ?? null, store.dutyMax)}
        {/if}
      </div>
      <Button size="sm" disabled={store.busy || curveEnabled} onclick={apply}>应用</Button>
    </div>
  </section>

  <section use:rise={{ delay: 0.12 }} class="grid grid-cols-2 gap-3">
    <div class="rounded-xl border border-border/60 bg-card/50 px-4 py-3">
      <p class="text-[11px] text-muted-foreground">CPU 转速</p>
      <p class="mt-1 text-xl font-semibold tabular-nums">
        {store.sample?.cpuRpm ?? '—'}
        <span class="text-xs font-normal text-muted-foreground">RPM</span>
      </p>
    </div>
    <div class="rounded-xl border border-border/60 bg-card/50 px-4 py-3">
      <p class="text-[11px] text-muted-foreground">GPU 转速</p>
      <p class="mt-1 text-xl font-semibold tabular-nums">
        {store.sample?.gpuRpm ?? '—'}
        <span class="text-xs font-normal text-muted-foreground">RPM</span>
      </p>
    </div>
  </section>

  <section use:rise={{ delay: 0.16 }} class="overflow-hidden rounded-xl border border-border/60 bg-card/50">
    <div class="flex items-start justify-between gap-4 px-4 py-3.5">
      <div class="min-w-0">
        <p class="text-xs font-medium">温度曲线</p>
        <p class="mt-1 max-w-2xl text-[11px] leading-relaxed text-muted-foreground">
          按 CPU 温度在节点之间线性插值出占空比，低于下限取下限，达到触发温度直接满速。只在本程序运行时生效，退出会恢复自动风扇。
        </p>
      </div>
      <Switch checked={curveEnabled} disabled={store.busy} onCheckedChange={toggleCurve} />
    </div>

    {#if curveEnabled}
      <div class="grid grid-cols-3 divide-x divide-border/40 border-y border-border/40 bg-muted/20">
        <div class="px-4 py-2">
          <p class="text-[10px] text-muted-foreground">当前温度</p>
          <p class="mt-0.5 text-sm font-semibold tabular-nums">
            {liveTemp ?? '—'}<span class="ml-0.5 text-[10px] font-normal text-muted-foreground">°C</span>
          </p>
        </div>
        <div class="px-4 py-2">
          <p class="text-[10px] text-muted-foreground">目标占空比</p>
          <p class="mt-0.5 text-sm font-semibold tabular-nums">
            {liveDuty}<span class="ml-0.5 text-[10px] font-normal text-muted-foreground">%</span>
          </p>
        </div>
        <div class="px-4 py-2">
          <p class="text-[10px] text-muted-foreground">已下发</p>
          <p class="mt-0.5 text-sm font-semibold tabular-nums">
            {livePwm >= 0 ? livePwm : '—'}<span class="ml-0.5 text-[10px] font-normal text-muted-foreground">/{store.dutyMax}</span>
          </p>
        </div>
      </div>
    {/if}

    {#if curveError}
      <p class="border-y border-destructive/20 bg-destructive/10 px-4 py-2 text-[11px] text-destructive">
        {curveError}
      </p>
    {/if}

    <div class="grid gap-5 p-4 lg:grid-cols-2">
      <div>
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-medium">温度节点</p>
          <button
            type="button"
            class="rounded-md px-2 py-1 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            onclick={addPoint}
          >
            + 添加
          </button>
        </div>
        <div class="mt-2 grid grid-cols-[4.5rem_4.5rem_auto] gap-2 text-[10px] text-muted-foreground">
          <span>温度 °C</span>
          <span>占空比 %</span>
        </div>
        <div class="mt-1 flex flex-col gap-1.5">
          {#each draft.points as point, index (index)}
            <div class="grid grid-cols-[4.5rem_4.5rem_auto] items-center gap-2">
              <input
                type="number"
                bind:value={point.tempC}
                min="20"
                max="110"
                class="{numberField} w-[4.5rem]"
              />
              <input
                type="number"
                bind:value={point.duty}
                min="0"
                max="100"
                class="{numberField} w-[4.5rem]"
              />
              <button
                type="button"
                class="justify-self-start rounded-md px-2 py-1 text-[10px] text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-muted-foreground"
                disabled={draft.points.length <= 2}
                onclick={() => removePoint(index)}
              >
                移除
              </button>
            </div>
          {/each}
        </div>
      </div>

      <div class="min-w-0">
        <p class="text-[11px] font-medium">曲线预览</p>
        <div class="mt-2 h-[176px] rounded-lg border border-border/40 bg-background/30">
          <CurvePreview points={draft.points} />
        </div>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-3 border-t border-border/40 px-4 py-3 sm:grid-cols-4">
      <label class="flex flex-col gap-1">
        <span class="text-[10px] text-muted-foreground">最低占空比 %</span>
        <input
          type="number"
          bind:value={draft.minDuty}
          min="0"
          max="100"
          class="{numberField} w-full max-w-[6.5rem]"
        />
      </label>
      <label class="flex flex-col gap-1">
        <span class="text-[10px] text-muted-foreground">最高占空比 %</span>
        <input
          type="number"
          bind:value={draft.maxDuty}
          min="0"
          max="100"
          class="{numberField} w-full max-w-[6.5rem]"
        />
      </label>
      <label class="flex flex-col gap-1">
        <span class="text-[10px] text-muted-foreground">满速触发 °C</span>
        <input
          type="number"
          bind:value={draft.panicTemp}
          min="60"
          max="110"
          class="{numberField} w-full max-w-[6.5rem]"
        />
      </label>
      <label class="flex flex-col gap-1">
        <span class="text-[10px] text-muted-foreground">死区 PWM</span>
        <input
          type="number"
          bind:value={draft.hysteresis}
          min="1"
          max="64"
          class="{numberField} w-full max-w-[6.5rem]"
        />
      </label>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border/40 px-4 py-3">
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-2">
          <span class="text-[11px] text-muted-foreground">作用于</span>
          <div class="flex gap-1 rounded-lg bg-muted/40 p-0.5">
            {#each fanTargets as item (item.id)}
              <button
                type="button"
                class={cn(
                  'rounded-md px-2 py-1 text-[11px] transition-colors',
                  draft.target === item.id
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                )}
                onclick={() => (draft.target = item.id)}
              >
                {item.label}
              </button>
            {/each}
          </div>
        </div>
        <span class="text-[10px] text-muted-foreground">最快每 4 秒下发一次</span>
      </div>
      <div class="flex gap-2">
        <Button size="sm" variant="ghost" onclick={resetDraft}>还原</Button>
        <Button size="sm" disabled={store.busy} onclick={saveCurve}>
          <Check class="size-3.5" />
          保存曲线
        </Button>
      </div>
    </div>
  </section>

  <section class="rounded-xl border border-destructive/30 bg-destructive/5 p-4">
    <button
      type="button"
      class="flex w-full items-center justify-between text-left"
      onclick={() => (turboOpen = !turboOpen)}
    >
      <span class="text-xs font-medium text-destructive">涡轮模式</span>
      <span class="text-[11px] text-muted-foreground">{turboOpen ? '收起' : '展开'}</span>
    </button>
    {#if turboOpen}
      <p class="mt-2 text-[11px] leading-relaxed text-muted-foreground">
        CLI 自己把这个模式标为不安全。风扇会以超出设计上限的速度运行，长期使用可能缩短风扇寿命。默认不记忆，每次都要重新确认。
      </p>
      <label class="mt-3 flex items-center gap-2 text-[11px] text-muted-foreground">
        <input type="checkbox" bind:checked={turboConfirmed} class="size-3.5 accent-destructive" />
        我了解风险，仍要启用
      </label>
      <Button
        class="mt-3"
        size="sm"
        variant="destructive"
        disabled={!turboConfirmed || store.busy || curveEnabled}
        onclick={applyTurbo}
      >
        以涡轮模式运行 {target === 'both' ? '双风扇' : target.toUpperCase()}
      </Button>
    {/if}
  </section>
</div>
