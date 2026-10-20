<script lang="ts">
  let { values = [], className = '' }: { values?: number[]; className?: string } = $props()

  const path = $derived.by(() => {
    const data = values.filter((value) => Number.isFinite(value))
    if (data.length < 2) {
      return ''
    }
    const low = Math.min(...data)
    let high = Math.max(...data)
    if (high - low < 2) {
      high = low + 2
    }
    const width = 100
    const height = 28
    return data
      .map((value, index) => {
        const x = (index / (data.length - 1)) * width
        const y = height - ((value - low) / (high - low)) * height
        return `${index === 0 ? 'M' : 'L'}${x.toFixed(2)},${y.toFixed(2)}`
      })
      .join(' ')
  })
</script>

<svg class={className} viewBox="0 0 100 28" preserveAspectRatio="none" aria-hidden="true">
  <path
    d={path}
    fill="none"
    stroke="currentColor"
    stroke-width="1.5"
    stroke-linecap="round"
    stroke-linejoin="round"
    vector-effect="non-scaling-stroke"
  />
</svg>
