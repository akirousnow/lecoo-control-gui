import * as backend from '../../wailsjs/go/main/App'
import { lecoo } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type {
  Battery,
  Capabilities,
  CurveState,
  FanCurve,
  FanTarget,
  Health,
  Info,
  MonitorState,
  Sample,
  Settings,
  WidgetPosition
} from './domain'

export const EVENTS = {
  sample: 'lecoo:sample',
  settings: 'lecoo:settings',
  battery: 'lecoo:battery',
  monitor: 'lecoo:monitor',
  curve: 'lecoo:curve'
} as const

export function failureMessage(error: unknown): string {
  if (typeof error === 'string') {
    return error
  }
  if (error instanceof Error) {
    return error.message
  }
  return '命令执行失败'
}

export function onEvent<T>(name: string, handler: (payload: T) => void): () => void {
  return EventsOn(name, (payload: T) => handler(payload))
}

export const api = {
  health: () => backend.Health() as Promise<Health>,
  info: () => backend.Info() as Promise<Info>,
  settings: () => backend.Settings() as Promise<Settings>,
  battery: () => backend.Battery() as Promise<Battery>,
  sample: () => backend.Sample() as Promise<Sample>,
  rawSettings: () => backend.RawSettings() as Promise<string>,
  setPower: (profile: string) => backend.SetPower(profile) as Promise<Settings>,
  setFan: (target: FanTarget, mode: string, pwm: number) =>
    backend.SetFan(target, mode, pwm) as Promise<Settings>,
  setChargePreset: (preset: string) => backend.SetChargePreset(preset) as Promise<Settings>,
  setChargeRange: (min: number, max: number) =>
    backend.SetChargeRange(min, max) as Promise<Settings>,
  setTelemetry: (enabled: boolean) => backend.SetTelemetry(enabled) as Promise<Settings>,
  reapply: () => backend.Reapply() as Promise<Settings>,
  reset: () => backend.Reset() as Promise<Settings>,
  cliPath: () => backend.CliPath() as Promise<string>,
  setCliPath: (path: string) => backend.SetCliPath(path) as Promise<string>,
  openHwtest: () => backend.OpenHwtest() as Promise<void>,
  startMonitor: (rateMs: number) => backend.StartMonitor(rateMs) as Promise<MonitorState>,
  monitorState: () => backend.MonitorState() as Promise<MonitorState>,
  curveState: () => backend.CurveState() as Promise<CurveState>,
  capabilities: () => backend.Capabilities() as Promise<Capabilities>,
  setCurve: (curve: FanCurve) =>
    backend.SetCurve(lecoo.FanCurve.createFrom(curve)) as Promise<CurveState>,
  minimise: () => backend.Minimise() as Promise<void>,
  toggleMaximise: () => backend.ToggleMaximise() as Promise<void>,
  isMaximised: () => backend.IsMaximised() as Promise<boolean>,
  quit: () => backend.Quit() as Promise<void>,
  setTheme: (theme: string) => backend.SetTheme(theme) as Promise<void>,
  setCompactWindow: (compact: boolean, alwaysOnTop: boolean) =>
    backend.SetCompactWindow(compact, alwaysOnTop) as Promise<void>,
  setAlwaysOnTop: (enabled: boolean) => backend.SetAlwaysOnTop(enabled) as Promise<void>,
  saveWidgetPosition: () => backend.SaveWidgetPosition() as Promise<WidgetPosition>,
  stopMonitor: () => backend.StopMonitor() as Promise<MonitorState>
}
