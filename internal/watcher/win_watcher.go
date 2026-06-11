//go:build windows

package watcher

import (
	"mfeeder/internal/config"
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("User32.dll") // dll containing EnumWindows function

	enumWindowsProc          = user32.NewProc("EnumWindows")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

type winWatcher struct{}

func (w *winWatcher) Watch(c *config.Conf) ([]Info, error) {
	config.GetExclusions()

	infoPtr := unsafe.Pointer(new(make([]Info, 0)))
	lParam := uintptr(infoPtr)

	cb := enumWindowsCallback()

	enumWindowsProc.Call(cb, lParam)

	return Info{}, nil
}

// hwnd is a window handle (basically a pointer to the window)
// lParam is a pointer to an application-defined value passed to the callback from EnumWindows
func enumWindowsCallback() uintptr {
	return syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		var pid int32 = 0
		getWindowThreadProcessId.Call(hwnd, pid)

		info := (*[]Info)(unsafe.Pointer(lParam))
		*info = append(*info, Info{})
		return 1
	})
}
