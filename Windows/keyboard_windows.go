package main

// The Kazakh keyboard helpers, on a low-level keyboard hook of their own thread: right Alt +
// a key types a Kazakh letter in the Russian layout (Windows' own "Русская", so there are still
// only two languages), and Kazakh typed with Russian letters gets its proper spelling.

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"kazqwerty/kazakh"
)

const russianLanguage = 0x0419

// kazakhLetters: right Alt + the key of the Russian letter types the Kazakh one, as on the
// Mac: а→ә, и→і, н→ң, г→ғ, к→қ, о→ө, у→ү, ц→ұ, х→һ. Never Q (Й).
var kazakhLetters = map[uint32][2]string{
	'F': {"ә", "Ә"}, 'B': {"і", "І"}, 'Y': {"ң", "Ң"}, 'U': {"ғ", "Ғ"}, 'R': {"қ", "Қ"},
	'J': {"ө", "Ө"}, 'E': {"ү", "Ү"}, 'W': {"ұ", "Ұ"}, vkOem4: {"һ", "Һ"},
}

// keyboardConfig is what the hook reads, swapped whole: the hook never takes a lock or waits
// for the disk, or Windows would take it for hung and drop it.
type keyboardConfig struct {
	rightAlt    bool
	autocorrect bool
	excluded    map[string]bool // exe names, lowercase
}

type keyboard struct {
	config     atomic.Pointer[keyboardConfig]
	vocabulary atomic.Pointer[kazakh.Vocabulary]
	window     uintptr // the app's message window, told about learned words
	ownPID     uint32

	learnedMu sync.Mutex
	learned   []string

	// Diagnostics.
	keys, letters, replaced atomic.Int64

	// Touched only on the hook thread.
	typing         kazakh.Typing
	foreground     uintptr
	foregroundExe  string
	rightAltHeld   bool
	rightAltPassed bool
	keptKeys       map[uint32]bool // pressed, but the app never saw it: its release is kept too
	pressed        map[uint32]bool // down as far as the app knows
}

func newKeyboard(window uintptr) *keyboard {
	k := &keyboard{window: window, keptKeys: map[uint32]bool{}, pressed: map[uint32]bool{}}
	k.ownPID = uint32(syscall.Getpid())
	return k
}

// takeLearned: unknown Kazakh words typed since the last call (UI thread).
func (k *keyboard) takeLearned() []string {
	k.learnedMu.Lock()
	defer k.learnedMu.Unlock()
	words := k.learned
	k.learned = nil
	return words
}

