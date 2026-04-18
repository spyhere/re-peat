//go:build windows

package configs

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func getSystemLocale() (string, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("GetUserDefaultLocalName")
	buf := make([]uint16, 85)

	r, _, err := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return "", err
	}
	return windows.UTF16ToString(buf), nil
}
