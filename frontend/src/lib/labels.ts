import type { ChargeTarget } from './domain'

const powerLabels: Record<string, string> = {
  silent: '静音',
  default: '默认',
  perf: '性能'
}

const fanModeLabels: Record<string, string> = {
  auto: '自动',
  full: '全速',
  custom: '自定义',
  turbo: '涡轮'
}

const fanTargetLabels: Record<string, string> = {
  cpu: 'CPU',
  gpu: 'GPU',
  both: '同步'
}

const chargePresetLabels: Record<string, string> = {
  full: '全容量',
  desk: '桌面',
  lifespan: '最大寿命',
  balanced: '平衡',
  high: '高容量'
}

const chargePresetPercents: Record<string, number> = {
  full: 100,
  desk: 40,
  lifespan: 60,
  balanced: 80,
  high: 95
}

export const powerProfiles = [
  { id: 'silent', label: '静音', hint: '最低功耗，风扇更安静' },
  { id: 'default', label: '默认', hint: '平衡功耗与噪音' },
  { id: 'perf', label: '性能', hint: '最高功耗，风扇更早升速' }
] as const

export const chargePresets = [
  { id: 'full', label: '全容量', percent: 100, hint: '每次充满，续航最长' },
  { id: 'high', label: '高容量', percent: 95, hint: '接近充满，损耗略高' },
  { id: 'balanced', label: '平衡', percent: 80, hint: '日常使用的折中' },
  { id: 'lifespan', label: '最大寿命', percent: 60, hint: '电池寿命优先' },
  { id: 'desk', label: '桌面', percent: 40, hint: '长期插电时推荐' }
] as const

export const fanTargets = [
  { id: 'both', label: '同步' },
  { id: 'cpu', label: 'CPU' },
  { id: 'gpu', label: 'GPU' }
] as const

export const fanModes = [
  { id: 'auto', label: '自动', hint: '交给 EC 控制' },
  { id: 'full', label: '全速', hint: '100% 占空比，噪音明显' },
  { id: 'custom', label: '自定义', hint: '手动设定 PWM' }
] as const

export function powerLabel(kind: string): string {
  return powerLabels[kind] ?? '未知'
}

export function fanModeLabel(kind: string): string {
  return fanModeLabels[kind] ?? '未知'
}

export function fanTargetLabel(target: string): string {
  return fanTargetLabels[target] ?? target
}

export function chargePresetLabel(preset: string): string {
  return chargePresetLabels[preset] ?? preset
}

export function chargePresetPercent(preset: string): number | null {
  return chargePresetPercents[preset] ?? null
}

export function describeCharge(target: ChargeTarget | null): string {
  if (!target) {
    return '读取中'
  }
  if (target.kind === 'preset') {
    const percent = chargePresetPercent(target.preset)
    const label = chargePresetLabel(target.preset)
    return percent == null ? label : `${label} ${percent}%`
  }
  if (target.kind === 'preserve') {
    if (target.min != null && target.max != null) {
      return `自定义 ${target.min}% .. ${target.max}%`
    }
    return '自定义区间'
  }
  return target.raw || '未知'
}

export function describeFan(mode: { kind: string; pwm: number } | null): string {
  if (!mode) {
    return '读取中'
  }
  if (mode.kind === 'custom') {
    return `自定义 ${mode.pwm}/255`
  }
  return fanModeLabel(mode.kind)
}
