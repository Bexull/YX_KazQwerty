package main

// Checks on a real Windows PC, since the app is built and tested on a Mac:
//
//	YX_KazQwerty.exe --selftest [folder]   checks what it can by itself and writes selftest.txt
//
// The keyboard check opens a window of its own (YX_KazQwerty.exe --selftest-editor), types into
// it like a person would, with the hook of this process in between, and reads back what appeared.

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unsafe"

	"kazqwerty/kazakh"
)

func folderArgument() string {
	folder := "."
	if len(os.Args) > 2 {
		folder = os.Args[2]
	}
	if absolute, err := filepath.Abs(folder); err == nil {
		folder = absolute
	}
	os.MkdirAll(folder, 0o755)
	return folder
}

// selfTest collects the report.
type selfTest struct {
	out                     io.Writer
	report                  strings.Builder
	passed, failed, skipped int
}

func (t *selfTest) print(format string, args ...any) {
	line := fmt.Sprintf(format, args...) + "\n"
	io.WriteString(t.out, line)
	t.report.WriteString(line)
}

func (t *selfTest) result(ok bool, name, detail string) {
	mark := "[OK]"
	if ok {
		t.passed++
	} else {
		t.failed++
		mark = "[ОШИБКА]"
	}
	t.print("%s %s: %s", mark, name, detail)
}

func (t *selfTest) skip(name, why string) {
	t.skipped++
	t.print("[ПРОПУСК] %s: %s", name, why)
}

func runSelfTest(folder string) int {
	t := &selfTest{out: io.Discard}
	// From Git Bash or with the output redirected the report goes to stdout as well; otherwise
	// (a console, Explorer) it opens in Notepad at the end.
	toNotepad := true
	if handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE); err == nil && handle != 0 &&
		handle != syscall.InvalidHandle {
		if kind, _ := syscall.GetFileType(handle); kind == syscall.FILE_TYPE_PIPE || kind == syscall.FILE_TYPE_DISK {
			t.out, toNotepad = os.Stdout, false
		}
	}
	t.print("YX_KazQwerty для Windows, самопроверка %s", time.Now().Format("02.01.2006 15:04"))
	t.system()
	vocabulary := t.dictionary()
	t.tray()
	t.keyboard(vocabulary)
	t.print("")
	t.print("Итого: прошло %d, ошибок %d, пропущено %d.", t.passed, t.failed, t.skipped)

	path := filepath.Join(folder, "selftest.txt")
	text := "\ufeff" + strings.ReplaceAll(t.report.String(), "\n", "\r\n")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.print("Не удалось записать %s: %v", path, err)
	} else if toNotepad {
		procShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("open"))), uintptr(unsafe.Pointer(utf16Ptr(path))), 0, 0,
			swShowNormal)
	}
	if t.failed > 0 {
		return 1
	}
	return 0
}

// MARK: - The system

func (t *selfTest) system() {
	var version struct {
		Size, Major, Minor, Build, Platform uint32
		ServicePack                         [128]uint16
	}
	version.Size = uint32(unsafe.Sizeof(version))
	procRtlGetVersion.Call(uintptr(unsafe.Pointer(&version)))
	name := "Windows 10"
	if version.Build >= 22000 {
		name = "Windows 11"
	}
	theme := func(value string) string {
		if systemUsesDarkTheme(value) {
			return "тёмная"
		}
		return "светлая"
	}
	t.print("%s (%d.%d.%d), масштаб %d%%, тема окон %s, панели задач %s", name, version.Major, version.Minor,
		version.Build, dpiOf(0)*100/96, theme("AppsUseLightTheme"), theme("SystemUsesLightTheme"))

	var layouts []string
	for _, layout := range keyboardLayouts() {
		text := fmt.Sprintf("%04X", layout&0xFFFF)
		switch layout & 0xFFFF {
		case 0x0409:
			text += " английская"
		case russianLanguage:
			text += " русская"
		case 0x043F:
			text += " казахская"
		}
		layouts = append(layouts, text)
	}
	autostart := "выключен"
	if launchAtLogin() {
		autostart = "включён"
	}
	t.print("Раскладки: %s. Автозапуск %s. Настройки: %s", strings.Join(layouts, ", "), autostart, newStore().folder)
}

