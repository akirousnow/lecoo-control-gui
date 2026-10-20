import { api } from './api'

export type Theme = 'light' | 'dark'

class ThemeStore {
  current = $state<Theme>('dark')

  init(): void {
    const stored = localStorage.getItem('lecoo-theme')
    this.apply(stored === 'light' || stored === 'dark' ? stored : 'dark')
  }

  toggle(): void {
    this.apply(this.current === 'dark' ? 'light' : 'dark')
  }

  private apply(theme: Theme): void {
    this.current = theme
    document.documentElement.classList.toggle('dark', theme === 'dark')
    localStorage.setItem('lecoo-theme', theme)
    void api.setTheme(theme).catch(() => undefined)
  }
}

export const theme = new ThemeStore()
