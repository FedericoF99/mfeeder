//go:build windows

package windows

import (
	"context"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
	"runtime"
	"slices"
	"syscall"
	"unsafe"
)

const (
	access = 0x1000

	EventSystemForeground    uint32 = 0x0003
	EventSystemMinimizeStart uint32 = 0x0016
	EventSystemMinimizeEnd   uint32 = 0x0017

	EventObjectCreate  uint32 = 0x8000
	EventObjectDestroy uint32 = 0x8001
	EventObjectShow    uint32 = 0x8002
	EventObjectHide    uint32 = 0x8003
)

type RawWindowsEvent struct {
	hwnd  uintptr
	event core.WindowEventType
}

type WinWatcher struct {
	cfg *config.Conf
}

func (w WinWatcher) Snapshot(ctx context.Context) ([]core.Window, error) {

	info := make([]core.Window, 0)
	infoPtr := unsafe.Pointer(&info)
	lParam := uintptr(infoPtr)

	cb := enumWindowsCallback(ctx, w.cfg)
	res, _, err := enumWindows.Call(cb, lParam)

	if res != 0 {
		err = nil
	}

	return info, err
}

func (w WinWatcher) Watch(ctx context.Context) (<-chan core.WindowEvent, error) {

	ch := make(chan core.WindowEvent)
	chRaw := make(chan RawWindowsEvent)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case raw, ok := <-chRaw:
				if !ok {
					return
				}

				fHwnd := getForegroundHandle()
				window, err := getWindowInfo(raw.hwnd, fHwnd, w.cfg)
				if window.Title == "" {
					continue
				}

				if slices.Contains(w.cfg.Exclusions(), window.Exe) {
					continue
				}

				class := getClassName(raw.hwnd)
				if slices.Contains(w.cfg.Exclusions(), class) {
					continue
				}

				if err == nil {
					ch <- core.WindowEvent{
						Window:      window,
						WindowEvent: raw.event,
					}
				}
			}
		}
	}()

	go func() {
		runtime.LockOSThread()
		cb := eventHookCallback(chRaw)

		foregroundEvHook, _, _ := setWinEventHook.Call(uintptr(EventSystemForeground), uintptr(EventSystemForeground), uintptr(0), cb, 0, 0, uintptr(0|2))
		minimizedEvHook, _, _ := setWinEventHook.Call(uintptr(EventSystemMinimizeStart), uintptr(EventSystemMinimizeEnd), uintptr(0), cb, 0, 0, uintptr(0|2))
		objectEvHook, _, _ := setWinEventHook.Call(uintptr(EventObjectCreate), uintptr(EventObjectHide), uintptr(0), cb, 0, 0, uintptr(0|2))

		var msg MSG
		for !ctx.Done() {
			ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)

			if int32(ret) <= 0 {
				break
			}
		}

		_, _, _ = procUnhookWinEvent.Call(foregroundEvHook)
		_, _, _ = procUnhookWinEvent.Call(minimizedEvHook)
		_, _, _ = procUnhookWinEvent.Call(objectEvHook)
		runtime.UnlockOSThread()
	}()

	return ch, nil
}

func (w WinWatcher) Close() error {
	//TODO implement me
	panic("implement me")
}

// hwnd is a window handle (basically a pointer to the window)
// lParam is a pointer to an application-defined value passed to the callback from EnumWindows
func enumWindowsCallback(ctx context.Context, c *config.Conf) uintptr {
	fHwnd := getForegroundHandle()

	return syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		if ctx.Err() != nil {
			log.Println("stopping enum window callback because context done")
			return 0
		}

		if !isOpenedWindow(hwnd) {
			return 1
		}

		window, err := getWindowInfo(hwnd, fHwnd, c)
		if err != nil {
			return 0
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

		chRaw <- RawWindowsEvent{hwnd: hwnd, event: windowEvent}

		return 0
	})
}
