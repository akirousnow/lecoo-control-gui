import { animate } from 'motion'

const easeOut = [0.16, 1, 0.3, 1] as const

export function rise(node: HTMLElement, params: { delay?: number; distance?: number } = {}) {
  const { delay = 0, distance = 8 } = params
  animate(node, { opacity: [0, 1], y: [distance, 0] }, { duration: 0.34, delay, ease: easeOut })
  return {}
}

export function fadeIn(node: HTMLElement) {
  animate(node, { opacity: [0.4, 1] }, { duration: 0.18 })
  return {}
}

export function pulseOnce(node: HTMLElement) {
  animate(node, { scale: [0.96, 1] }, { duration: 0.22, ease: easeOut })
  return {}
}
