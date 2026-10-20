import { api, EVENTS, failureMessage, onEvent } from './api'
import type { Battery, FanTarget, Health, Info, MonitorState, Sample, Settings } from './domain'

export type HistoryPoint = {
  t: number
  cpu: number | null
  sys: number | null
  cpuRpm: number | null
  gpuRpm: number | null
}

const MAX_HISTORY = 120
const RECONCILE_INTERVAL = 30000

class Store {
  settings = $state<Settings | null>(null)
  sample = $state<Sample | null>(null)
  battery = $state<Battery | null>(null)
  info = $state<Info | null>(null)
  health = $state<Health | null>(null)
  monitor = $state<MonitorState>({ running: false, rateMs: 0, error: '' })
  history = $state<HistoryPoint[]>([])
  lastSampleAt = $state(0)
  error = $state('')
  notice = $state('')
  busy = $state(false)
  stale = $state(false)

  connect(): () => void {
    const offSample = onEvent<Sample>(EVENTS.sample, (sample) => this.applySample(sample))
    const offSettings = onEvent<Settings>(EVENTS.settings, (settings) => {
      this.settings = settings
    })
    const offBattery = onEvent<Battery>(EVENTS.battery, (battery) => {
      this.battery = battery
    })
    const offMonitor = onEvent<MonitorState>(EVENTS.monitor, (state) => this.applyMonitor(state))

    const reconcile = setInterval(() => {
      void this.refreshSettings()
      void this.refreshBattery()
    }, RECONCILE_INTERVAL)

    return () => {
      offSample()
      offSettings()
      offBattery()
      offMonitor()
      clearInterval(reconcile)
    }
  }

  async init(): Promise<void> {
    this.error = ''
    try {
      const health = await api.health()
      this.health = health
      if (!health.cliFound) {
        this.error = `找不到 ${health.cliPath}`
        return
      }
      if (!health.daemonUp) {
        this.error = health.message || '守护进程未响应'
        return
      }
      const [settings, info, battery, monitor] = await Promise.all([
        api.settings(),
        api.info(),
        api.battery(),
        api.monitorState()
      ])
      this.settings = settings
      this.info = info
      this.battery = battery
      this.monitor = monitor
    } catch (error) {
      this.error = failureMessage(error)
    }
  }

  async refreshSettings(): Promise<void> {
    try {
      this.settings = await api.settings()
    } catch (error) {
      this.error = failureMessage(error)
    }
  }

  async refreshBattery(): Promise<void> {
    try {
      this.battery = await api.battery()
    } catch (error) {
      this.error = failureMessage(error)
    }
  }

  async retry(): Promise<void> {
    await this.init()
  }

  async setPower(profile: string): Promise<void> {
    await this.mutate(() => api.setPower(profile), '电源模式已切换')
  }

  async setFan(target: FanTarget, mode: string, pwm: number): Promise<void> {
    await this.mutate(() => api.setFan(target, mode, pwm), '风扇设置已下发')
  }

  async setChargePreset(preset: string): Promise<void> {
    await this.mutate(() => api.setChargePreset(preset), '充电限制已更新')
  }

  async setChargeRange(min: number, max: number): Promise<void> {
    await this.mutate(() => api.setChargeRange(min, max), '充电区间已更新')
  }

  async setTelemetry(enabled: boolean): Promise<void> {
    await this.mutate(() => api.setTelemetry(enabled), enabled ? '遥测已开启' : '遥测已关闭')
  }

  async reapply(): Promise<void> {
    await this.mutate(() => api.reapply(), '设置已重新下发')
  }

  async reset(): Promise<void> {
    await this.mutate(() => api.reset(), '守护进程设置已恢复默认')
  }

  async updateCliPath(path: string): Promise<void> {
    try {
      const applied = await api.setCliPath(path)
      this.health = this.health ? { ...this.health, cliPath: applied } : null
      this.notice = 'CLI 路径已更新'
      await this.init()
    } catch (error) {
      this.error = failureMessage(error)
    }
  }

  private async mutate(action: () => Promise<Settings>, message: string): Promise<void> {
    this.busy = true
    this.error = ''
    try {
      this.settings = await action()
      this.notice = message
      setTimeout(() => {
        if (this.notice === message) {
          this.notice = ''
        }
      }, 2600)
    } catch (error) {
      this.error = failureMessage(error)
    } finally {
      this.busy = false
    }
  }

  private applySample(sample: Sample): void {
    this.sample = sample
    this.lastSampleAt = sample.at
    this.stale = false
    const point: HistoryPoint = {
      t: sample.at,
      cpu: sample.cpuTempC ?? null,
      sys: sample.sysTempC ?? null,
      cpuRpm: sample.cpuRpm ?? null,
      gpuRpm: sample.gpuRpm ?? null
    }
    const next = this.history.concat(point)
    this.history = next.length > MAX_HISTORY ? next.slice(next.length - MAX_HISTORY) : next
  }

  private applyMonitor(state: MonitorState): void {
    this.monitor = { running: state.running, rateMs: state.rateMs, error: state.error ?? '' }
    if (state.error) {
      this.stale = true
      this.error = state.error
    }
  }
}

export const store = new Store()
