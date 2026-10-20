<script lang="ts">
  import { store } from '$lib/store.svelte'
  import { cn } from '$lib/utils'
  import { routes } from '$lib/routes'

  let { current, onNavigate }: { current: string; onNavigate: (id: string) => void } = $props()

  const rate = $derived(store.monitor.rateMs > 0 ? `${store.monitor.rateMs} 毫秒` : '未运行')
</script>

<nav class="flex h-full w-[188px] shrink-0 flex-col border-r border-border/60 bg-sidebar/30 px-3 py-4">
  <div class="px-2">
    <p class="text-[11px] text-muted-foreground">
      {store.info ? `${store.info.controller} · ${store.info.hramOffset}` : 'N155A'}
    </p>
  </div>

  <div class="mt-6 flex flex-col gap-0.5">
    {#each routes as item (item.id)}
      <button
        type="button"
        class={cn(
          'flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm transition-colors',
          current === item.id
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
        )}
        onclick={() => onNavigate(item.id)}
      >
        <item.icon class="size-4" />
        {item.label}
      </button>
    {/each}
  </div>

  <div class="mt-auto rounded-lg border border-border/50 px-2.5 py-2">
    <div class="flex items-center gap-1.5">
      <span
        class={cn(
          'size-1.5 rounded-full',
          store.monitor.running ? 'bg-emerald-400' : 'bg-muted-foreground/50'
        )}
      ></span>
      <span class="text-[11px] text-muted-foreground">
        {store.monitor.running ? '实时采集中' : '采集已停止'}
      </span>
    </div>
    <p class="mt-1 text-[11px] text-muted-foreground/70">间隔 {rate}</p>
  </div>
</nav>
