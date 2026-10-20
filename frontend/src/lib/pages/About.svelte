<script lang="ts">
  import { onMount } from 'svelte'
  import { Button } from '$lib/components/ui/button'
  import { Switch } from '$lib/components/ui/switch'
  import { api, failureMessage } from '$lib/api'
  import { rise } from '$lib/motion'
  import { store } from '$lib/store.svelte'

  let raw = $state('')
  let rawError = $state('')
  let confirmText = $state('')
  let cliPath = $state('')

  const telemetry = $derived(store.settings?.telemetryEnabled ?? false)

  const rows = $derived([
    { label: '控制器', value: store.info?.controller ?? '—' },
    {
      label: '控制器版本',
      value: store.info ? String(store.info.controllerVersion).padStart(2, '0') : '—'
    },
    { label: 'HRAM 偏移', value: store.info?.hramOffset ?? '—' },
    { label: '守护进程', value: store.info?.daemonVersion ?? '—' },
    { label: 'CLI 版本', value: store.info?.cliVersion ?? '—' }
  ])

  onMount(() => {
    cliPath = store.health?.cliPath ?? ''
    void loadRaw()
  })

  async function loadRaw(): Promise<void> {
    rawError = ''
    try {
      raw = await api.rawSettings()
    } catch (error) {
      rawError = failureMessage(error)
    }
  }

  async function toggleTelemetry(enabled: boolean): Promise<void> {
    await store.setTelemetry(enabled)
    await loadRaw()
  }

  async function copyId(): Promise<void> {
    const id = store.settings?.telemetryClientId
    if (!id) {
      return
    }
    await navigator.clipboard.writeText(id)
    store.notice = '遥测 ID 已复制'
  }

  async function reset(): Promise<void> {
    if (confirmText !== 'RESET') {
      return
    }
    await store.reset()
    confirmText = ''
    await loadRaw()
  }

  async function savePath(): Promise<void> {
    await store.updateCliPath(cliPath)
    await loadRaw()
  }

  async function openHwtest(): Promise<void> {
    try {
      await api.openHwtest()
    } catch (error) {
      store.error = failureMessage(error)
    }
  }
</script>

<div class="flex flex-col gap-5">
  <header>
    <h1 class="text-lg font-semibold">关于与高级</h1>
    <p class="mt-0.5 text-xs text-muted-foreground">设备信息、遥测与原始状态</p>
  </header>

  <section use:rise class="rounded-xl border border-border/60 bg-card/50 p-4">
    <dl class="grid gap-2 sm:grid-cols-2">
      {#each rows as row (row.label)}
        <div class="flex items-center justify-between gap-3 border-b border-border/30 pb-1.5">
          <dt class="text-[11px] text-muted-foreground">{row.label}</dt>
          <dd class="font-mono text-xs">{row.value}</dd>
        </div>
      {/each}
    </dl>
  </section>

  <section use:rise={{ delay: 0.05 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
    <div class="flex items-center justify-between gap-4">
      <div>
        <p class="text-xs font-medium">匿名遥测</p>
        <p class="mt-1 text-[11px] text-muted-foreground">
          守护进程默认开启，会上报机器型号、传感器读数与错误日志。关闭不影响其他功能。
        </p>
      </div>
      <Switch checked={telemetry} disabled={store.busy} onCheckedChange={toggleTelemetry} />
    </div>
    <div class="mt-3 flex items-center justify-between gap-3 rounded-lg bg-muted/40 px-2.5 py-2">
      <span class="font-mono text-[11px] text-muted-foreground">
        {store.settings?.telemetryClientId ?? '—'}
      </span>
      <Button size="xs" variant="ghost" onclick={copyId}>复制</Button>
    </div>
  </section>

  <section use:rise={{ delay: 0.1 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
    <p class="text-xs font-medium">CLI 路径</p>
    <div class="mt-2 flex gap-2">
      <input
        bind:value={cliPath}
        class="h-8 flex-1 rounded-lg border border-border/60 bg-background px-2.5 font-mono text-[11px] outline-none focus:border-ring"
        spellcheck="false"
      />
      <Button size="sm" variant="outline" onclick={savePath}>保存</Button>
    </div>
    <p class="mt-2 text-[11px] text-muted-foreground">
      修改后会重新探测守护进程并刷新全部状态。
    </p>
  </section>

  <section use:rise={{ delay: 0.15 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
    <p class="text-xs font-medium">操作</p>
    <div class="mt-3 flex flex-wrap gap-2">
      <Button size="sm" variant="outline" disabled={store.busy} onclick={() => void store.reapply()}>
        重新下发设置
      </Button>
      <Button size="sm" variant="outline" onclick={openHwtest}>打开硬件自检</Button>
      <Button size="sm" variant="ghost" onclick={loadRaw}>刷新原始状态</Button>
    </div>
    <p class="mt-2 text-[11px] text-muted-foreground">
      硬件自检是交互式向导，会在独立的控制台窗口里运行。
    </p>
  </section>

  <section use:rise={{ delay: 0.2 }} class="rounded-xl border border-destructive/30 bg-destructive/5 p-4">
    <p class="text-xs font-medium text-destructive">恢复守护进程默认</p>
    <p class="mt-1 text-[11px] text-muted-foreground">
      会把守护进程的全部设置重置为出厂值，包括电源模式、风扇与充电限制。键入 RESET 以启用。
    </p>
    <div class="mt-3 flex gap-2">
      <input
        bind:value={confirmText}
        placeholder="RESET"
        class="h-8 w-32 rounded-lg border border-border/60 bg-background px-2.5 font-mono text-[11px] outline-none focus:border-destructive/60"
        spellcheck="false"
      />
      <Button
        size="sm"
        variant="destructive"
        disabled={confirmText !== 'RESET' || store.busy}
        onclick={reset}
      >
        恢复默认
      </Button>
    </div>
  </section>

  <section use:rise={{ delay: 0.25 }} class="rounded-xl border border-border/60 bg-card/50 p-4">
    <div class="flex items-center justify-between">
      <p class="text-xs font-medium">原始状态</p>
      <span class="text-[11px] text-muted-foreground">daemon settings read</span>
    </div>
    {#if rawError}
      <p class="mt-2 rounded-md bg-destructive/10 px-2.5 py-2 text-[11px] text-destructive">
        {rawError}
      </p>
    {:else}
      <pre class="mt-2 max-h-56 overflow-auto rounded-lg bg-muted/40 p-3 font-mono text-[11px] leading-relaxed text-muted-foreground">{raw || '读取中'}</pre>
    {/if}
  </section>
</div>
