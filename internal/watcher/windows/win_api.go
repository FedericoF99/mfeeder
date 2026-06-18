//go:build windows

package windows

import (
	"syscall"
)

var (
	// USER32
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
	registerClassExW         = user32.NewProc("RegisterClassExW")
	createWindowExW          = user32.NewProc("CreateWindowExW")
	defWindowProcW           = user32.NewProc("DefWindowProcW")
	translateMessage         = user32.NewProc("TranslateMessage")
	dispatchMessageW         = user32.NewProc("DispatchMessageW")
	destroyWindow            = user32.NewProc("DestroyWindow")
	postQuitMessage          = user32.NewProc("PostQuitMessage")
	postMessageW             = user32.NewProc("PostMessageW")

	// KERNEL32
	kernel32                  = syscall.NewLazyDLL("Kernel32.dll")
	openProcess               = kernel32.NewProc("OpenProcess")
	queryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
	procCloseHandle           = kernel32.NewProc("CloseHandle")
	getModuleHandleW          = kernel32.NewProc("GetModuleHandleW")

	// DWM
	dwmapi                = syscall.NewLazyDLL("Dwmapi.dll")
	dwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
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

	WmClose      = 0x0010
	WmEndsession = 0x0016
	WmDestroy    = 0x0002

	WsOverlapped = 0x00000000
)

type point struct {
	x int32
	y int32
}

type msg struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

type wndclassex struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   uintptr
	icon       uintptr
	cursor     uintptr
	background uintptr
	menuName   *uint16
	className  *uint16
	iconSm     uintptr
}
