<script lang="ts">
  import { Copy, Minus, Moon, PictureInPicture2, Square, Sun, X } from '@lucide/svelte'
  import appIcon from '../../assets/app-icon.png'
  import { onMount } from 'svelte'
  import { api } from '$lib/api'
  import { theme } from '$lib/theme.svelte'
  import { view } from '$lib/view.svelte'
  import { cn } from '$lib/utils'

  let maximised = $state(false)

  async function sync(): Promise<void> {
    try {
      maximised = await api.isMaximised()
    } catch {
      maximised = false
    }
  }

  async function toggleMaximise(): Promise<void> {
    await api.toggleMaximise()
    await sync()
  }

  onMount(() => {
    void sync()
    const handler = () => void sync()
    window.addEventListener('resize', handler)
    return () => window.removeEventListener('resize', handler)
  })
</script>

<header
  class="flex h-10 shrink-0 select-none items-center justify-between border-b border-border/60 bg-background pl-4"
  style="--wails-draggable:drag"
>
  <div class="flex items-center gap-2">
    <img src={appIcon} alt="" class="size-4 rounded-[4px]" />
    <span class="text-xs font-medium tracking-wide">来酷控制中心</span>
  </div>

  <div class="flex h-full items-stretch" style="--wails-draggable:no-drag">
    <button
      type="button"
      class="flex w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
      aria-label="精简模式"
      title="精简模式，可当作桌面小组件"
      onclick={() => view.toggleCompact()}
    >
      <PictureInPicture2 class="size-3.5" />
    </button>
    <button
      type="button"
      class="flex w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
      aria-label="切换主题"
      title={theme.current === 'dark' ? '切换到白天模式' : '切换到夜间模式'}
      onclick={() => theme.toggle()}
    >
      {#if theme.current === 'dark'}
        <Sun class="size-3.5" />
      {:else}
        <Moon class="size-3.5" />
      {/if}
    </button>
    <button
      type="button"
      class="flex w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
      aria-label="最小化"
      onclick={() => api.minimise()}
    >
      <Minus class="size-3.5" />
    </button>
    <button
      type="button"
      class="flex w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
      aria-label={maximised ? '还原' : '最大化'}
      onclick={toggleMaximise}
    >
      {#if maximised}
        <Copy class="size-3.5" />
      {:else}
        <Square class="size-3" />
      {/if}
    </button>
    <button
      type="button"
      class={cn(
        'flex w-11 items-center justify-center text-muted-foreground transition-colors',
        'hover:bg-destructive hover:text-white'
      )}
      aria-label="关闭"
      onclick={() => api.quit()}
    >
      <X class="size-3.5" />
    </button>
  </div>
</header>
