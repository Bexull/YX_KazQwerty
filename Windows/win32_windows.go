package main

// Win32 functions, structures and constants, through syscall only (no cgo, no packages).

import (
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	version  = syscall.NewLazyDLL("version.dll")

	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procShowWindow               = user32.NewProc("ShowWindow")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procSetTimer                 = user32.NewProc("SetTimer")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetKeyboardLayout        = user32.NewProc("GetKeyboardLayout")
	procToUnicodeEx              = user32.NewProc("ToUnicodeEx")
	procGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	procGetKeyState              = user32.NewProc("GetKeyState")
	procSetWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx           = user32.NewProc("CallNextHookEx")
	procSendInput                = user32.NewProc("SendInput")
	procCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	procAppendMenuW              = user32.NewProc("AppendMenuW")
	procTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	procDestroyMenu              = user32.NewProc("DestroyMenu")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procGetDpiForWindow          = user32.NewProc("GetDpiForWindow")
	procGetDpiForSystem          = user32.NewProc("GetDpiForSystem")
	procSetProcessDpiAwarenessCx = user32.NewProc("SetProcessDpiAwarenessContext")
	procRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")
	procDestroyIcon              = user32.NewProc("DestroyIcon")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
	procGetGUIThreadInfo         = user32.NewProc("GetGUIThreadInfo")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procSetWinEventHook          = user32.NewProc("SetWinEventHook")
	procSendMessageW             = user32.NewProc("SendMessageW")
	procSendMessageTimeoutW      = user32.NewProc("SendMessageTimeoutW")
	procFindWindowExW            = user32.NewProc("FindWindowExW")
	procMoveWindow               = user32.NewProc("MoveWindow")
	procSetFocus                 = user32.NewProc("SetFocus")
	procLoadKeyboardLayoutW      = user32.NewProc("LoadKeyboardLayoutW")
	procActivateKeyboardLayout   = user32.NewProc("ActivateKeyboardLayout")
	procGetKeyboardLayoutList    = user32.NewProc("GetKeyboardLayoutList")
	procVkKeyScanExW             = user32.NewProc("VkKeyScanExW")
	procMapVirtualKeyExW         = user32.NewProc("MapVirtualKeyExW")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")

	procGetModuleHandleW           = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentThreadId         = kernel32.NewProc("GetCurrentThreadId")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procCreateMutexW               = kernel32.NewProc("CreateMutexW")
	procGetModuleFileNameW         = kernel32.NewProc("GetModuleFileNameW")

	procRtlGetVersion = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlGetVersion")

	procCreateFontW = gdi32.NewProc("CreateFontW")

	procShellNotifyIconW       = shell32.NewProc("Shell_NotifyIconW")
	procShellNotifyIconGetRect = shell32.NewProc("Shell_NotifyIconGetRect")
	procShellExecuteW          = shell32.NewProc("ShellExecuteW")
	procCoInitializeEx         = ole32.NewProc("CoInitializeEx")

	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegCloseKey      = advapi32.NewProc("RegCloseKey")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW   = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW  = advapi32.NewProc("RegDeleteValueW")

	procGetFileVersionInfoSizeW = version.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfoW     = version.NewProc("GetFileVersionInfoW")
	procVerQueryValueW          = version.NewProc("VerQueryValueW")
)

const (
	wmNull                   = 0x0000
	wmDestroy                = 0x0002
	wmSize                   = 0x0005
	wmActivate               = 0x0006
	wmSetText                = 0x000C
	wmGetText                = 0x000D
	wmGetTextLength          = 0x000E
	wmPaint                  = 0x000F
	wmClose                  = 0x0010
	wmSetFont                = 0x0030
	wmInputLangChangeRequest = 0x0050
	wmEraseBkgnd             = 0x0014
	wmSettingChange          = 0x001A
	wmNcCalcSize             = 0x0083
	wmNcHitTest              = 0x0084
	wmKeyDown                = 0x0100
	wmKeyUp                  = 0x0101
	wmSysKeyDown             = 0x0104
	wmSysKeyUp               = 0x0105
	wmCommand                = 0x0111
	wmTimer                  = 0x0113
	wmMouseMove              = 0x0200
	wmLButtonDown            = 0x0201
	wmLButtonUp              = 0x0202
	wmRButtonDown            = 0x0204
	wmRButtonUp              = 0x0205
	wmMButtonDown            = 0x0207
	wmMouseWheel             = 0x020A
	wmMouseLeave             = 0x02A3
	wmDpiChanged             = 0x02E0
	wmApp                    = 0x8000
	wmTrayIcon               = wmApp + 1
	wmLearned                = wmApp + 2
	wmCaptureChanged         = 0x0215

	waInactive = 0

	wsPopup            = 0x80000000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsOverlappedWindow = 0x00CF0000
	wsThickFrame       = 0x00040000
	esMultiline        = 0x0004
	esAutoVScroll      = 0x0040
	wsExTopmost        = 0x00000008
	wsExToolWindow     = 0x00000080
	wsExNoActivate     = 0x08000000

	swHide          = 0
	swShowNormal    = 1
	swShow          = 5
	swpNoZOrder     = 0x0004
	swpNoActivate   = 0x0010
	swpShowWindow   = 0x0040
	swpFrameChanged = 0x0020

	csHRedraw = 0x0002
	csVRedraw = 0x0001

	vkBack       = 0x08
	vkTab        = 0x09
	vkReturn     = 0x0D
	vkSpace      = 0x20
	vkShift      = 0x10
	vkControl    = 0x11
	vkMenu       = 0x12
	vkCapital    = 0x14
	vkEscape     = 0x1B
	vkPrior      = 0x21
	vkNext       = 0x22
	vkEnd        = 0x23
	vkHome       = 0x24
	vkLeft       = 0x25
	vkUp         = 0x26
	vkRight      = 0x27
	vkDown       = 0x28
	vkInsert     = 0x2D
	vkDelete     = 0x2E
	vkLWin       = 0x5B
	vkRWin       = 0x5C
	vkLShift     = 0xA0
	vkRShift     = 0xA1
	vkLControl   = 0xA2
	vkRControl   = 0xA3
	vkLMenu      = 0xA4
	vkRMenu      = 0xA5
	vkF1         = 0x70
	vkF24        = 0x87
	vkOem4       = 0xDB // [ on a US keyboard: х in the Russian layout
	whKeyboardLL = 13
	whMouseLL    = 14

	llkhfExtended = 0x01
	llkhfInjected = 0x10

	inputKeyboard     = 1
	keyeventfExtended = 0x0001
	keyeventfKeyUp    = 0x0002
	keyeventfUnicode  = 0x0004

	mfString       = 0x0000
	mfGrayed       = 0x0001
	mfChecked      = 0x0008
	mfPopup        = 0x0010
	mfSeparator    = 0x0800
	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100
	tpmBottomAlign = 0x0020
	tpmRightAlign  = 0x0008

	nimAdd     = 0
	nimModify  = 1
	nimDelete  = 2
	nifMessage = 0x01
	nifIcon    = 0x02
	nifTip     = 0x04

	monitorDefaultToNearest = 2

	dwmwaWindowCornerPreference = 33
	dwmwaBorderColor            = 34
	dwmwaUseImmersiveDarkMode   = 20
	dwmwcpRound                 = 2

	htClient = 1

	idcArrow = 32512

	srcCopy = 0x00CC0020

	diNormal = 0x0003

	processQueryLimitedInformation = 0x1000

	tmeLeave = 0x00000002

	hkeyCurrentUser = 0x80000001
	keyRead         = 0x20019
	keySetValue     = 0x0002
	regSz           = 1
	regDword        = 4

	smCxSmIcon = 49
)

