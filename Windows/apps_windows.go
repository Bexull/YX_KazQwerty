package main

// Which app a process is and its name, for "don't correct in …".

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// processNames caches display names of apps by exe path.
type processNames struct {
	names map[string]string // exe path -> display name
}

func newProcessNames() *processNames {
	return &processNames{names: map[string]string{}}
}

func processPath(pid uint32) string {
	handle, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer syscall.CloseHandle(syscall.Handle(handle))
	buffer := make([]uint16, 1024)
	size := uint32(len(buffer))
	if ok, _, _ := procQueryFullProcessImageNameW.Call(handle, 0, uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size))); ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buffer[:size])
}

// displayName: the file description ("Google Chrome", "Microsoft Teams"), else the file name.
func (p *processNames) displayName(path string) string {
	if path == "" {
		return ""
	}
	if name, ok := p.names[path]; ok {
		return name
	}
	name := versionString(path, "FileDescription")
	if name == "" {
		name = versionString(path, "ProductName")
	}
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	p.names[path] = name
	return name
}

func versionString(path, field string) string {
	pathPtr := utf16Ptr(path)
	size, _, _ := procGetFileVersionInfoSizeW.Call(uintptr(unsafe.Pointer(pathPtr)), 0)
	if size == 0 {
		return ""
	}
	data := make([]byte, size)
	if ok, _, _ := procGetFileVersionInfoW.Call(uintptr(unsafe.Pointer(pathPtr)), 0, size,
		uintptr(unsafe.Pointer(&data[0]))); ok == 0 {
		return ""
	}
	var pointer uintptr
	var length uint32
	if ok, _, _ := procVerQueryValueW.Call(uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(utf16Ptr(`\VarFileInfo\Translation`))), uintptr(unsafe.Pointer(&pointer)),
		uintptr(unsafe.Pointer(&length))); ok == 0 || length < 4 {
		return ""
	}
	language := *(*uint16)(unsafe.Pointer(pointer))
	codePage := *(*uint16)(unsafe.Pointer(pointer + 2))
	query := fmt.Sprintf(`\StringFileInfo\%04x%04x\%s`, language, codePage, field)
	if ok, _, _ := procVerQueryValueW.Call(uintptr(unsafe.Pointer(&data[0])), uintptr(unsafe.Pointer(utf16Ptr(query))),
		uintptr(unsafe.Pointer(&pointer)), uintptr(unsafe.Pointer(&length))); ok == 0 || length == 0 {
		return ""
	}
	text := strings.TrimSpace(utf16PtrToString(pointer))
	runtime.KeepAlive(data) // pointer points into data
	return text
}

// exeName of the window's process, lowercase ("chrome.exe"); "" if unknown.
func exeName(pid uint32) string {
	return strings.ToLower(filepath.Base(processPath(pid)))
}