func keyboardLayouts() []uintptr {
	count, _, _ := procGetKeyboardLayoutList.Call(0, 0)
	if count == 0 {
		return nil
	}
	layouts := make([]uintptr, count)
	count, _, _ = procGetKeyboardLayoutList.Call(count, uintptr(unsafe.Pointer(&layouts[0])))
	return layouts[:count]
}

// layoutFor: the installed keyboard layout of a language, 0 if there is none.
func layoutFor(language uintptr) uintptr {
	for _, layout := range keyboardLayouts() {
		if layout&0xFFFF == language {
			return layout
		}
	}
	return 0
}

func (t *selfTest) dictionary() *kazakh.Vocabulary {
	folder, err := fs.Sub(data, "data/kazakh")
	var vocabulary *kazakh.Vocabulary
	if err == nil {
		vocabulary, err = kazakh.Load(folder)
	}
	if err != nil {
		t.result(false, "Словарь", err.Error())
		return nil
	}
	vocabulary = vocabulary.WithWords(kazakh.Dictionary("", true))
	got := screenAfter("сиздерди шакырамыз ", vocabulary)
	t.result(got == "сіздерді шақырамыз ", "Словарь", fmt.Sprintf("%d написаний, «сиздерди шакырамыз» → «%s»",
		vocabulary.CorpusSize(), strings.TrimSpace(got)))
	return vocabulary
}

func (t *selfTest) tray() {
	window := plainWindow("YXKazQwertySelfTestTray")
	defer procDestroyWindow.Call(window)
	icon := trayGlyph(systemUsesDarkTheme("SystemUsesLightTheme"))
	defer procDestroyIcon.Call(icon)
	data := notifyIconData{Hwnd: window, ID: trayID, Flags: nifIcon | nifTip, Icon: icon}
	data.Size = uint32(unsafe.Sizeof(data))
	copy(data.Tip[:], syscall.StringToUTF16(appName+": самопроверка"))
	if added, _, _ := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&data))); added == 0 {
		t.result(false, "Значок в трее", "Windows не приняла значок (Shell_NotifyIcon)")
		return
	}
	id := notifyIconIdentifier{Hwnd: window, ID: trayID}
	id.Size = uint32(unsafe.Sizeof(id))
	var bounds rect
	located, _, _ := procShellNotifyIconGetRect.Call(uintptr(unsafe.Pointer(&id)), uintptr(unsafe.Pointer(&bounds)))
	removed, _, _ := procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
	detail := "появился и убран"
	if located == 0 {
		detail += fmt.Sprintf(", размер на экране %d×%d", bounds.Right-bounds.Left, bounds.Bottom-bounds.Top)
	} else {
		detail += ", но Windows не сказала, где он"
	}
	t.result(removed != 0, "Значок в трее", detail)
}

// plainWindow is a hidden window that only does the default, to own a tray icon.
func plainWindow(className string) uintptr {
	class := wndClassEx{WndProc: procDefWindowProcW.Addr(), Instance: getModuleHandle(), ClassName: utf16Ptr(className)}
	class.Size = uint32(unsafe.Sizeof(class))
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))
	window, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(class.ClassName)), 0, 0, 0, 0, 0, 0, 0, 0,
		getModuleHandle(), 0)
	return window
}

// MARK: - Keyboard

const editorClass = "YXKazQwertySelfTestEditor"

// keyCases: what the keyboard check types. ⌫ is backspace, ⇪ Caps Lock; Kazakh letters are
// typed with right Alt (and Shift after it for capitals). What should appear comes from the
// engine itself, except that Russian text must stay as typed.
var keyCases = []struct {
	name, keys string
	unchanged  bool
}{
	{"казахские буквы", "сиздерди шакырамыз ", false},
	{"частица отдельно", "Салеметсизбе ", false},
	{"первый ⌫ стирает только пробел", "Салеметсизбе ⌫!", false},
	{"отмена двумя ⌫", "Салеметсизбе ⌫⌫", false},
	{"го → ғой", "келе жатамго ", false},
	{"русский текст не трогается", "Привет, как дела? ", true},
	{"правый Alt и Shift", "Қазақ тілі ", false},
	{"Caps Lock", "⇪салеметсизбе ә⇪", false},
}

