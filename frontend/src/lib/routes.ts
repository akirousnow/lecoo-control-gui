import { BatteryCharging, Fan, Info, LayoutDashboard, Zap, type LucideIcon } from '@lucide/svelte'
import type { Component } from 'svelte'

import About from './pages/About.svelte'
import Charge from './pages/Charge.svelte'
import Dashboard from './pages/Dashboard.svelte'
import Fans from './pages/Fans.svelte'
import Power from './pages/Power.svelte'

export interface Route {
  id: string
  label: string
  icon: LucideIcon
  component: Component
}

export const routes: Route[] = [
  { id: 'overview', label: '总览', icon: LayoutDashboard, component: Dashboard },
  { id: 'fans', label: '风扇', icon: Fan, component: Fans },
  { id: 'power', label: '电源', icon: Zap, component: Power },
  { id: 'charge', label: '充电', icon: BatteryCharging, component: Charge },
  { id: 'about', label: '关于', icon: Info, component: About }
]
