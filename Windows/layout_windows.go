package main

import "unsafe"

// Sizes and offsets of the Windows structures as the x64 SDK has them. Windows reads these
// from memory as they are, so the build stops on any mismatch: each line indexes a one-element
// array with the difference, which only compiles when it is zero.
var (
	_ = [1]int{}[unsafe.Sizeof(msg{})-48]
	_ = [1]int{}[unsafe.Sizeof(wndClassEx{})-80]
	_ = [1]int{}[unsafe.Sizeof(kbdLLHookStruct{})-24]
	_ = [1]int{}[unsafe.Offsetof(kbdLLHookStruct{}.ExtraInfo)-16]
	_ = [1]int{}[unsafe.Sizeof(msLLHookStruct{})-32]
	_ = [1]int{}[unsafe.Offsetof(msLLHookStruct{}.ExtraInfo)-24]
	_ = [1]int{}[unsafe.Sizeof(keyboardInput{})-40] // INPUT
	_ = [1]int{}[unsafe.Offsetof(keyboardInput{}.Vk)-8]
	_ = [1]int{}[unsafe.Offsetof(keyboardInput{}.ExtraInfo)-24]
	_ = [1]int{}[unsafe.Sizeof(notifyIconData{})-976]
	_ = [1]int{}[unsafe.Offsetof(notifyIconData{}.Icon)-32]
	_ = [1]int{}[unsafe.Offsetof(notifyIconData{}.Tip)-40]
	_ = [1]int{}[unsafe.Offsetof(notifyIconData{}.GuidItem)-952]
	_ = [1]int{}[unsafe.Sizeof(notifyIconIdentifier{})-40]
	_ = [1]int{}[unsafe.Sizeof(guiThreadInfo{})-72]
	_ = [1]int{}[unsafe.Offsetof(guiThreadInfo{}.Focus)-16]
	_ = [1]int{}[unsafe.Sizeof(guid{})-16]
)
