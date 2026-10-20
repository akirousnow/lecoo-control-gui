<script lang="ts">
  import { Check } from '@lucide/svelte'
  import { powerProfiles } from '$lib/labels'
  import { rise } from '$lib/motion'
  import { store } from '$lib/store.svelte'
  import { cn } from '$lib/utils'

  const current = $derived(store.settings?.power.kind ?? '')

  async function choose(profile: string) {
    if (profile === current) {
      return
    }
    await store.setPower(profile)
  }
</script>

<div class="flex flex-col gap-5">
  <header>
    <h1 class="text-lg font-semibold">电源</h1>
    <p class="mt-0.5 text-xs text-muted-foreground">切换即生效，完成后自动读回确认</p>
  </header>

  <div class="grid gap-3 md:grid-cols-3">
    {#each powerProfiles as profile, index (profile.id)}
      <button
        type="button"
        use:rise={{ delay: index * 0.05 }}
        class={cn(
          'rounded-xl border p-4 text-left transition-colors',
          current === profile.id
            ? 'border-primary/60 bg-primary/10'
            : 'border-border/60 bg-card/50 hover:border-border hover:bg-card/80'
        )}
        disabled={store.busy}
        onclick={() => choose(profile.id)}
      >
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">{profile.label}</span>
          {#if current === profile.id}
            <Check class="size-4 text-primary" />
          {/if}
        </div>
        <p class="mt-2 text-[11px] leading-relaxed text-muted-foreground">{profile.hint}</p>
      </button>
    {/each}
  </div>

  <p class="text-[11px] text-muted-foreground">
    性能模式与风扇联动。选择自动风扇时，EC 会更早提高转速，噪音随之上升。
  </p>
</div>