// run installs the hooks and pumps their messages; call it on a goroutine of its own. ready, if
// not nil, is told whether the hooks went in.
func (k *keyboard) run(ready chan<- bool) {
	runtime.LockOSThread()
	keyboardProc := syscall.NewCallback(func(code, wParam, lParam uintptr) uintptr {
		if int32(code) >= 0 && k.onKey(wParam, (*kbdLLHookStruct)(unsafe.Pointer(lParam))) {
			return 1
		}
		next, _, _ := procCallNextHookEx.Call(0, code, wParam, lParam)
		return next
	})
	mouseProc := syscall.NewCallback(func(code, wParam, lParam uintptr) uintptr {
		if int32(code) >= 0 {
			switch wParam {
			case wmLButtonDown, wmRButtonDown, wmMButtonDown:
				k.typing.Reset() // a click may move the cursor
			}
		}
		next, _, _ := procCallNextHookEx.Call(0, code, wParam, lParam)
		return next
	})
	module := getModuleHandle()
	keyboardHook, _, _ := procSetWindowsHookExW.Call(whKeyboardLL, keyboardProc, module, 0)
	mouseHook, _, _ := procSetWindowsHookExW.Call(whMouseLL, mouseProc, module, 0)
	if ready != nil {
		ready <- keyboardHook != 0 && mouseHook != 0
	}
	var m msg
	for {
		result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(result) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// onKey reports whether to keep the key from the app.
func (k *keyboard) onKey(message uintptr, info *kbdLLHookStruct) bool {
	if info.ExtraInfo == injectedMarker {
		return false // our own keys
	}
	down := message == wmKeyDown || message == wmSysKeyDown
	vk := info.VkCode
	if k.keptKeys[vk] {
		if !down {
			delete(k.keptKeys, vk)
			return true // we kept its press, so its release goes too
		}
		if !k.rightAltHeld {
			return true // it repeats while held, but right Alt is up now
		}
	}
	keep := k.handleKey(vk, down, info)
	if down && !keep {
		k.pressed[vk] = true
	} else if !down {
		delete(k.pressed, vk)
	}
	return keep
}

func (k *keyboard) handleKey(vk uint32, down bool, info *kbdLLHookStruct) bool {
	if down {
		k.keys.Add(1)
	}
	config := k.config.Load()
	if config == nil {
		return false
	}

	window, _, _ := procGetForegroundWindow.Call()
	var pid uint32
	thread, _, _ := procGetWindowThreadProcessId.Call(window, uintptr(unsafe.Pointer(&pid)))
	focus := focusOf(thread)
	if focus != 0 {
		// In Store apps the window in front is ApplicationFrameHost's; the app's own one has the
		// focus, and its layout.
		var focusPID uint32
		if focusThread, _, _ := procGetWindowThreadProcessId.Call(focus, uintptr(unsafe.Pointer(&focusPID))); focusThread != 0 {
			thread, pid = focusThread, focusPID
		}
	}
	layout, _, _ := procGetKeyboardLayout.Call(thread)
	if window != k.foreground {
		k.foreground = window
		k.foregroundExe = exeName(pid) // not cached: Windows reuses process IDs
		k.typing.Reset()
	}
	if layout&0xFFFF != russianLanguage || pid == k.ownPID {
		if vk == vkRMenu && !down && k.rightAltHeld {
			k.rightAltHeld = false
			return !k.rightAltPassed
		}
		k.typing.Reset()
		return false
	}

	if config.rightAlt {
		if keep, handled := k.rightAlt(vk, info, down, config); handled {
			return keep
		}
	}
	if !down {
		return false
	}
	if !config.autocorrect || config.excluded[k.foregroundExe] || passwordFocused(focus) {
		k.typing.Reset()
		return false
	}
	return k.autocorrect(vk, info, layout)
}

// rightAlt: right Alt + key types a Kazakh letter. handled reports whether the key was ours.
func (k *keyboard) rightAlt(vk uint32, info *kbdLLHookStruct, down bool, config *keyboardConfig) (keep, handled bool) {
	if vk == vkLControl && info.ScanCode == 0x21D {
		return true, true // the Ctrl that layouts with AltGr send along with right Alt
	}
	if vk == vkRMenu {
		if down {
			if !k.rightAltHeld {
				k.rightAltHeld, k.rightAltPassed = true, false
			}
			return !k.rightAltPassed, true
		}
		k.rightAltHeld = false
		return !k.rightAltPassed, true // the app saw a press only if we passed Alt on
	}
	if !k.rightAltHeld || k.rightAltPassed || !down {
		return false, false
	}
	switch vk {
	case vkShift, vkLShift, vkRShift, vkCapital:
		return false, true // Shift for a capital letter, pressed after right Alt: Alt stays ours
	}
	if letter, ok := kazakhLetters[vk]; ok {
		text := letter[0]
		if keyDown(vkShift) != capsLockOn() {
			text = letter[1]
		}
		sendKeys(unicodeKeys(text))
		if !k.pressed[vk] {
			k.keptKeys[vk] = true
		}
		k.letters.Add(1)
		if config.autocorrect && !config.excluded[k.foregroundExe] {
			if vocabulary := k.vocabulary.Load(); vocabulary != nil {
				k.typing.Text(text, vocabulary)
			}
		}
		return true, true
	}
	// Right Alt with another key (Alt+Tab, Alt+F4): hand Alt to the system after all. The key is
	// kept and sent again right after Alt, so the app can't get them in the wrong order.
	sendKeys([]keyboardInput{{Vk: vkRMenu, Flags: keyeventfExtended}, sameKey(info)})
	k.rightAltPassed = true
	k.typing.Reset()
	return true, true
}

func (k *keyboard) autocorrect(vk uint32, info *kbdLLHookStruct, layout uintptr) bool {
	vocabulary := k.vocabulary.Load()
	if vocabulary == nil {
		return false
	}
	switch vk {
	case vkShift, vkLShift, vkRShift, vkCapital, vkControl, vkLControl, vkRControl, vkMenu, vkLMenu, vkRMenu,
		vkLWin, vkRWin:
		return false
	case vkLeft, vkRight, vkUp, vkDown, vkHome, vkEnd, vkPrior, vkNext, vkDelete, vkInsert, vkEscape:
		k.typing.Reset() // the cursor moved
		return false
	}
	if keyDown(vkControl) || keyDown(vkMenu) || keyDown(vkLWin) || keyDown(vkRWin) || (vk >= vkF1 && vk <= vkF24) {
		k.typing.Reset() // a shortcut
		return false
	}
	if vk == vkBack {
		if action := k.typing.Backspace(); action.Kind == kazakh.Undo {
			k.apply(action, nil)
			return true
		}
		return false
	}
	text := translateKey(vk, info.ScanCode, layout)
	if text == "" || isControlText(text) {
		k.typing.Reset()
		return false
	}
	action := k.typing.Text(text, vocabulary)
	if words := k.typing.TakeLearned(); len(words) > 0 {
		k.learnedMu.Lock()
		k.learned = append(k.learned, words...)
		k.learnedMu.Unlock()
		procPostMessageW.Call(k.window, wmLearned, 0, 0)
	}
	if action.Kind != kazakh.Replace {
		return false
	}
	k.replaced.Add(1)
	k.apply(action, info)
	return true
}

// apply deletes and types for a replacement or an undo; for a replacement the key that
// ended the word is typed again after it.
func (k *keyboard) apply(action kazakh.Action, key *kbdLLHookStruct) {
	inputs := append(backspaceKeys(action.Delete), unicodeKeys(action.Insert)...)
	if key != nil {
		inputs = append(inputs, sameKey(key))
	}
	sendKeys(inputs)
}

// sameKey presses the key of a hook event again.
func sameKey(key *kbdLLHookStruct) keyboardInput {
	input := keyboardInput{Vk: uint16(key.VkCode), Scan: uint16(key.ScanCode)}
	if key.Flags&llkhfExtended != 0 {
		input.Flags = keyeventfExtended
	}
	return input
}

// translateKey: what the key types in the given layout with the current Shift / Caps Lock.
func translateKey(vk, scan uint32, layout uintptr) string {
	var state [256]byte
	if keyDown(vkShift) {
		state[vkShift], state[vkLShift] = 0x80, 0x80
	}
	if capsLockOn() {
		state[vkCapital] = 0x01
	}
	var buffer [8]uint16
	// Flag 4: don't touch the keyboard state (dead keys) of this thread.
	count, _, _ := procToUnicodeEx.Call(uintptr(vk), uintptr(scan), uintptr(unsafe.Pointer(&state[0])),
		uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 4, layout)
	if int32(count) <= 0 {
		return ""
	}
	return syscall.UTF16ToString(buffer[:count])
}

func isControlText(text string) bool {
	for _, letter := range text {
		if letter < 0x20 && letter != '\r' && letter != '\n' && letter != '\t' {
			return true
		}
	}
	return false
}

// focusOf: the window with the keyboard focus in a GUI thread, 0 if unknown.
func focusOf(thread uintptr) uintptr {
	var info guiThreadInfo
	info.Size = uint32(unsafe.Sizeof(info))
	if ok, _, _ := procGetGUIThreadInfo.Call(thread, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return 0
	}
	return info.Focus
}

// passwordFocused: the focus is a classic password box (browsers' ones can't be told).
func passwordFocused(focus uintptr) bool {
	if focus == 0 {
		return false
	}
	var class [64]uint16
	length, _, _ := procGetClassNameW.Call(focus, uintptr(unsafe.Pointer(&class[0])), uintptr(len(class)))
	if !strings.Contains(strings.ToLower(syscall.UTF16ToString(class[:length])), "edit") {
		return false
	}
	gwlStyle := -16
	style, _, _ := procGetWindowLongPtrW.Call(focus, uintptr(gwlStyle))
	return style&0x20 != 0 // ES_PASSWORD
}