func (t *selfTest) keyboard(vocabulary *kazakh.Vocabulary) {
	const name = "Клавиатура"
	switch {
	case vocabulary == nil:
		t.skip(name, "словарь не загрузился")
		return
	case layoutFor(russianLanguage) == 0:
		t.skip(name, "нет русской раскладки: добавьте русский язык в «Параметры → Время и язык»")
		return
	case alreadyRunning():
		t.skip(name, "YX_KazQwerty запущен и помешает: выйдите из него (значок в трее → «Выйти») и запустите проверку снова")
		return
	}
	hook := newKeyboard(0)
	hook.config.Store(&keyboardConfig{rightAlt: true, autocorrect: true})
	hook.vocabulary.Store(vocabulary)
	ready := make(chan bool, 1)
	go hook.run(ready)
	if !<-ready {
		t.result(false, name, "Windows не дала перехватить клавиатуру (SetWindowsHookEx)")
		return
	}

	editor := exec.Command(exePath(), "--selftest-editor")
	if err := editor.Start(); err != nil {
		t.result(false, name, "окно проверки не открылось: "+err.Error())
		return
	}
	window, edit := findEditor(editor.Process.Pid)
	defer closeEditor(editor, window)
	if window == 0 {
		t.result(false, name, "окно проверки не открылось")
		return
	}
	if !bringToFront(window) {
		t.result(false, name, "окно проверки не вышло на передний план: запустите проверку ещё раз и не трогайте мышь")
		return
	}
	// Windows 10 and 11 keep one input language for all apps by default and give it to the window
	// in front, so the editor asks for Russian the way Win+Space would, now that it is in front.
	procPostMessageW.Call(window, wmInputLangChangeRequest, 0, layoutFor(russianLanguage))
	thread, _, _ := procGetWindowThreadProcessId.Call(window, 0)
	layout := uintptr(0)
	for deadline := time.Now().Add(3 * time.Second); layout == 0 && time.Now().Before(deadline); {
		if current, _, _ := procGetKeyboardLayout.Call(thread); current&0xFFFF == russianLanguage {
			layout = current
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if layout == 0 {
		t.result(false, name, "в окне проверки не включилась русская раскладка")
		return
	}

	keys := &typist{window: window, layout: layout, letters: map[rune]uint32{}}
	for vk, letter := range kazakhLetters {
		keys.letters[[]rune(letter[0])[0]] = vk
	}
	defer keys.restoreCapsLock()
	for _, c := range keyCases {
		setText(edit, "")
		keys.tap(vkEnd, false) // the cursor "moved": the autocorrect starts afresh
		if !keys.typeKeys(c.keys) {
			t.result(false, name+", "+c.name, "окно проверки перестало быть активным, печать остановлена")
			return
		}
		time.Sleep(400 * time.Millisecond)
		got, want := getText(edit), screenAfter(c.keys, vocabulary)
		if c.unchanged {
			want = c.keys
		}
		detail := fmt.Sprintf("«%s» → «%s»", c.keys, got)
		if got != want {
			detail += fmt.Sprintf(", а должно быть «%s»", want)
		}
		t.result(got == want, name+", "+c.name, detail)
	}
}

// screenAfter is what typing keys should leave on screen by the engine alone, the way the
// hook applies its answers.
func screenAfter(keys string, vocabulary *kazakh.Vocabulary) string {
	var typing kazakh.Typing
	var screen []rune
	replace := func(count int, text string) {
		screen = append(screen[:max(0, len(screen)-count)], []rune(text)...)
	}
	caps := false
	for _, key := range keys {
		switch key {
		case '⇪':
			caps = !caps
		case '⌫':
			if action := typing.Backspace(); action.Kind == kazakh.Undo {
				replace(action.Delete, action.Insert)
			} else {
				replace(1, "")
			}
		default:
			if caps && unicode.IsUpper(key) {
				key = unicode.ToLower(key)
			} else if caps {
				key = unicode.ToUpper(key)
			}
			if action := typing.Text(string(key), vocabulary); action.Kind == kazakh.Replace {
				replace(action.Delete, action.Insert)
			}
			screen = append(screen, key)
		}
	}
	return string(screen)
}

func findEditor(pid int) (window, edit uintptr) {
	class := uintptr(unsafe.Pointer(utf16Ptr(editorClass)))
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		for window, _, _ = procFindWindowExW.Call(0, 0, class, 0); window != 0; window, _, _ =
			procFindWindowExW.Call(0, window, class, 0) {
			var owner uint32
			procGetWindowThreadProcessId.Call(window, uintptr(unsafe.Pointer(&owner)))
			if int(owner) != pid {
				continue
			}
			if edit, _, _ = procFindWindowExW.Call(window, 0, uintptr(unsafe.Pointer(utf16Ptr("Edit"))), 0); edit != 0 {
				return window, edit
			}
		}
	}
	return 0, 0
}

func closeEditor(editor *exec.Cmd, window uintptr) {
	if window != 0 {
		procPostMessageW.Call(window, wmClose, 0, 0)
	}
	done := make(chan struct{})
	go func() {
		editor.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		editor.Process.Kill()
	}
}

func foreground() uintptr {
	window, _, _ := procGetForegroundWindow.Call()
	return window
}

// bringToFront waits for the editor to come to the front. Windows may refuse it that; then
// this process asks for it, sharing input with the window in front for a moment.
func bringToFront(window uintptr) bool {
	waitFor := func(timeout time.Duration) bool {
		for deadline := time.Now().Add(timeout); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
			if foreground() == window {
				return true
			}
		}
		return false
	}
	if waitFor(2 * time.Second) {
		return true
	}
	frontThread, _, _ := procGetWindowThreadProcessId.Call(foreground(), 0)
	ownThread, _, _ := procGetCurrentThreadId.Call()
	procAttachThreadInput.Call(ownThread, frontThread, 1)
	procSetForegroundWindow.Call(window)
	procAttachThreadInput.Call(ownThread, frontThread, 0)
	return waitFor(2 * time.Second)
}

