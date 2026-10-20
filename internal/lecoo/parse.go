package lecoo

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	debugFieldRe = regexp.MustCompile(`^ {4}([a-z_]+): (.*)$`)
	customRe     = regexp.MustCompile(`^Custom\s*\(\s*(\d+)\s*,?\s*\)$`)
	minRe        = regexp.MustCompile(`min:\s*(\d+)`)
	maxRe        = regexp.MustCompile(`max:\s*(\d+)`)
	celsiusRe    = regexp.MustCompile(`(\d+)\s*°C`)
	rpmRe        = regexp.MustCompile(`(\d+)\s*RPM`)
	versionRe    = regexp.MustCompile(`版本\s*(\d+)`)
	hramRe       = regexp.MustCompile(`0x[0-9A-Fa-f]+`)
	percentRe    = regexp.MustCompile(`(\d+)\s*%`)
)

func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func splitDebugFields(raw string) map[string]string {
	fields := map[string]string{}
	name := ""
	parts := []string{}
	flush := func() {
		if name != "" {
			fields[name] = strings.TrimSuffix(normalize(strings.Join(parts, " ")), ",")
		}
		name = ""
		parts = parts[:0]
	}
	for _, line := range strings.Split(raw, "\n") {
		if line == "}" {
			flush()
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if m := debugFieldRe.FindStringSubmatch(line); m != nil {
			flush()
			name = m[1]
			parts = append(parts, m[2])
			continue
		}
		if name != "" {
			parts = append(parts, strings.TrimSpace(line))
		}
	}
	flush()
	return fields
}

func parseFanMode(raw string) FanMode {
	v := normalize(raw)
	switch v {
	case "Auto":
		return FanMode{Kind: "auto"}
	case "Full":
		return FanMode{Kind: "full"}
	case "Turbo":
		return FanMode{Kind: "turbo"}
	}
	if m := customRe.FindStringSubmatch(v); m != nil {
		pwm, _ := strconv.Atoi(m[1])
		return FanMode{Kind: "custom", PWM: pwm}
	}
	return FanMode{Kind: "unknown"}
}

func parsePowerProfile(raw string) PowerProfile {
	switch normalize(raw) {
	case "Silent":
		return PowerProfile{Kind: "silent"}
	case "Default":
		return PowerProfile{Kind: "default"}
	case "Performance":
		return PowerProfile{Kind: "perf"}
	}
	return PowerProfile{Kind: "unknown", Raw: normalize(raw)}
}

func parseChargeTarget(raw string) ChargeTarget {
	v := normalize(raw)
	if strings.HasPrefix(v, "Preserve") {
		target := ChargeTarget{Kind: "preserve"}
		if m := minRe.FindStringSubmatch(v); m != nil {
			n, _ := strconv.Atoi(m[1])
			target.Min = &n
		}
		if m := maxRe.FindStringSubmatch(v); m != nil {
			n, _ := strconv.Atoi(m[1])
			target.Max = &n
		}
		return target
	}
	switch v {
	case "Full":
		return ChargeTarget{Kind: "preset", Preset: "full"}
	case "Desk":
		return ChargeTarget{Kind: "preset", Preset: "desk"}
	case "Lifespan":
		return ChargeTarget{Kind: "preset", Preset: "lifespan"}
	case "Balanced":
		return ChargeTarget{Kind: "preset", Preset: "balanced"}
	case "High":
		return ChargeTarget{Kind: "preset", Preset: "high"}
	}
	return ChargeTarget{Kind: "unknown", Raw: v}
}

func ParseSettings(raw string) (Settings, error) {
	fields := splitDebugFields(raw)
	if _, ok := fields["power_profile"]; !ok {
		return Settings{}, errors.New("settings block not recognised")
	}
	enabled := normalize(fields["telemetry_enabled"]) == "true"
	settings := Settings{
		TelemetryEnabled:  enabled,
		TelemetryClientID: normalize(fields["telemetry_client_id"]),
		Power:             parsePowerProfile(fields["power_profile"]),
		Charge:            parseChargeTarget(fields["charge"]),
		FanCPU:            parseFanMode(fields["fan_mode_cpu"]),
		FanGPU:            parseFanMode(fields["fan_mode_gpu"]),
		ReadAt:            time.Now().UnixMilli(),
	}
	return settings, nil
}

func firstInts(re *regexp.Regexp, raw string, n int) []int {
	matches := re.FindAllStringSubmatch(raw, n)
	out := make([]int, 0, len(matches))
	for _, m := range matches {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

func ParseSample(raw string) Sample {
	sample := Sample{At: time.Now().UnixMilli()}
	temps := firstInts(celsiusRe, raw, 2)
	if len(temps) > 0 {
		sample.CPUTempC = &temps[0]
	}
	if len(temps) > 1 {
		sample.SysTempC = &temps[1]
	}
	rpms := firstInts(rpmRe, raw, 2)
	if len(rpms) > 0 {
		sample.CPURpm = &rpms[0]
	}
	if len(rpms) > 1 {
		sample.GPURpm = &rpms[1]
	}
	return sample
}

func (s Sample) Empty() bool {
	return s.CPUTempC == nil && s.SysTempC == nil && s.CPURpm == nil && s.GPURpm == nil
}

func ParseBattery(raw string) (Battery, error) {
	lines := []string{}
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	if len(lines) < 2 {
		return Battery{}, errors.New("charge output not recognised")
	}
	battery := Battery{At: time.Now().UnixMilli()}
	body := lines[1:]
	if len(body) > 0 {
		battery.Target = valueAfterColon(body[0])
	}
	if len(body) > 1 {
		battery.Effective = valueAfterColon(body[1])
	}
	if len(body) > 2 {
		if m := percentRe.FindStringSubmatch(body[2]); m != nil {
			v, _ := strconv.Atoi(m[1])
			battery.Percent = &v
		}
	}
	if battery.Target == "" {
		return Battery{}, errors.New("charge output not recognised")
	}
	return battery, nil
}

func valueAfterColon(line string) string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return strings.TrimSpace(line)
	}
	return strings.TrimSpace(line[idx+1:])
}

func ParseInfo(raw string) Info {
	lines := []string{}
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	info := Info{}
	if len(lines) > 0 {
		first := valueAfterColon(lines[0])
		if idx := strings.Index(first, "("); idx > 0 {
			info.Controller = strings.TrimSpace(first[:idx])
		} else {
			info.Controller = first
		}
		if m := versionRe.FindStringSubmatch(first); m != nil {
			info.ControllerVersion, _ = strconv.Atoi(m[1])
		}
	}
	for _, line := range lines {
		if m := hramRe.FindStringSubmatch(line); m != nil {
			info.HramOffset = m[0]
			break
		}
	}
	if len(lines) > 0 {
		info.DaemonVersion = valueAfterColon(lines[len(lines)-1])
	}
	return info
}
