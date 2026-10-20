import { api } from './api'

class ViewStore {
  compact = $state(false)
  pinned = $state(true)

  init(): void {
    this.compact = localStorage.getItem('lecoo-compact') === '1'
    this.pinned = localStorage.getItem('lecoo-pinned') !== '0'
    if (this.compact) {
      void api.setCompactWindow(true, this.pinned).catch(() => undefined)
    }
  }

  toggleCompact(): void {
    this.compact = !this.compact
    localStorage.setItem('lecoo-compact', this.compact ? '1' : '0')
    void api.setCompactWindow(this.compact, this.pinned).catch(() => undefined)
  }

  togglePinned(): void {
    this.pinned = !this.pinned
    localStorage.setItem('lecoo-pinned', this.pinned ? '1' : '0')
    void api.setAlwaysOnTop(this.pinned).catch(() => undefined)
  }
}

export const view = new ViewStore()
