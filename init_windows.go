//go:build windows
// +build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func prepareConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")

	var mode uint32
	handle := os.Stdout.Fd()
	ret, _, _ := getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if ret != 0 {
		mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
		setConsoleMode.Call(handle, uintptr(mode))
	}
}
