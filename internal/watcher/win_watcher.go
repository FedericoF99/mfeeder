//go:build windows

package watcher

import (
	"context"
	"mfeeder/internal/config"
	"slices"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                   = syscall.NewLazyDLL("User32.dll")
	enumWindows              = user32.NewProc("EnumWindows")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	isWindowVisible          = user32.NewProc("IsWindowVisible")
	getForegroundWindow      = user32.NewProc("GetForegroundWindow")
	getWindowTextW           = user32.NewProc("GetWindowTextW")
	isIconic                 = user32.NewProc("IsIconic")
	getWindow                = user32.NewProc("GetWindow")
	getWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	setWinEventHook          = user32.NewProc("SetWinEventHook")

	kernel32                  = syscall.NewLazyDLL("Kernel32.dll")
	openProcess               = kernel32.NewProc("OpenProcess")
	queryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")

	dwmapi                = syscall.NewLazyDLL("Dwmapi.dll")
	dwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
)

const (
	access = 0x1000
)

type winWatcher struct{}

func newWatcher() winWatcher { return winWatcher{} }

func (w winWatcher) Snapshot(ctx context.Context, c *config.Conf) ([]Window, error) {

	info := make([]Window, 0)
	infoPtr := unsafe.Pointer(&info)
	lParam := uintptr(infoPtr)

	cb := enumWindowsCallback(c)
	res, _, err := enumWindows.Call(cb, lParam)

	if res != 0 {
		err = nil
	}

	return info, err
}

func (w winWatcher) Watch(ctx context.Context) (<-chan WindowEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (w winWatcher) Close() error {
	//TODO implement me
	panic("implement me")
}

// hwnd is a window handle (basically a pointer to the window)
// lParam is a pointer to an application-defined value passed to the callback from EnumWindows
func enumWindowsCallback(c *config.Conf) uintptr {
	fHwnd, _, _ := getForegroundWindow.Call()

	return syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		if !isAppWindow(hwnd) {
			return 1
		}

		var pid int32 = 0
		_, _, _ = getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

		proc, _, err := openProcess.Call(access, 0, uintptr(pid))
		if proc == 0 {
			println(err.Error())
		}

		eName := make([]uint16, 256)
		eNameLen := uint32(len(eName))
		_, _, _ = queryFullProcessImageName.Call(proc, 0, uintptr(unsafe.Pointer(&eName[0])), uintptr(unsafe.Pointer(&eNameLen)))
		exeName := windows.UTF16ToString(eName)

		if slices.Contains(c.Exclusions(), exeName) {
			return 1
		}

		wName := make([]uint16, 256)
		_, _, _ = getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&wName[0])), uintptr(len(wName)))

		info := (*[]Window)(unsafe.Pointer(lParam))
		*info = append(*info, Window{
			Pid:     int(pid),
			Title:   windows.UTF16ToString(wName),
			Exe:     exeName,
			Focused: fHwnd == hwnd,
		})

		return 1
	})
}

func isAppWindow(hwnd uintptr) bool {
	isVisible, _, _ := isWindowVisible.Call(hwnd)
	if isVisible == 0 {
		return false
	}

	isIcon, _, _ := isIconic.Call(hwnd)
	if isIcon == 1 {
		return false
	}

	var clocked int32 = 0
	_, _, _ = dwmGetWindowAttribute.Call(hwnd, uintptr(14), uintptr(unsafe.Pointer(&clocked)), unsafe.Sizeof(clocked))

	if clocked != 0 {
		return false
	}

	exStyle, _, _ := getWindowLongPtrW.Call(hwnd, ^uintptr(19))
	if exStyle&0x00000080 != 0 {
		return false
	}

	owner, _, _ := getWindow.Call(hwnd, uintptr(4))
	if owner != 0 {
		return false
	}

	return true
}
