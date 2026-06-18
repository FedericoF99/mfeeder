//go:build windows

package windows

import (
	"context"
	"mfeeder/internal/config"
	"mfeeder/internal/shutdown"
	"testing"
)

func TestWatcher(t *testing.T) {
	w := &WinWatcher{Cfg: &config.Conf{}}

	ctx, cancel := context.WithCancel(context.Background())
	sdManager := shutdown.Manager{Cancel: cancel}

	ch, _ := w.Watch(&sdManager)

	postMessageW.Call(w.hwnd, uintptr(WmClose), 0, 0)

	for {
		select {
		case <-ctx.Done():
			w.WG.Wait()
			println("main loop stopped")
			return
		case event := <-ch:
			println(event.WindowEvent, event.Window.Title)
		}
	}

}
