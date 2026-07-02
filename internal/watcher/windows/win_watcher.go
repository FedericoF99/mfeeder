//go:build windows

package windows

import (
	"context"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/shutdown"
	"mfeeder/internal/watcher/core"
	"runtime"
	"slices"
	"sync"
	"syscall"
	"unsafe"
)

type RawWindowsEvent struct {
	hwnd  uintptr
	event core.WindowEventType
}

type WinWatcher struct {
	Cfg  *config.Conf
	WG   sync.WaitGroup
	hwnd uintptr
}

func (w *WinWatcher) Snapshot(ctx context.Context) ([]core.Window, error) {

	info := make([]core.Window, 0)
	infoPtr := unsafe.Pointer(&info)
	lParam := uintptr(infoPtr)

	cb := enumWindowsCallback(ctx, w.Cfg)
	res, _, err := enumWindows.Call(cb, lParam)

	if res != 0 {
		err = nil
	}

	return info, err
}

func (w *WinWatcher) Watch(sdManager *shutdown.Manager) (<-chan core.WindowEvent, error) {

	ch := make(chan core.WindowEvent, 50)
	chRaw := make(chan RawWindowsEvent, 50)
	ready := make(chan error, 1)

	w.WG.Add(2)
	go func() {
		defer w.WG.Done()
		defer close(ch)
		w.eventLoop(ch, chRaw)
	}()

	go func() {
		defer w.WG.Done()
		defer close(chRaw)
		w.winMessageLoop(sdManager, chRaw, ready)
	}()

	err := <-ready
	close(ready)
	return ch, err
}

func (w *WinWatcher) Close(sdManager *shutdown.Manager) {
	sdManager.Shutdown()

	if w.hwnd != 0 {
		_, _, _ = postMessageW.Call(w.hwnd, uintptr(WmClose), 0, 0)
	}

	w.WG.Wait()
}

// hwnd is a window handle (basically a pointer to the window)
// lParam is a pointer to an application-defined value passed to the callback from EnumWindows
func enumWindowsCallback(ctx context.Context, c *config.Conf) uintptr {
	fHwnd := getForegroundHandle()

	return syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		if ctx.Err() != nil {
			return 0
		}

		if !isOpenedWindow(hwnd) {
			return 1
		}

		window, err := getWindowInfo(hwnd, fHwnd)
		if err != nil {
			return 1
		}
		if isExcluded(hwnd, window, c) {
			return 1
		}

		info := (*[]core.Window)(unsafe.Pointer(lParam))
		*info = append(*info, window)

		return 1
	})
}

func eventHookCallback(chRaw chan<- RawWindowsEvent) uintptr {
	return syscall.NewCallback(func(hWinEventHook uintptr, event uint32, hwnd uintptr, idObject int32, idChild int32, idEventThread uint32, dwmsEventTime uint32) uintptr {
		if idObject != 0 {
			return 0
		}
		if idChild != 0 {
			return 0
		}
		if !slices.Contains([]uint32{
			EventSystemForeground, EventObjectShow,
			EventObjectHide, EventObjectDestroy,
			EventSystemMinimizeStart, EventSystemMinimizeEnd}, event) {
			return 0
		}

		var windowEvent core.WindowEventType
		var ok bool

		switch event {
		case EventSystemForeground:
			windowEvent = core.WindowFocused
			ok = isOpenedWindow(hwnd)
		case EventObjectShow:
			windowEvent = core.WindowOpened
			ok = isOpenedWindow(hwnd)
		case EventObjectHide:
			windowEvent = core.WindowClosed
			ok = isAppWindow(hwnd)
		case EventObjectDestroy:
			windowEvent = core.WindowClosed
			ok = isAppWindow(hwnd)
		case EventSystemMinimizeStart:
			windowEvent = core.WindowClosed
			ok = isOpenedWindow(hwnd)
		case EventSystemMinimizeEnd:
			windowEvent = core.WindowOpened
			ok = isAppWindow(hwnd)
		}

		if !ok {
			return 0
		}

		select {
		case chRaw <- RawWindowsEvent{hwnd: hwnd, event: windowEvent}:
		default:
			// drop event
		}

		return 0
	})
}

