package main

import (
	"context"
	"os/exec"
	"syscall"

	"lecoo-control-gui/internal/lecoo"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	svc *lecoo.Service
}

func NewApp() *App {
	return &App{svc: lecoo.NewService()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.svc.SetEmitter(func(name string, payload any) {
		runtime.EventsEmit(ctx, name, payload)
	})
	_, _ = a.svc.StartMonitor(1000)
}

func (a *App) shutdown(ctx context.Context) {
	a.svc.Close()
}

func (a *App) Info() (lecoo.Info, error) {
	return a.svc.Info()
}

func (a *App) Health() lecoo.Health {
	return a.svc.Health()
}

func (a *App) RawSettings() (string, error) {
	return a.svc.RawSettings()
}

func (a *App) Settings() (lecoo.Settings, error) {
	return a.svc.Settings()
}

func (a *App) Battery() (lecoo.Battery, error) {
	return a.svc.Battery()
}

func (a *App) Sample() (lecoo.Sample, error) {
	return a.svc.Sample()
}

func (a *App) SetPower(profile string) (lecoo.Settings, error) {
	return a.svc.SetPower(profile)
}

func (a *App) SetFan(target string, mode string, pwm int) (lecoo.Settings, error) {
	return a.svc.SetFan(target, mode, pwm)
}

func (a *App) SetChargePreset(preset string) (lecoo.Settings, error) {
	return a.svc.SetChargePreset(preset)
}

func (a *App) SetChargeRange(min int, max int) (lecoo.Settings, error) {
	return a.svc.SetChargeRange(min, max)
}

func (a *App) SetTelemetry(enabled bool) (lecoo.Settings, error) {
	return a.svc.SetTelemetry(enabled)
}

func (a *App) Reapply() (lecoo.Settings, error) {
	return a.svc.Reapply()
}

func (a *App) Reset() (lecoo.Settings, error) {
	return a.svc.Reset()
}

func (a *App) CliPath() string {
	return a.svc.CliPath()
}

func (a *App) SetCliPath(path string) string {
	return a.svc.SetCliPath(path)
}

func (a *App) MonitorState() lecoo.MonitorState {
	return a.svc.MonitorState()
}

func (a *App) StartMonitor(rateMs int) (lecoo.MonitorState, error) {
	return a.svc.StartMonitor(rateMs)
}

func (a *App) StopMonitor() lecoo.MonitorState {
	return a.svc.StopMonitor()
}

func (a *App) OpenHwtest() error {
	cmd := exec.Command(a.svc.CliPath(), "hwtest")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010}
	return cmd.Start()
}

func (a *App) Minimise() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) ToggleMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) IsMaximised() bool {
	return runtime.WindowIsMaximised(a.ctx)
}

func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

func (a *App) SetTheme(theme string) {
	if theme == "light" {
		runtime.WindowSetLightTheme(a.ctx)
		return
	}
	runtime.WindowSetDarkTheme(a.ctx)
}

func (a *App) SetCompactWindow(compact bool, alwaysOnTop bool) {
	if compact {
		runtime.WindowSetMinSize(a.ctx, 360, 220)
		runtime.WindowSetSize(a.ctx, 380, 232)
		runtime.WindowSetAlwaysOnTop(a.ctx, alwaysOnTop)
		return
	}
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	runtime.WindowSetMinSize(a.ctx, 1060, 700)
	runtime.WindowSetSize(a.ctx, 1320, 880)
	runtime.WindowCenter(a.ctx)
}

func (a *App) SetAlwaysOnTop(enabled bool) {
	runtime.WindowSetAlwaysOnTop(a.ctx, enabled)
}
