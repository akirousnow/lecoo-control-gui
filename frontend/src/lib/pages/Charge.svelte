<script lang="ts">
  import { Check } from '@lucide/svelte'
  import Range from '$lib/components/Range.svelte'
  import { Button } from '$lib/components/ui/button'
  import { chargePresets, describeCharge } from '$lib/labels'
  import { rise } from '$lib/motion'
  import { store } from '$lib/store.svelte'
  import { cn } from '$lib/utils'

  let min = $state(40)
  let max = $state(80)

  const charge = $derived(store.settings?.charge ?? null)
  const activePreset = $derived(charge?.kind === 'preset' ? charge.preset : '')
  const valid = $derived(min >= 20 && max <= 100 && min < max)

  $effect(() => {
    const target = charge
    if (target?.kind === 'preserve' && target.min != null && target.max != null) {
      min = target.min
      max = target.max
    }
  })

  async function choose(preset: string) {
    if (preset === activePreset) {
      return
    }
    await store.setChargePreset(preset)
  }

  async function applyRange() {
    if (!valid) {
      return
    }
    await store.setChargeRange(min, max)
  }
</script>

<div class="flex flex-col gap-5">
  <header class="flex items-end justify-between">
    <div>
      <h1 class="text-lg font-semibold">充电</h1>
      <p class="mt-0.5 text-xs text-muted-foreground">限制充电上限，延缓电池损耗</p>
    </div>
    <span class="text-[11px] text-muted-foreground">{describeCharge(charge)}</span>
  </header>

  <div class="grid gap-2 sm:grid-cols-3 xl:grid-cols-5">
    {#each chargePresets as preset, index (preset.id)}
      <button
        type="button"
        use:rise={{ delay: index * 0.04 }}
        class={cn(
          'rounded-xl border px-3 py-3 text-left transition-colors',
          activePreset === preset.id
            ? 'border-primary/60 bg-primary/10'
            : 'border-border/60 bg-card/50 hover:border-border hover:bg-card/80'
        )}
        disabled={store.busy}
        onclick={() => choose(preset.id)}
      >
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium">{preset.label}</span>
          {#if activePreset === preset.id}
            <Check class="size-3.5 text-primary" />
          {/if}
        </div>
        <p class="mt-1 text-xl font-semibold tabular-nums">{preset.percent}%</p>
        <p class="mt-1 text-[10px] leading-snug text-muted-foreground">{preset.hint}</p>
      </button>
    {/each}
  </div>

  <section use:rise={{ delay: 0.2 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
    <div class="flex items-center justify-between">
      <p class="text-xs text-muted-foreground">自定义区间</p>
      <span class="text-xs tabular-nums">{min}% .. {max}%</span>
    </div>

    <div class="mt-4 space-y-4">
      <div>
        <div class="flex items-center justify-between text-[11px] text-muted-foreground">
          <span>下限</span>
          <span class="tabular-nums">{min}%</span>
        </div>
        <div class="mt-1.5">
          <Range bind:value={min} min={20} max={100} />
        </div>
      </div>
      <div>
        <div class="flex items-center justify-between text-[11px] text-muted-foreground">
          <span>上限</span>
          <span class="tabular-nums">{max}%</span>
        </div>
        <div class="mt-1.5">
          <Range bind:value={max} min={20} max={100} />
        </div>
      </div>
    </div>

    {#if !valid}
      <p class="mt-3 rounded-md bg-destructive/10 px-2.5 py-2 text-[11px] text-destructive">
        区间需要落在 20 到 100 之间，且下限小于上限。
      </p>
    {/if}

    <div class="mt-4 flex justify-end">
      <Button size="sm" disabled={!valid || store.busy} onclick={applyRange}>应用区间</Button>
    </div>
  </section>

  <section class="grid grid-cols-3 gap-3">
    <div class="rounded-xl border border-border/60 bg-card/50 px-3 py-3">
      <p class="text-[11px] text-muted-foreground">目标</p>
      <p class="mt-1 text-xs">{store.battery?.target ?? '—'}</p>
    </div>
    <div class="rounded-xl border border-border/60 bg-card/50 px-3 py-3">
      <p class="text-[11px] text-muted-foreground">生效</p>
      <p class="mt-1 text-xs">{store.battery?.effective ?? '—'}</p>
    </div>
    <div class="rounded-xl border border-border/60 bg-card/50 px-3 py-3">
      <p class="text-[11px] text-muted-foreground">电池</p>
      <p class="mt-1 text-xs tabular-nums">{store.battery?.percent ?? '—'}%</p>
    </div>
  </section>

  <p class="text-[11px] leading-relaxed text-muted-foreground">
    桌面模式把上限压到 40%，适合长期插电使用。目标是请求值，生效是硬件确认后的实际值，两者不一致时以生效为准。
  </p>
</div>
