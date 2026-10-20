package main

import (
	_ "embed"
	"log"

	"fyne.io/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

func (a *App) startTray() {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("tray stopped: %v", recovered)
			}
		}()
		systray.Run(a.onTrayReady, func() {})
	}()
}

func (a *App) stopTray() {
	systray.Quit()
}

func (a *App) onTrayReady() {
	systray.SetIcon(trayIcon)
	systray.SetTooltip("来酷控制中心")

	show := systray.AddMenuItem("显示主窗口", "把窗口调到最前")
	compact := systray.AddMenuItem("切换精简模式", "在完整界面与桌面小组件之间切换")
	systray.AddSeparator()
	quit := systray.AddMenuItem("退出", "结束程序")

	go func() {
		for {
			select {
			case <-show.ClickedCh:
				runtime.WindowUnminimise(a.ctx)
				runtime.WindowShow(a.ctx)
			case <-compact.ClickedCh:
				runtime.EventsEmit(a.ctx, "app:compact-toggle")
			case <-quit.ClickedCh:
				runtime.Quit(a.ctx)
				return
			}
		}
	}()
}
