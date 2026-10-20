package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func widgetPositionPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "lecoo-control-gui", "widget-position.json"), nil
}

func loadWidgetPosition() (WidgetPosition, bool) {
	path, err := widgetPositionPath()
	if err != nil {
		return WidgetPosition{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return WidgetPosition{}, false
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	var position WidgetPosition
	if err := json.Unmarshal(data, &position); err != nil {
		return WidgetPosition{}, false
	}
	return position, true
}

func (a *App) SaveWidgetPosition() (WidgetPosition, error) {
	x, y := runtime.WindowGetPosition(a.ctx)
	position := WidgetPosition{X: x, Y: y}
	path, err := widgetPositionPath()
	if err != nil {
		return position, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return position, err
	}
	data, err := json.MarshalIndent(position, "", "  ")
	if err != nil {
		return position, err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return position, err
	}
	return position, nil
}
