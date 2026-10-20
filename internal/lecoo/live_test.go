package lecoo

import (
	"context"
	"os"
	"testing"
	"time"
)

func liveOnly(t *testing.T) {
	t.Helper()
	if os.Getenv("LECOO_LIVE") != "1" {
		t.Skip("set LECOO_LIVE=1 to run against the real lecoo-ctrl.exe")
	}
}

func TestLiveReads(t *testing.T) {
	liveOnly(t)
	runner := NewRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	settings, err := runner.Settings(ctx)
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if settings.Power.Kind == "unknown" {
		t.Errorf("power profile not recognised: %+v", settings.Power)
	}
	if settings.FanCPU.Kind == "unknown" {
		t.Errorf("cpu fan mode not recognised: %+v", settings.FanCPU)
	}
	t.Logf("settings: power=%s charge=%+v fan=%s/%s telemetry=%v",
		settings.Power.Kind, settings.Charge, settings.FanCPU.Kind, settings.FanGPU.Kind, settings.TelemetryEnabled)

	info, err := runner.Info(ctx)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.Controller == "" || info.HramOffset == "" {
		t.Errorf("info incomplete: %+v", info)
	}
	t.Logf("info: %+v", info)

	battery, err := runner.Battery(ctx)
	if err != nil {
		t.Fatalf("Battery: %v", err)
	}
	if battery.Percent == nil {
		t.Errorf("battery percent missing: %+v", battery)
	}
	t.Logf("battery: %+v", battery)

	sample, err := runner.Sample(ctx)
	if err != nil {
		t.Fatalf("Sample: %v", err)
	}
	if sample.CPUTempC == nil || sample.CPURpm == nil {
		t.Errorf("sample incomplete: %+v", sample)
	}
	t.Logf("sample: %+v", sample)
}

func TestLiveMonitor(t *testing.T) {
	liveOnly(t)
	service := NewService()
	defer service.Close()

	samples := make(chan Sample, 16)
	failures := make(chan string, 4)
	service.SetEmitter(func(name string, payload any) {
		if name == EventSample {
			select {
			case samples <- payload.(Sample):
			default:
			}
		}
		if name == EventMonitor {
			if state, ok := payload.(MonitorState); ok && state.Error != "" {
				select {
				case failures <- state.Error:
				default:
				}
			}
		}
	})

	if _, err := service.StartMonitor(400); err != nil {
		t.Fatalf("StartMonitor: %v", err)
	}

	timeout := time.After(8 * time.Second)
	received := 0
	for received < 3 {
		select {
		case sample := <-samples:
			if sample.CPUTempC == nil || sample.SysTempC == nil || sample.CPURpm == nil || sample.GPURpm == nil {
				t.Fatalf("sample missing fields: cpuTemp=%v sysTemp=%v cpuRpm=%v gpuRpm=%v",
					sample.CPUTempC, sample.SysTempC, sample.CPURpm, sample.GPURpm)
			}
			received++
			t.Logf("sample %d: %d°C/%d°C %d/%d RPM", received, *sample.CPUTempC, *sample.SysTempC, *sample.CPURpm, *sample.GPURpm)
		case message := <-failures:
			t.Fatalf("monitor reported an error: %s", message)
		case <-timeout:
			t.Fatalf("only %d samples in 8s", received)
		}
	}

	service.StopMonitor()
	if state := service.MonitorState(); state.Running {
		t.Errorf("monitor still running after StopMonitor: %+v", state)
	}
}

func TestLiveWriteRoundTrip(t *testing.T) {
	liveOnly(t)
	service := NewService()
	defer service.Close()

	original, err := service.Settings()
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if original.Power.Kind == "unknown" {
		t.Skip("power profile not recognised, refusing to write")
	}

	target := "silent"
	if original.Power.Kind == "silent" {
		target = "perf"
	}

	updated, err := service.SetPower(target)
	if err != nil {
		t.Fatalf("SetPower(%s): %v", target, err)
	}
	if updated.Power.Kind != target {
		t.Fatalf("readback after write = %q, want %q", updated.Power.Kind, target)
	}
	t.Logf("wrote power=%s and read it back", target)

	restored, err := service.SetPower(original.Power.Kind)
	if err != nil {
		t.Fatalf("restore SetPower(%s): %v", original.Power.Kind, err)
	}
	if restored.Power.Kind != original.Power.Kind {
		t.Fatalf("restored profile = %q, want %q", restored.Power.Kind, original.Power.Kind)
	}
	t.Logf("restored power=%s", restored.Power.Kind)
}

func TestLiveCurveAppliesAndRestores(t *testing.T) {
	liveOnly(t)
	service := NewService()
	defer service.Close()

	if _, err := service.StartMonitor(300); err != nil {
		t.Fatalf("StartMonitor: %v", err)
	}
	defer service.StopMonitor()

	original, err := service.Settings()
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	t.Logf("fan before: cpu=%+v gpu=%+v", original.FanCPU, original.FanGPU)

	curve := DefaultCurve()
	curve.Enabled = true
	curve.Target = "both"
	curve.Points = []CurvePoint{{TempC: 30, Duty: 20}, {TempC: 100, Duty: 20}}
	curve.MinDuty = 20
	curve.MaxDuty = 20
	curve.Hysteresis = 1
	curve.PanicTemp = 110

	if _, err := service.SetCurve(curve); err != nil {
		t.Fatalf("SetCurve: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	state := service.CurveState()
	for time.Now().Before(deadline) {
		state = service.CurveState()
		if state.AppliedPWM > 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if state.AppliedPWM != 51 {
		t.Fatalf("applied pwm = %d, want 51 (20%% of 255)", state.AppliedPWM)
	}
	t.Logf("curve applied pwm=%d duty=%d temp=%v", state.AppliedPWM, state.Duty, state.TempC)

	settings, err := service.Settings()
	if err != nil {
		t.Fatalf("Settings after curve: %v", err)
	}
	if settings.FanCPU.Kind != "custom" || settings.FanCPU.PWM != 51 {
		t.Fatalf("cpu fan = %+v, want custom 51", settings.FanCPU)
	}

	curve.Enabled = false
	if _, err := service.SetCurve(curve); err != nil {
		t.Fatalf("disable curve: %v", err)
	}

	settings, err = service.Settings()
	if err != nil {
		t.Fatalf("Settings after disable: %v", err)
	}
	if settings.FanCPU.Kind != "auto" || settings.FanGPU.Kind != "auto" {
		t.Fatalf("fan after disable = %+v/%+v, want auto/auto", settings.FanCPU, settings.FanGPU)
	}
	t.Logf("curve disabled, fans back to auto")
}
