//go:build windows

package windows

import (
	"fmt"
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
	"slices"
	"strings"
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
	procUnhookWinEvent       = user32.NewProc("UnhookWinEvent")
	procGetMessage           = user32.NewProc("GetMessageW")
	getClassNameW            = user32.NewProc("GetClassNameW")

	kernel32                  = syscall.NewLazyDLL("Kernel32.dll")
	openProcess               = kernel32.NewProc("OpenProcess")
	queryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
	procCloseHandle           = kernel32.NewProc("CloseHandle")

	dwmapi                = syscall.NewLazyDLL("Dwmapi.dll")
	dwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
)

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

func isOpenedWindow(hwnd uintptr) bool {
	isVisible, _, _ := isWindowVisible.Call(hwnd)
	if isVisible == 0 {
		return false
	}

	isIcon, _, _ := isIconic.Call(hwnd)
	if isIcon == 1 {
		return false
	}

	return isAppWindow(hwnd)
}

func isAppWindow(hwnd uintptr) bool {
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

func getWindowInfo(hwnd uintptr, fHwnd uintptr, c *config.Conf) (core.Window, error) {
	pid := getPid(hwnd)

	proc, _, _ := openProcess.Call(access, 0, uintptr(pid))
	if proc == 0 {
		return core.Window{}, fmt.Errorf("failed to open process")
	}
	defer procCloseHandle.Call(proc)

	exeName := getExeName(proc)

	if slices.Contains(c.Exclusions(), exeName) {
		return core.Window{}, fmt.Errorf("window excluded")
	}

	windowTitle := getWindowTitle(hwnd)

	window := core.Window{
		Pid:     int(pid),
		Title:   windowTitle,
		Exe:     exeName,
		Focused: fHwnd == hwnd,
	}

	return window, nil
}

func getPid(hwnd uintptr) int32 {
	var pid int32 = 0
	_, _, _ = getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid
}

func getExeName(proc uintptr) string {
	eName := make([]uint16, 256)
	eNameLen := uint32(len(eName))
	_, _, _ = queryFullProcessImageName.Call(proc, 0, uintptr(unsafe.Pointer(&eName[0])), uintptr(unsafe.Pointer(&eNameLen)))

	name := windows.UTF16ToString(eName)
	i := strings.LastIndex(name, "\\")
	if i != -1 {
		name = name[i+1:]
	}

	return strings.TrimSuffix(name, ".exe")
}

func getWindowTitle(hwnd uintptr) string {
	wName := make([]uint16, 256)
	_, _, _ = getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&wName[0])), uintptr(len(wName)))
	return windows.UTF16ToString(wName)
}

func getForegroundHandle() uintptr {
	fHwnd, _, _ := getForegroundWindow.Call()
	return fHwnd
}

func getClassName(hwnd uintptr) string {
	wName := make([]uint16, 256)
	_, _, _ = getClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&wName[0])), uintptr(len(wName)))
	return windows.UTF16ToString(wName)

}