func (w *WinWatcher) eventLoop(ch chan<- core.WindowEvent, chRaw <-chan RawWindowsEvent) {
	for {
		select {
		case raw, ok := <-chRaw:
			if !ok {
				println("event loop stopped")
				return
			}

			fHwnd := getForegroundHandle()
			window, err := getWindowInfo(raw.hwnd, fHwnd)
			if err != nil {
				continue
			}
			if isExcluded(raw.hwnd, window, w.Cfg) {
				continue
			}

			ch <- core.WindowEvent{
				Window:      window,
				WindowEvent: raw.event,
			}
		}
	}
}

func (w *WinWatcher) winMessageLoop(sdManager *shutdown.Manager, chRaw chan<- RawWindowsEvent, ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cb := eventHookCallback(chRaw)

	hwnd, err := createHiddenWindow(sdManager)
	if err != nil {
		ready <- err
		return
	}

	w.hwnd = hwnd

	foregroundEvHook, _, err := setWinEventHook.Call(uintptr(EventSystemForeground), uintptr(EventSystemForeground), uintptr(0), cb, 0, 0, uintptr(0|2))
	if foregroundEvHook == 0 {
		ready <- err
		return
	}
	defer func() { _, _, _ = procUnhookWinEvent.Call(foregroundEvHook) }()

	minimizedEvHook, _, err := setWinEventHook.Call(uintptr(EventSystemMinimizeStart), uintptr(EventSystemMinimizeEnd), uintptr(0), cb, 0, 0, uintptr(0|2))
	if minimizedEvHook == 0 {
		ready <- err
		return
	}
	defer func() { _, _, _ = procUnhookWinEvent.Call(minimizedEvHook) }()

	objectEvHook, _, err := setWinEventHook.Call(uintptr(EventObjectCreate), uintptr(EventObjectHide), uintptr(0), cb, 0, 0, uintptr(0|2))
	if objectEvHook == 0 {
		ready <- err
		return
	}
	defer func() { _, _, _ = procUnhookWinEvent.Call(objectEvHook) }()

	ready <- nil

	var m msg
	for {
		mPtr := uintptr(unsafe.Pointer(&m))
		ret, _, _ := procGetMessage.Call(mPtr, 0, 0, 0)

		if ret < 0 {
			log.Println("procGetMessage failed")
			sdManager.Shutdown()
			return
		}
		if ret == 0 {
			println("message loop stopped")
			break
		}

		_, _, _ = translateMessage.Call(mPtr)
		_, _, _ = dispatchMessageW.Call(mPtr)
	}
}

func createHiddenWindow(sdManager *shutdown.Manager) (uintptr, error) {
	// handle to the current module
	hInstance, _, err := getModuleHandleW.Call(0)
	if hInstance == 0 {
		return 0, err
	}

	windowName, err := syscall.UTF16PtrFromString("MFeeder Hidden Window")
	if err != nil {
		return 0, err
	}

	className, err := syscall.UTF16PtrFromString("MFeederClass")
	if err != nil {
		return 0, err
	}

	wProc := windowProcCallback(sdManager)

	w := wndclassex{
		size:      uint32(unsafe.Sizeof(wndclassex{})),
		wndProc:   wProc,
		instance:  hInstance,
		className: className,
	}

	atom, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(&w)))
	if atom == 0 {
		return 0, err
	}

	hwnd, _, err := createWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(windowName)),
		uintptr(WsOverlapped), 0, 0, 0, 0, 0, 0, hInstance, 0)

	if hwnd == 0 {
		return 0, err
	}

	return hwnd, nil
}

func windowProcCallback(sdManager *shutdown.Manager) uintptr {
	return syscall.NewCallback(func(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
		switch msg {
		case WmClose:
			sdManager.Shutdown()
			_, _, _ = postQuitMessage.Call(0)
			return 0
		case WmEndsession:
			if wParam != 0 {
				sdManager.Shutdown()
				_, _, _ = postQuitMessage.Call(0)
			}
			return 0
		case WmDestroy:
			sdManager.Shutdown()
			_, _, _ = postQuitMessage.Call(0)
			return 0
		}

		ret, _, _ := defWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	})
}

func isExcluded(hwnd uintptr, window core.Window, cfg *config.Conf) bool {
	if window.Title == "" {
		return true
	}

	exclusions := cfg.Exclusions()
	if slices.Contains(exclusions, window.Title) {
		return true
	}
	if slices.Contains(exclusions, window.Exe) {
		return true
	}

	class, err := getClassName(hwnd)
	if err != nil {
		return true
	}
	if slices.Contains(exclusions, class) {
		return true
	}

	return false
}
