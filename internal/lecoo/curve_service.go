package lecoo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func curvePath() (string, error) {
	return configPath("fan-curve.json")
}

func loadCurve() FanCurve {
	path, err := curvePath()
	if err != nil {
		return DefaultCurve()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultCurve()
	}
	var curve FanCurve
	if err := json.Unmarshal(data, &curve); err != nil {
		return DefaultCurve()
	}
	curve.Enabled = false
	normalized, err := curve.Normalized()
	if err != nil {
		return DefaultCurve()
	}
	return normalized
}

func (s *Service) persistCurve(curve FanCurve) error {
	path, err := curvePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	stored := curve
	stored.Enabled = false
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func (s *Service) curveStateLocked() CurveState {
	state := CurveState{
		Curve:      s.curve,
		TempC:      s.lastTemp,
		Duty:       s.lastDuty,
		AppliedPWM: s.lastPwm,
		LastError:  s.lastError,
	}
	if !s.lastWriteAt.IsZero() {
		state.LastWriteAt = s.lastWriteAt.UnixMilli()
	}
	return state
}

func (s *Service) CurveState() CurveState {
	s.curveMu.Lock()
	defer s.curveMu.Unlock()
	return s.curveStateLocked()
}

func (s *Service) SetCurve(curve FanCurve) (CurveState, error) {
	normalized, err := curve.Normalized()
	if err != nil {
		return s.CurveState(), err
	}

	s.curveMu.Lock()
	wasEnabled := s.curve.Enabled
	previousTarget := s.curve.Target
	s.curve = normalized
	s.lastPwm = -1
	s.lastWriteAt = time.Time{}
	s.lastError = ""
	s.curveMu.Unlock()

	if wasEnabled && (!normalized.Enabled || previousTarget != normalized.Target) {
		if _, err := s.runner.Exec(context.Background(), "fan", previousTarget, "auto"); err != nil {
			s.curveMu.Lock()
			s.lastError = err.Error()
			s.curveMu.Unlock()
		}
	}

	if err := s.persistCurve(normalized); err != nil {
		s.curveMu.Lock()
		s.lastError = err.Error()
		s.curveMu.Unlock()
	}

	state := s.CurveState()
	s.emitEvent(EventCurve, state)
	return state, nil
}

func (s *Service) stepCurve(sample Sample) {
	s.curveMu.Lock()
	curve := s.curve
	enabled := curve.Enabled
	temp := sample.CPUTempC
	lastPwm := s.lastPwm
	lastWriteAt := s.lastWriteAt
	s.curveMu.Unlock()

	if !enabled || temp == nil {
		return
	}

	duty := curve.DutyFor(*temp)
	pwm := curve.PWMDuty(*temp)

	if lastPwm >= 0 {
		elapsed := time.Since(lastWriteAt)
		if elapsed < curveMinInterval {
			s.publishCurve(temp, duty, lastPwm, "")
			return
		}
		if absInt(pwm-lastPwm) < curve.Hysteresis && elapsed < curveSlowInterval {
			s.publishCurve(temp, duty, lastPwm, "")
			return
		}
	}

	if _, err := s.runner.Exec(context.Background(), "fan", curve.Target, "custom", strconv.Itoa(pwm)); err != nil {
		s.curveMu.Lock()
		s.lastError = err.Error()
		s.curveMu.Unlock()
		s.publishCurve(temp, duty, lastPwm, err.Error())
		return
	}

	s.curveMu.Lock()
	s.lastPwm = pwm
	s.lastWriteAt = time.Now()
	s.lastError = ""
	s.curveMu.Unlock()
	s.publishCurve(temp, duty, pwm, "")
}

func (s *Service) publishCurve(temp *int, duty int, applied int, lastError string) {
	s.curveMu.Lock()
	s.lastTemp = temp
	s.lastDuty = duty
	if lastError != "" {
		s.lastError = lastError
	}
	state := s.curveStateLocked()
	s.curveMu.Unlock()
	state.AppliedPWM = applied
	s.emitEvent(EventCurve, state)
}

func (s *Service) stopCurve() {
	s.curveMu.Lock()
	enabled := s.curve.Enabled
	target := s.curve.Target
	s.curve.Enabled = false
	s.curveMu.Unlock()
	if !enabled {
		return
	}
	_, _ = s.runner.Exec(context.Background(), "fan", target, "auto")
	s.emitEvent(EventCurve, s.CurveState())
}