type point struct{ X, Y int32 }

type rect struct{ Left, Top, Right, Bottom int32 }

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	_       uint32
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type kbdLLHookStruct struct {
	VkCode    uint32
	ScanCode  uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type msLLHookStruct struct {
	Pt        point
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

// keyboardInput is INPUT with its KEYBDINPUT member (40 bytes on 64-bit Windows).
type keyboardInput struct {
	Type      uint32
	_         uint32
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	_         uint32
	ExtraInfo uintptr
	_         [8]byte
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	Size            uint32
	Hwnd            uintptr
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            uintptr
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        guid
	BalloonIcon     uintptr
}

type notifyIconIdentifier struct {
	Size     uint32
	Hwnd     uintptr
	ID       uint32
	GuidItem guid
}

type guiThreadInfo struct {
	Size      uint32
	Flags     uint32
	Active    uintptr
	Focus     uintptr
	Capture   uintptr
	MenuOwner uintptr
	MoveSize  uintptr
	Caret     uintptr
	CaretRect rect
}

func utf16Ptr(text string) *uint16 {
	pointer, _ := syscall.UTF16PtrFromString(text)
	return pointer
}

// utf16PtrToString reads a NUL-terminated UTF-16 string that Windows owns.
func utf16PtrToString(pointer uintptr) string {
	if pointer == 0 {
		return ""
	}
	var units []uint16
	for i := uintptr(0); ; i += 2 {
		unit := *(*uint16)(unsafe.Pointer(pointer + i))
		if unit == 0 {
			break
		}
		units = append(units, unit)
	}
	return syscall.UTF16ToString(units)
}

func loword(value uintptr) int32 { return int32(int16(value & 0xFFFF)) }
func hiword(value uintptr) int32 { return int32(int16((value >> 16) & 0xFFFF)) }

func getModuleHandle() uintptr {
	handle, _, _ := procGetModuleHandleW.Call(0)
	return handle
}

func dpiOf(hwnd uintptr) int {
	if procGetDpiForWindow.Find() == nil && hwnd != 0 {
		if dpi, _, _ := procGetDpiForWindow.Call(hwnd); dpi != 0 {
			return int(dpi)
		}
	}
	if procGetDpiForSystem.Find() == nil {
		if dpi, _, _ := procGetDpiForSystem.Call(); dpi != 0 {
			return int(dpi)
		}
	}
	return 96
}

func keyDown(vk int) bool {
	state, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return state&0x8000 != 0
}

func capsLockOn() bool {
	state, _, _ := procGetKeyState.Call(vkCapital)
	return state&1 != 0
}

// sendKeys injects keyboard input, marked so that our own hook lets it pass.
func sendKeys(inputs []keyboardInput) {
	if len(inputs) == 0 {
		return
	}
	for i := range inputs {
		inputs[i].Type = inputKeyboard
		inputs[i].ExtraInfo = injectedMarker
	}
	procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(inputs[0]))
}

func unicodeKeys(text string) []keyboardInput {
	var inputs []keyboardInput
	for _, unit := range syscall.StringToUTF16(text) {
		if unit == 0 {
			break
		}
		inputs = append(inputs,
			keyboardInput{Scan: unit, Flags: keyeventfUnicode},
			keyboardInput{Scan: unit, Flags: keyeventfUnicode | keyeventfKeyUp})
	}
	return inputs
}

func backspaceKeys(count int) []keyboardInput {
	var inputs []keyboardInput
	for i := 0; i < count; i++ {
		inputs = append(inputs, keyboardInput{Vk: vkBack}, keyboardInput{Vk: vkBack, Flags: keyeventfKeyUp})
	}
	return inputs
}

// injectedMarker tags the keys we send ourselves ("KZKB").
const injectedMarker = 0x4B5A4B42