func setText(edit uintptr, text string) {
	sendTo(edit, wmSetText, 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getText(edit uintptr) string {
	length := sendTo(edit, wmGetTextLength, 0, 0)
	buffer := make([]uint16, length+1)
	sendTo(edit, wmGetText, length+1, uintptr(unsafe.Pointer(&buffer[0])))
	return syscall.UTF16ToString(buffer)
}

// sendTo sends a message to the editor, which is another process: it gives up after two
// seconds if that process hangs, so the check can't hang with it.
//
//go:uintptrescapes
func sendTo(window, message, wParam, lParam uintptr) uintptr {
	var result uintptr
	procSendMessageTimeoutW.Call(window, message, wParam, lParam, 0x0002 /* SMTO_ABORTIFHUNG */, 2000,
		uintptr(unsafe.Pointer(&result)))
	return result
}

// typist types into the editor like a person, through the hooks (its keys carry no marker).
// Keys go down only while the editor is in front, and every key pressed is released.
type typist struct {
	window, layout uintptr
	letters        map[rune]uint32 // Kazakh letter -> the key that types it with right Alt
	capsToggles    int
}

func (k *typist) send(vk uint16, up bool) {
	input := keyboardInput{Type: inputKeyboard, Vk: vk}
	scan, _, _ := procMapVirtualKeyExW.Call(uintptr(vk), 0, k.layout) // MAPVK_VK_TO_VSC
	input.Scan = uint16(scan)
	switch vk {
	case vkRMenu, vkRControl, vkPrior, vkNext, vkEnd, vkHome, vkLeft, vkUp, vkRight, vkDown, vkInsert, vkDelete:
		input.Flags |= keyeventfExtended
	}
	if up {
		input.Flags |= keyeventfKeyUp
	}
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))
	time.Sleep(15 * time.Millisecond)
}

func (k *typist) down(vk uint16) bool {
	if foreground() != k.window {
		return false
	}
	k.send(vk, false)
	return true
}

func (k *typist) tap(vk uint16, shift bool) bool {
	if shift && !k.down(vkLShift) {
		return false
	}
	ok := k.down(vk)
	if ok {
		k.send(vk, true)
		if vk == vkCapital {
			k.capsToggles++
		}
	}
	if shift {
		k.send(vkLShift, true)
	}
	return ok
}

