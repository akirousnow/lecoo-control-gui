export type FanMode = {
  kind: string
  pwm: number
}

export type PowerProfile = {
  kind: string
  raw: string
}

export type ChargeTarget = {
  kind: string
  preset: string
  min?: number
  max?: number
  raw: string
}

export type Settings = {
  telemetryEnabled: boolean
  telemetryClientId: string
  power: PowerProfile
  charge: ChargeTarget
  fanCpu: FanMode
  fanGpu: FanMode
  readAt: number
}

export type Sample = {
  at: number
  cpuTempC?: number
  sysTempC?: number
  cpuRpm?: number
  gpuRpm?: number
}

export type Battery = {
  target: string
  effective: string
  percent?: number
  at: number
}

export type Info = {
  controller: string
  controllerVersion: number
  hramOffset: string
  daemonVersion: string
  cliVersion: string
}

export type MonitorState = {
  running: boolean
  rateMs: number
  error: string
}

export type Health = {
  cliPath: string
  cliFound: boolean
  daemonUp: boolean
  version: string
  message: string
}

export type FanTarget = 'cpu' | 'gpu' | 'both'
