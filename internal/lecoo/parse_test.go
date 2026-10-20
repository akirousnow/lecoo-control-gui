package lecoo

import "testing"

const settingsPlain = `CurrentSettings {
    telemetry_enabled: true,
    telemetry_client_id: 1003880566478677932,
    keyboard_backlight: Medium,
    led_mode: Auto,
    power_profile: Default,
    charge: Full,
    fan_mode_cpu: Auto,
    fan_mode_gpu: Auto,
}
`

const settingsCustom = `CurrentSettings {
    telemetry_enabled: false,
    telemetry_client_id: 1003880566478677932,
    keyboard_backlight: Custom(
        50,
    ),
    led_mode: Auto,
    power_profile: Silent,
    charge: Preserve(
        Some(
            ChargeRange {
                min: 40,
                max: 80,
            },
        ),
    ),
    fan_mode_cpu: Custom(
        120,
    ),
    fan_mode_gpu: Auto,
}
`

func TestParseSettingsPlain(t *testing.T) {
	got, err := ParseSettings(settingsPlain)
	if err != nil {
		t.Fatalf("ParseSettings: %v", err)
	}
	if !got.TelemetryEnabled {
		t.Error("telemetry should be enabled")
	}
	if got.TelemetryClientID != "1003880566478677932" {
		t.Errorf("client id = %q", got.TelemetryClientID)
	}
	if got.Power.Kind != "default" {
		t.Errorf("power = %q, want default", got.Power.Kind)
	}
	if got.Charge.Kind != "preset" || got.Charge.Preset != "full" {
		t.Errorf("charge = %+v, want preset full", got.Charge)
	}
	if got.FanCPU.Kind != "auto" || got.FanGPU.Kind != "auto" {
		t.Errorf("fans = %v/%v, want auto/auto", got.FanCPU.Kind, got.FanGPU.Kind)
	}
}

func TestParseSettingsCustom(t *testing.T) {
	got, err := ParseSettings(settingsCustom)
	if err != nil {
		t.Fatalf("ParseSettings: %v", err)
	}
	if got.TelemetryEnabled {
		t.Error("telemetry should be disabled")
	}
	if got.Power.Kind != "silent" {
		t.Errorf("power = %q, want silent", got.Power.Kind)
	}
	if got.Charge.Kind != "preserve" {
		t.Fatalf("charge kind = %q, want preserve", got.Charge.Kind)
	}
	if got.Charge.Min == nil || *got.Charge.Min != 40 {
		t.Errorf("charge min = %v, want 40", got.Charge.Min)
	}
	if got.Charge.Max == nil || *got.Charge.Max != 80 {
		t.Errorf("charge max = %v, want 80", got.Charge.Max)
	}
	if got.FanCPU.Kind != "custom" || got.FanCPU.PWM != 120 {
		t.Errorf("cpu fan = %+v, want custom 120", got.FanCPU)
	}
	if got.FanGPU.Kind != "auto" {
		t.Errorf("gpu fan = %q, want auto", got.FanGPU.Kind)
	}
}

func TestParseSettingsRejectsForeignOutput(t *testing.T) {
	if _, err := ParseSettings("⚡ 电源配置:\n   当前: Silent\n"); err == nil {
		t.Fatal("expected an error for non-settings output")
	}
}

func TestParseSampleMonitoringLine(t *testing.T) {
	line := "🌡️ CPU: 91°C, 系统: 56°C | 💨 风扇: 3809 RPM (CPU), 4312 RPM (GPU)"
	got := ParseSample(line)
	if got.CPUTempC == nil || *got.CPUTempC != 91 {
		t.Errorf("cpu temp = %v", got.CPUTempC)
	}
	if got.SysTempC == nil || *got.SysTempC != 56 {
		t.Errorf("sys temp = %v", got.SysTempC)
	}
	if got.CPURpm == nil || *got.CPURpm != 3809 {
		t.Errorf("cpu rpm = %v", got.CPURpm)
	}
	if got.GPURpm == nil || *got.GPURpm != 4312 {
		t.Errorf("gpu rpm = %v", got.GPURpm)
	}
}

func TestParseSampleTempsOnly(t *testing.T) {
	got := ParseSample("🌡️ 温度:\n   CPU: 79 °C\n   系统: 56 °C\n")
	if got.CPURpm != nil || got.GPURpm != nil {
		t.Error("temps output has no fan speeds")
	}
	if got.CPUTempC == nil || *got.CPUTempC != 79 {
		t.Errorf("cpu temp = %v", got.CPUTempC)
	}
}

func TestParseBatteryPreset(t *testing.T) {
	raw := "🔋 充电状态:\n   目标:   满容量\n   生效:   满容量\n   电池:   69%\n"
	got, err := ParseBattery(raw)
	if err != nil {
		t.Fatalf("ParseBattery: %v", err)
	}
	if got.Target != "满容量" || got.Effective != "满容量" {
		t.Errorf("target/effective = %q/%q", got.Target, got.Effective)
	}
	if got.Percent == nil || *got.Percent != 69 {
		t.Errorf("percent = %v, want 69", got.Percent)
	}
}

func TestParseBatteryPreserve(t *testing.T) {
	raw := "🔋 充电状态:\n   目标:   保护模式 (40% .. 80%)\n   生效:   保护模式 (40% .. 80%)\n   电池:   23%\n"
	got, err := ParseBattery(raw)
	if err != nil {
		t.Fatalf("ParseBattery: %v", err)
	}
	if got.Target != "保护模式 (40% .. 80%)" {
		t.Errorf("target = %q", got.Target)
	}
	if got.Percent == nil || *got.Percent != 23 {
		t.Errorf("percent = %v, want 23", got.Percent)
	}
}

func TestParseInfo(t *testing.T) {
	raw := "控制器: IT5571 (版本 07)\nHRAM偏移: 0x0400\n守护进程版本: 0.5.2-beta\n"
	got := ParseInfo(raw)
	if got.Controller != "IT5571" {
		t.Errorf("controller = %q", got.Controller)
	}
	if got.ControllerVersion != 7 {
		t.Errorf("version = %d, want 7", got.ControllerVersion)
	}
	if got.HramOffset != "0x0400" {
		t.Errorf("hram = %q", got.HramOffset)
	}
	if got.DaemonVersion != "0.5.2-beta" {
		t.Errorf("daemon = %q", got.DaemonVersion)
	}
}
