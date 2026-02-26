//go:build windows

// console_windows.go configures Windows console code pages for UTF-8 I/O.
// console_windows.go는 Windows 콘솔 코드페이지를 UTF-8로 설정합니다.
package main

import "syscall"

func init() {
	// Keep cmd output/input in UTF-8 so Korean strings are not garbled.
	const cpUTF8 = 65001
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	_, _, _ = kernel32.NewProc("SetConsoleCP").Call(cpUTF8)
	_, _, _ = kernel32.NewProc("SetConsoleOutputCP").Call(cpUTF8)
}
