package main

import (
	"os"

	"fyne.io/systray"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/selfbase-dev/s2-sync/internal/service"
)

// registerTray installs NSStatusItem callbacks into the existing
// NSApplication without owning the main event loop. Using Register (not
// Run) lets Wails keep the main thread — the Wails run loop dispatches
// tray target-actions for us. systray.Quit should be called on
// shutdown (see OnShutdown in main).
func registerTray(app *App) {
	systray.Register(func() { onTrayReady(app) }, func() {})
}

func onTrayReady(app *App) {
	systray.SetTitle("S2")
	systray.SetTooltip("s2sync")

	mShow := systray.AddMenuItem("Show window", "Open the s2sync window")
	mStatus := systray.AddMenuItem("Status: idle", "Current sync status")
	mStatus.Disable()
	systray.AddSeparator()
	mStart := systray.AddMenuItem("Start sync", "Resume syncing")
	mStop := systray.AddMenuItem("Stop sync", "Pause syncing")
	mStop.Hide()
	systray.AddSeparator()
	mAutostart := systray.AddMenuItemCheckbox("Start at login", "Launch s2sync automatically when you log in", service.IsAutostartEnabled())
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit s2sync", "Exit")

	// Keep tray labels/visibility in sync with service events.
	go func() {
		ch := app.svc.Subscribe()
		for range ch {
			refreshTray(app, mStatus, mStart, mStop)
		}
	}()

	// Menu item click handling. systray guarantees these channels are
	// drained on the goroutine running them; keep them lightweight.
	for {
		select {
		case <-mShow.ClickedCh:
			if app.ctx != nil {
				wailsruntime.WindowShow(app.ctx)
			}
		case <-mStart.ClickedCh:
			if app.ctx == nil {
				continue
			}
			st := app.svc.Status()
			if st.Mount != nil {
				_ = app.svc.Start(app.ctx, *st.Mount)
			} else {
				wailsruntime.WindowShow(app.ctx)
			}
		case <-mStop.ClickedCh:
			_ = app.svc.Stop()
		case <-mAutostart.ClickedCh:
			toggleAutostart(mAutostart)
		case <-mQuit.ClickedCh:
			_ = app.svc.Stop()
			if app.ctx != nil {
				wailsruntime.Quit(app.ctx)
			}
			return
		}
	}
}

func refreshTray(app *App, mStatus, mStart, mStop *systray.MenuItem) {
	switch app.svc.Status().Status {
	case service.StatusRunning:
		systray.SetTitle("S2 ●")
		mStatus.SetTitle("Status: running")
		mStart.Hide()
		mStop.Show()
	case service.StatusStopping:
		systray.SetTitle("S2 …")
		mStatus.SetTitle("Status: stopping…")
		mStart.Hide()
		mStop.Show()
	case service.StatusError:
		systray.SetTitle("S2 ✕")
		mStatus.SetTitle("Status: error")
		mStart.Show()
		mStop.Hide()
	default:
		systray.SetTitle("S2")
		mStatus.SetTitle("Status: idle")
		mStart.Show()
		mStop.Hide()
	}
}

func toggleAutostart(item *systray.MenuItem) {
	enable := !item.Checked()
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if err := service.SetAutostart(enable, exe); err != nil {
		return
	}
	if enable {
		item.Check()
	} else {
		item.Uncheck()
	}
}
