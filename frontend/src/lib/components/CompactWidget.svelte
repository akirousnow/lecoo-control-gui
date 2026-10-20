<script lang="ts">
  import { Maximize2, Pin, PinOff, X } from '@lucide/svelte'
  import appIcon from '../../assets/app-icon.png'
  import { api } from '$lib/api'
  import { powerLabel } from '$lib/labels'
  import { store } from '$lib/store.svelte'
  import { view } from '$lib/view.svelte'

  const blank = (value: number | null | undefined) => (value == null ? '—' : String(value))

  const percent = $derived(store.battery?.percent ?? 0)

  const batteryText = $derived(
    store.battery?.percent == null ? '—' : `${store.battery.percent}%`
  )

  const cells = $derived([
    {
      label: 'CPU 温度',
      value: blank(store.sample?.cpuTempC),
      unit: '°C',
      tone: 'text-orange-500 dark:text-orange-300'
    },
    {
      label: '系统温度',
      value: blank(store.sample?.sysTempC),
      unit: '°C',
      tone: 'text-amber-600 dark:text-amber-200'
    },
    {
      label: 'CPU 风扇',
      value: blank(store.sample?.cpuRpm),
      unit: 'RPM',
      tone: 'text-sky-600 dark:text-sky-300'
    },
    {
      label: 'GPU 风扇',
      value: blank(store.sample?.gpuRpm),
      unit: 'RPM',
      tone: 'text-indigo-500 dark:text-indigo-300'
    }
  ])
</script>

<div class="flex h-screen w-screen flex-col overflow-hidden bg-background text-foreground">
  <header
    class="flex h-8 shrink-0 items-center justify-between border-b border-border/60 pl-2.5"
    style="--wails-draggable:drag"
  >
    <div class="flex min-w-0 items-center gap-1.5">
      <img src={appIcon} alt="" class="size-3.5 shrink-0 rounded-[3px]" />
      <span class="truncate text-[11px] font-medium">来酷</span>
      <span class="shrink-0 text-[10px] text-muted-foreground">
        {store.monitor.running ? '实时' : '已停止'}
      </span>
    </div>
    <div class="flex h-full items-stretch" style="--wails-draggable:no-drag">
      <button
        type="button"
        class="flex w-7 items-center justify-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        aria-label="置顶"
        title={view.pinned ? '取消置顶' : '置顶显示'}
        onclick={() => view.togglePinned()}
      >
        {#if view.pinned}
          <Pin class="size-3" />
        {:else}
          <PinOff class="size-3" />
        {/if}
      </button>
      <button
        type="button"
        class="flex w-7 items-center justify-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        aria-label="展开完整界面"
        title="展开完整界面"
        onclick={() => view.toggleCompact()}
      >
        <Maximize2 class="size-3" />
      </button>
      <button
        type="button"
        class="flex w-7 items-center justify-center text-muted-foreground transition-colors hover:bg-destructive hover:text-white"
        aria-label="关闭"
        onclick={() => api.quit()}
      >
        <X class="size-3" />
      </button>
    </div>
  </header>

  <div class="grid flex-1 grid-cols-2 gap-px bg-border/50">
    {#each cells as cell (cell.label)}
      <div class="flex flex-col justify-center gap-0.5 bg-background px-3">
        <span class="text-[10px] text-muted-foreground">{cell.label}</span>
        <span class="flex items-baseline gap-0.5">
          <span class="text-xl font-semibold tabular-nums {cell.tone}">{cell.value}</span>
          <span class="text-[10px] text-muted-foreground">{cell.unit}</span>
        </span>
      </div>
    {/each}
  </div>

  <footer class="flex h-[30px] shrink-0 items-center gap-2 border-t border-border/60 px-3">
    <span class="text-[10px] text-muted-foreground">电池</span>
    <span class="text-[11px] font-medium tabular-nums">{batteryText}</span>
    <div class="h-1 flex-1 overflow-hidden rounded-full bg-muted">
      <div
        class="h-full rounded-full bg-primary transition-[width] duration-500"
        style={`width:${percent}%`}
      ></div>
    </div>
    <span class="text-[10px] text-muted-foreground">
      {powerLabel(store.settings?.power.kind ?? '')}
    </span>
  </footer>
</div>
