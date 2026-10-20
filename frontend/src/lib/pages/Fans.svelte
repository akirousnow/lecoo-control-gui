<script lang="ts">
  import { Button } from '$lib/components/ui/button'
  import Range from '$lib/components/Range.svelte'
  import type { FanTarget } from '$lib/domain'
  import { describeFan, fanModes, fanTargets } from '$lib/labels'
  import { rise } from '$lib/motion'
  import { store } from '$lib/store.svelte'
  import { cn } from '$lib/utils'

  let target = $state<FanTarget>('both')
  let mode = $state('auto')
  let pwm = $state(128)
  let turboOpen = $state(false)
  let turboConfirmed = $state(false)

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

  const rpmForTarget = $derived(
    target === 'gpu' ? store.sample?.gpuRpm : target === 'cpu' ? store.sample?.cpuRpm : null
  )

  async function apply() {
    await store.setFan(target, mode, pwm)
  }

  async function applyTurbo() {
    await store.setFan(target, 'turbo', 0)
    turboConfirmed = false
  }
</script>

<div class="flex flex-col gap-5">
  <header>
    <h1 class="text-lg font-semibold">风扇</h1>
    <p class="mt-0.5 text-xs text-muted-foreground">选择目标与模式，松手下发并读回确认</p>
  </header>

  <section use:rise class="rounded-xl border border-border/60 bg-card/50 p-4">
    <p class="text-xs text-muted-foreground">目标</p>
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
          <span class="text-xs tabular-nums">{pwm} / 255</span>
        </div>
        <div class="mt-2">
          <Range bind:value={pwm} min={0} max={255} />
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
        当前 {describeFan(store.settings?.fanCpu ?? null)}
        {#if target !== 'cpu'}
          <span class="mx-1 text-border">·</span>
          GPU {describeFan(store.settings?.fanGpu ?? null)}
        {/if}
      </div>
      <Button size="sm" disabled={store.busy} onclick={apply}>应用</Button>
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
        disabled={!turboConfirmed || store.busy}
        onclick={applyTurbo}
      >
        以涡轮模式运行 {target === 'both' ? '双风扇' : target.toUpperCase()}
      </Button>
    {/if}
  </section>
</div>
