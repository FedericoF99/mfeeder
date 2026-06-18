//go:build windows

package windows

import (
	"mfeeder/internal/watcher/core"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

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
	ret, _, _ := dwmGetWindowAttribute.Call(hwnd, uintptr(14), uintptr(unsafe.Pointer(&clocked)), unsafe.Sizeof(clocked))
	if ret != 0 {
		return false
	}

	if clocked != 0 {
		return false
	}

	exStyle, _, _ := getWindowLongPtrW.Call(hwnd, ^uintptr(19))
	if exStyle&0x00000080 != 0 || exStyle == 0 {
		return false
	}

	owner, _, _ := getWindow.Call(hwnd, uintptr(4))
	if owner != 0 {
		return false
	}

	return true
}

func getWindowInfo(hwnd uintptr, fHwnd uintptr) (core.Window, error) {
	pid, err := getPid(hwnd)
	if err != nil {
		return core.Window{}, err
	}

	proc, _, err := openProcess.Call(access, 0, uintptr(pid))
	if proc == 0 {
		return core.Window{}, err
	}
	defer procCloseHandle.Call(proc)

	exeName, err := getExeName(proc)
	if err != nil {
		return core.Window{}, err
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

func getPid(hwnd uintptr) (int32, error) {
	var pid int32 = 0
	ret, _, err := getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if ret == 0 {
		return pid, err
	}
	return pid, nil
}

func getExeName(proc uintptr) (string, error) {
	eName := make([]uint16, 256)
	eNameLen := uint32(len(eName))
	ret, _, err := queryFullProcessImageName.Call(proc, 0, uintptr(unsafe.Pointer(&eName[0])), uintptr(unsafe.Pointer(&eNameLen)))
	if ret == 0 {
		return "", err
	}

	name := windows.UTF16ToString(eName)
	i := strings.LastIndex(name, "\\")
	if i != -1 {
		name = name[i+1:]
	}

	return strings.TrimSuffix(name, ".exe"), nil
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

func getClassName(hwnd uintptr) (string, error) {
	wName := make([]uint16, 256)
	ret, _, err := getClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&wName[0])), uintptr(len(wName)))
	if int32(ret) == 0 {
		return "", err
	}
	return windows.UTF16ToString(wName), nil

}
