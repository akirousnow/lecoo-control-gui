package lecoo

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultCliPath = `C:\Program Files\LecooControlCenter\lecoo-ctrl.exe`

func DefaultPath() string {
	if p := os.Getenv("LECOO_CTRL_PATH"); p != "" {
		return p
	}
	if _, err := os.Stat(defaultCliPath); err == nil {
		return defaultCliPath
	}
	if p, err := exec.LookPath("lecoo-ctrl.exe"); err == nil {
		return p
	}
	return defaultCliPath
}

type Runner struct {
	Path    string
	Timeout time.Duration
}

func NewRunner() *Runner {
	return &Runner{Path: DefaultPath(), Timeout: 6 * time.Second}
}

func (r *Runner) command(ctx context.Context, args ...string) *exec.Cmd {
	path := r.Path
	if path == "" {
		path = DefaultPath()
	}
	cmd := exec.CommandContext(ctx, path, args...)
	hideConsoleWindow(cmd)
	return cmd
}

func (r *Runner) Exec(ctx context.Context, args ...string) (string, error) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 6 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := r.command(runCtx, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", r.classify(err, runCtx, stderr.String(), args)
	}
	return stdout.String(), nil
}

func (r *Runner) classify(err error, ctx context.Context, stderr string, args []string) error {
	if ctx.Err() == context.DeadlineExceeded {
		return &CliError{Kind: "timeout", Message: "命令超时", Detail: strings.Join(args, " ")}
	}
	var execErr *exec.Error
	if errors.As(err, &execErr) || errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
		return &CliError{Kind: "missing-cli", Message: "找不到 lecoo-ctrl.exe", Detail: r.Path}
	}
	msg := firstLine(stderr)
	if msg == "" {
		msg = err.Error()
	}
	return &CliError{Kind: "rejected", Message: msg, Detail: stderr}
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(line, "error:"))
	}
	return ""
}

func (r *Runner) Settings(ctx context.Context) (Settings, error) {
	out, err := r.Exec(ctx, "daemon", "settings", "read")
	if err != nil {
		return Settings{}, err
	}
	return ParseSettings(out)
}

func (r *Runner) Battery(ctx context.Context) (Battery, error) {
	out, err := r.Exec(ctx, "charge")
	if err != nil {
		return Battery{}, err
	}
	return ParseBattery(out)
}

func (r *Runner) Info(ctx context.Context) (Info, error) {
	out, err := r.Exec(ctx, "info")
	if err != nil {
		return Info{}, err
	}
	info := ParseInfo(out)
	version, err := r.Exec(ctx, "--version")
	if err == nil {
		info.CliVersion = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(version), "lecoo-ctrl.exe"))
	}
	return info, nil
}

func (r *Runner) RawSettings(ctx context.Context) (string, error) {
	return r.Exec(ctx, "daemon", "settings", "read")
}

func (r *Runner) Health(ctx context.Context) Health {
	health := Health{CliPath: r.Path}
	if _, err := os.Stat(r.Path); err == nil {
		health.CliFound = true
	}
	if version, err := r.Exec(ctx, "--version"); err == nil {
		health.Version = strings.TrimSpace(version)
	}
	if _, err := r.Exec(ctx, "daemon", "settings", "read"); err == nil {
		health.DaemonUp = true
	} else {
		health.Message = err.Error()
	}
	return health
}

func (r *Runner) Sample(ctx context.Context) (Sample, error) {
	temps, err := r.Exec(ctx, "temps")
	if err != nil {
		return Sample{}, err
	}
	sample := ParseSample(temps)
	fans, err := r.Exec(ctx, "fans")
	if err == nil {
		extra := ParseSample(fans)
		sample.CPURpm = extra.CPURpm
		sample.GPURpm = extra.GPURpm
	}
	return sample, nil
}
