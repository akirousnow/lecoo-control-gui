package lecoo

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

const (
	EventSample   = "lecoo:sample"
	EventSettings = "lecoo:settings"
	EventBattery  = "lecoo:battery"
	EventMonitor  = "lecoo:monitor"
	EventCurve    = "lecoo:curve"

	curveMinInterval  = 4 * time.Second
	curveSlowInterval = 30 * time.Second
)

var powerProfiles = map[string]bool{"silent": true, "default": true, "perf": true}

var fanTargets = map[string]bool{"cpu": true, "gpu": true, "both": true}

var fanModes = map[string]bool{"auto": true, "full": true, "turbo": true, "custom": true}

var chargePresets = map[string]bool{"full": true, "desk": true, "lifespan": true, "balanced": true, "high": true}

type monitorRun struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	rateMs int
}

type Service struct {
	runner *Runner
	job    *killJob
	emit   func(string, any)

	mu  sync.Mutex
	mon *monitorRun

	curveMu     sync.Mutex
	curve       FanCurve
	lastPwm     int
	lastTemp    *int
	lastDuty    int
	lastWriteAt time.Time
	lastError   string
}

func NewService() *Service {
	service := &Service{runner: NewRunner(), lastPwm: -1, curve: loadCurve()}
	if job, err := newKillJob(); err == nil {
		service.job = job
	}
	return service
}

func (s *Service) SetEmitter(emit func(string, any)) {
	s.emit = emit
}

func (s *Service) emitEvent(name string, payload any) {
	if s.emit != nil {
		s.emit(name, payload)
	}
}

func invalid(message string) *CliError {
	return &CliError{Kind: "invalid", Message: message}
}

func (s *Service) CliPath() string {
	return s.runner.Path
}

func (s *Service) SetCliPath(path string) string {
	s.runner.Path = path
	return s.runner.Path
}

func (s *Service) Settings() (Settings, error) {
	return s.refreshSettings(context.Background())
}

func (s *Service) Battery() (Battery, error) {
	battery, err := s.runner.Battery(context.Background())
	if err != nil {
		return Battery{}, err
	}
	s.emitEvent(EventBattery, battery)
	return battery, nil
}

func (s *Service) Sample() (Sample, error) {
	return s.runner.Sample(context.Background())
}

func (s *Service) Info() (Info, error) {
	return s.runner.Info(context.Background())
}

func (s *Service) Health() Health {
	return s.runner.Health(context.Background())
}

func (s *Service) RawSettings() (string, error) {
	return s.runner.RawSettings(context.Background())
}

func (s *Service) refreshSettings(ctx context.Context) (Settings, error) {
	settings, err := s.runner.Settings(ctx)
	if err != nil {
		return Settings{}, err
	}
	s.emitEvent(EventSettings, settings)
	return settings, nil
}

func (s *Service) write(ctx context.Context, args ...string) (Settings, error) {
	if _, err := s.runner.Exec(ctx, args...); err != nil {
		return Settings{}, err
	}
	return s.refreshSettings(ctx)
}

func (s *Service) SetPower(profile string) (Settings, error) {
	if !powerProfiles[profile] {
		return Settings{}, invalid("未知的电源模式")
	}
	return s.write(context.Background(), "power", profile)
}

func (s *Service) SetFan(target string, mode string, pwm int) (Settings, error) {
	if !fanTargets[target] {
		return Settings{}, invalid("未知的风扇目标")
	}
	if !fanModes[mode] {
		return Settings{}, invalid("未知的风扇模式")
	}
	args := []string{"fan", target, mode}
	if mode == "custom" {
		if pwm < 0 || pwm > fanDutyMax {
			return Settings{}, invalid("占空比需要落在 0 到 220")
		}
		args = append(args, strconv.Itoa(pwm))
	}

	s.curveMu.Lock()
	curveWasEnabled := s.curve.Enabled
	s.curve.Enabled = false
	s.curveMu.Unlock()

	settings, err := s.write(context.Background(), args...)
	if curveWasEnabled {
		s.emitEvent(EventCurve, s.CurveState())
	}
	return settings, err
}

