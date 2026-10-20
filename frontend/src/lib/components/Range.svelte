<script lang="ts">
  let {
    value = $bindable(0),
    min = 0,
    max = 255,
    step = 1,
    disabled = false
  }: {
    value?: number
    min?: number
    max?: number
    step?: number
    disabled?: boolean
  } = $props()
</script>

<input
  type="range"
  {min}
  {max}
  {step}
  {disabled}
  bind:value
  class="h-1.5 w-full cursor-pointer appearance-none rounded-full bg-muted accent-primary disabled:cursor-not-allowed disabled:opacity-50"
/>