// typeKeys types text in the editor's layout; see keyCases for ⌫, ⇪ and Kazakh letters.
func (k *typist) typeKeys(text string) bool {
	for _, key := range text {
		ok := false
		vk, isKazakh := k.letters[unicode.ToLower(key)]
		switch {
		case key == '⌫':
			ok = k.tap(vkBack, false)
		case key == '⇪':
			ok = k.tap(vkCapital, false)
		case isKazakh:
			if ok = k.down(vkRMenu); ok {
				ok = k.tap(uint16(vk), unicode.IsUpper(key))
				k.send(vkRMenu, true)
			}
		default:
			scan, _, _ := procVkKeyScanExW.Call(uintptr(key), k.layout)
			if int16(scan) == -1 {
				return false
			}
			ok = k.tap(uint16(scan&0xFF), scan&0x100 != 0)
		}
		if !ok {
			return false
		}
	}
	return true
}

func (k *typist) restoreCapsLock() {
	if k.capsToggles%2 == 1 {
		k.send(vkCapital, false)
		k.send(vkCapital, true)
	}
}

// runSelfTestEditor is the window the keyboard check types into: a text box in the Russian
// layout, in a process of its own (the hook leaves the app's own windows alone). It gives the
// user's layout back when it closes, and closes by itself if the check never comes back.
func runSelfTestEditor() {
	previous, _, _ := procGetKeyboardLayout.Call(0)
	russian, _, _ := procLoadKeyboardLayoutW.Call(uintptr(unsafe.Pointer(utf16Ptr("00000419"))), 0)
	var edit uintptr
	proc := syscall.NewCallback(func(window, message, wParam, lParam uintptr) uintptr {
		switch message {
		case wmActivate:
			if loword(wParam) != waInactive {
				procActivateKeyboardLayout.Call(russian, 0)
				procSetFocus.Call(edit)
				return 0
			}
		case wmSize:
			procMoveWindow.Call(edit, 0, 0, uintptr(loword(lParam)), uintptr(hiword(lParam)), 1)
			return 0
		case wmClose, wmTimer:
			procActivateKeyboardLayout.Call(previous, 0)
			procDestroyWindow.Call(window)
			return 0
		case wmDestroy:
			procPostQuitMessage.Call(0)
			return 0
		}
		result, _, _ := procDefWindowProcW.Call(window, message, wParam, lParam)
		return result
	})
	class := wndClassEx{WndProc: proc, Instance: getModuleHandle(), ClassName: utf16Ptr(editorClass),
		Background: 6 /* COLOR_WINDOW+1 */}
	class.Size = uint32(unsafe.Sizeof(class))
	class.Cursor, _, _ = procLoadCursorW.Call(0, idcArrow)
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))

	dpi := dpiOf(0)
	width, height := 760*dpi/96, 200*dpi/96
	screenWidth, _, _ := procGetSystemMetrics.Call(0)  // SM_CXSCREEN
	screenHeight, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN
	window, _, _ := procCreateWindowExW.Call(wsExTopmost, uintptr(unsafe.Pointer(class.ClassName)),
		uintptr(unsafe.Pointer(utf16Ptr(appName+": проверка клавиатуры, не трогайте клавиатуру и мышь"))),
		wsOverlappedWindow, uintptr((int(screenWidth)-width)/2), uintptr((int(screenHeight)-height)/3),
		uintptr(width), uintptr(height), 0, 0, getModuleHandle(), 0)
	edit, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("EDIT"))), 0,
		wsChild|wsVisible|esMultiline|esAutoVScroll, 0, 0, 0, 0, window, 1, getModuleHandle(), 0)
	var client rect
	procGetClientRect.Call(window, uintptr(unsafe.Pointer(&client)))
	procMoveWindow.Call(edit, 0, 0, uintptr(client.Right), uintptr(client.Bottom), 1)
	font, _, _ := procCreateFontW.Call(uintptr(-28*dpi/96), 0, 0, 0, 400, 0, 0, 0, 1 /* DEFAULT_CHARSET */, 0, 0,
		5 /* CLEARTYPE_QUALITY */, 0, uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))))
	procSendMessageW.Call(edit, wmSetFont, font, 1)
	procSetTimer.Call(window, 1, 90_000, 0)
	procShowWindow.Call(window, swShow)
	procSetForegroundWindow.Call(window)

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