func (s *Service) SetChargePreset(preset string) (Settings, error) {
	if !chargePresets[preset] {
		return Settings{}, invalid("未知的充电预设")
	}
	return s.write(context.Background(), "charge", preset)
}

func (s *Service) SetChargeRange(min int, max int) (Settings, error) {
	if min < 20 || max > 100 || min >= max {
		return Settings{}, invalid("自定义区间需要 20 到 100 且下限小于上限")
	}
	return s.write(context.Background(), "charge", "--min", strconv.Itoa(min), "--max", strconv.Itoa(max))
}

func (s *Service) SetTelemetry(enabled bool) (Settings, error) {
	action := "disable"
	if enabled {
		action = "enable"
	}
	return s.write(context.Background(), "daemon", "telemetry", action)
}

func (s *Service) Reapply() (Settings, error) {
	return s.write(context.Background(), "daemon", "settings", "apply")
}

func (s *Service) Reset() (Settings, error) {
	return s.write(context.Background(), "daemon", "settings", "reset")
}

func (s *Service) MonitorState() MonitorState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mon == nil {
		return MonitorState{}
	}
	return MonitorState{Running: true, RateMs: s.mon.rateMs}
}

func (s *Service) StartMonitor(rateMs int) (MonitorState, error) {
	if rateMs < 200 {
		rateMs = 200
	}
	if rateMs > 10000 {
		rateMs = 10000
	}
	s.StopMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	rate := strconv.FormatFloat(float64(rateMs)/1000, 'f', 3, 64)
	cmd := s.runner.command(ctx, "monitoring", "--rate", rate)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return MonitorState{}, &CliError{Kind: "internal", Message: err.Error()}
	}
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Start(); err != nil {
		cancel()
		return MonitorState{}, s.runner.classify(err, ctx, errBuf.String(), []string{"monitoring"})
	}
	if s.job != nil {
		_ = s.job.attach(cmd)
	}
	run := &monitorRun{ctx: ctx, cancel: cancel, done: make(chan struct{}), rateMs: rateMs}
	s.mu.Lock()
	s.mon = run
	s.mu.Unlock()
	go s.pump(run, cmd, stdout, &errBuf)
	state := MonitorState{Running: true, RateMs: rateMs}
	s.emitEvent(EventMonitor, state)
	return state, nil
}

func (s *Service) pump(run *monitorRun, cmd *exec.Cmd, stdout io.Reader, errBuf *bytes.Buffer) {
	defer close(run.done)
	scanner := bufio.NewScanner(stdout)
	scanner.Split(splitSample)
	scanner.Buffer(make([]byte, 0, 4096), 1<<20)
	for scanner.Scan() {
		sample := ParseSample(scanner.Text())
		if sample.Empty() {
			continue
		}
		s.emitEvent(EventSample, sample)
		s.stepCurve(sample)
	}
	waitErr := cmd.Wait()
	s.mu.Lock()
	if s.mon == run {
		s.mon = nil
	}
	s.mu.Unlock()
	state := MonitorState{}
	if waitErr != nil && run.ctx.Err() == nil {
		message := firstLine(errBuf.String())
		if message == "" {
			message = waitErr.Error()
		}
		state.Error = message
	}
	s.emitEvent(EventMonitor, state)
}

func (s *Service) StopMonitor() MonitorState {
	s.mu.Lock()
	run := s.mon
	s.mon = nil
	s.mu.Unlock()
	if run == nil {
		return MonitorState{}
	}
	run.cancel()
	<-run.done
	return MonitorState{}
}

func (s *Service) Close() {
	s.stopCurve()
	s.StopMonitor()
	if s.job != nil {
		s.job.close()
	}
}

func splitSample(data []byte, atEOF bool) (int, []byte, error) {
	for i, b := range data {
		if b == '\r' || b == '\n' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}
