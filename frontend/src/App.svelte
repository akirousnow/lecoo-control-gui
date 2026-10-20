<script lang="ts">
  import { onMount } from 'svelte'
  import CompactWidget from '$lib/components/CompactWidget.svelte'
  import Sidebar from '$lib/components/Sidebar.svelte'
  import Titlebar from '$lib/components/Titlebar.svelte'
  import { Button } from '$lib/components/ui/button'
  import About from '$lib/pages/About.svelte'
  import Charge from '$lib/pages/Charge.svelte'
  import Dashboard from '$lib/pages/Dashboard.svelte'
  import Fans from '$lib/pages/Fans.svelte'
  import Power from '$lib/pages/Power.svelte'
  import { store } from '$lib/store.svelte'
  import { theme } from '$lib/theme.svelte'
  import { view } from '$lib/view.svelte'

  let page = $state('overview')

  onMount(() => {
    theme.init()
    view.init()
    const detach = store.connect()
    void store.init()
    return detach
  })
</script>

{#if view.compact}
  <CompactWidget />
{:else}
  <div class="flex h-screen w-screen flex-col overflow-hidden bg-background text-foreground">
    <Titlebar />

    <div class="flex min-h-0 flex-1 overflow-hidden">
      <Sidebar current={page} onNavigate={(id) => (page = id)} />

      <main class="min-w-0 flex-1 overflow-y-auto overflow-x-hidden">
        <div class="mx-auto w-full max-w-5xl px-6 py-6">
          {#if store.error}
            <div
              class="mb-4 flex items-start justify-between gap-3 rounded-xl border border-destructive/40 bg-destructive/10 px-4 py-3"
            >
              <div class="min-w-0">
                <p class="text-xs font-medium text-destructive">命令失败</p>
                <p class="mt-1 break-words text-[11px] leading-relaxed text-destructive/90">
                  {store.error}
                </p>
              </div>
              <Button size="xs" variant="outline" onclick={() => void store.retry()}>重试</Button>
            </div>
          {/if}

          {#if store.notice}
            <div
              class="mb-4 rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-4 py-2.5 text-[11px] text-emerald-700 dark:text-emerald-300"
            >
              {store.notice}
            </div>
          {/if}

          {#if page === 'fans'}
            <Fans />
          {:else if page === 'power'}
            <Power />
          {:else if page === 'charge'}
            <Charge />
          {:else if page === 'about'}
            <About />
          {:else}
            <Dashboard />
          {/if}
        </div>
      </main>
    </div>
  </div>
{/if}
