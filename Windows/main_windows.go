// YX_KazQwerty for Windows: Kazakh letters on the right Alt in Windows' own Russian layout, and
// autocorrect for Kazakh typed with Russian letters. The twin of the Mac app in the folder above.
package main

import (
	"embed"
	"io/fs"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"unsafe"

	"kazqwerty/kazakh"
)

//go:embed data/kazakh
var data embed.FS

const (
	appName        = "YX_KazQwerty"
	trayID         = 1
	timerUserWords = 2
	commandLogin   = 1
	commandLetters = 3
	commandAutofix = 4
	commandWords   = 5
	commandExclude = 6
	commandForget  = 7
	commandQuit    = 8
)

type application struct {
	window         uintptr // hidden window for the tray and timers
	processes      *processNames
	store          *store
	keyboard       *keyboard
	base           *kazakh.Vocabulary
	loaded         atomic.Pointer[kazakh.Vocabulary]
	taskbarCreated uintptr
	trayIcon       uintptr
	lastApp        struct{ exe, name string } // the app used before the tray was clicked
}

func main() {
	runtime.LockOSThread()
	if procSetProcessDpiAwarenessCx.Find() == nil {
		procSetProcessDpiAwarenessCx.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
	}
	procCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED, for ShellExecute
	startGdiplus()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--selftest":
			os.Exit(runSelfTest(folderArgument()))
		case "--selftest-editor":
			runSelfTestEditor()
			return
		}
	}
	if alreadyRunning() {
		return
	}

	app := &application{processes: newProcessNames(), store: newStore()}
	app.createWindow()
	app.keyboard = newKeyboard(app.window)
	app.applySettings()
	go app.loadVocabulary()
	go app.keyboard.run(nil)
	app.addTrayIcon()
	app.watchForeground()
	procSetTimer.Call(app.window, timerUserWords, 2000, 0)

	var m msg
	for {
		result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(result) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	app.removeTrayIcon()
}

// alreadyRunning reports whether another YX_KazQwerty runs in this session (and from now on,
// that this one does).
func alreadyRunning() bool {
	_, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(utf16Ptr(`Local\YX_KazQwerty`))))
	errno, ok := err.(syscall.Errno)
	return ok && errno == 183 // ERROR_ALREADY_EXISTS
}

func (app *application) createWindow() {
	className := utf16Ptr("YXKazQwertyMain")
	class := wndClassEx{WndProc: syscall.NewCallback(app.windowProc), Instance: getModuleHandle(), ClassName: className}
	class.Size = uint32(unsafe.Sizeof(class))
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))
	app.window, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(appName))), 0, 0, 0, 0, 0, 0, 0, getModuleHandle(), 0)
	app.taskbarCreated, _, _ = procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(utf16Ptr("TaskbarCreated"))))
}

func (app *application) loadVocabulary() {
	folder, err := fs.Sub(data, "data/kazakh")
	if err != nil {
		return
	}
	vocabulary, err := kazakh.Load(folder)
	if err != nil {
		return
	}
	app.loaded.Store(vocabulary)
	procPostMessageW.Call(app.window, wmLearned, 1, 0)
}

// applySettings hands the current settings and word lists to the keyboard hook.
func (app *application) applySettings() {
	current := app.store.snapshot()
	app.keyboard.config.Store(&keyboardConfig{
		rightAlt:    current.RightAlt,
		autocorrect: current.Autocorrect,
		excluded:    app.store.exclusions(),
	})
	if app.base != nil {
		app.keyboard.vocabulary.Store(app.base.WithWords(app.store.words()))
	}
}

func (app *application) windowProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch {
	case message == wmTrayIcon:
		switch lParam & 0xFFFF {
		case wmLButtonUp, wmRButtonUp:
			app.showMenu()
		}
		return 0
	case message == wmLearned:
		if wParam == 1 { // the vocabulary finished loading
			app.base = app.loaded.Load()
			app.applySettings()
			return 0
		}
		if app.store.remember(app.keyboard.takeLearned()) {
			app.applySettings()
		}
		return 0
	case message == wmTimer && wParam == timerUserWords:
		if app.store.reloadUserWords() {
			app.applySettings()
		}
		return 0
	case message == wmSettingChange:
		app.updateTrayIcon()
		return 0
	case message == app.taskbarCreated && app.taskbarCreated != 0:
		app.addTrayIcon() // Explorer restarted
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return result
}

// MARK: - Tray icon

func (app *application) trayData() notifyIconData {
	data := notifyIconData{Hwnd: app.window, ID: trayID, Flags: nifMessage | nifIcon | nifTip,
		CallbackMessage: wmTrayIcon, Icon: app.trayIcon}
	data.Size = uint32(unsafe.Sizeof(data))
	copy(data.Tip[:], syscall.StringToUTF16(appName+": казахская клавиатура"))
	return data
}

func (app *application) addTrayIcon() {
	old := app.trayIcon
	app.trayIcon = trayGlyph(systemUsesDarkTheme("SystemUsesLightTheme"))
	data := app.trayData()
	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&data)))
	if old != 0 {
		procDestroyIcon.Call(old)
	}
}

func (app *application) updateTrayIcon() {
	old := app.trayIcon
	app.trayIcon = trayGlyph(systemUsesDarkTheme("SystemUsesLightTheme"))
	data := app.trayData()
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&data)))
	if old != 0 {
		procDestroyIcon.Call(old)
	}
}

func (app *application) removeTrayIcon() {
	data := app.trayData()
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
}

// trayGlyph: a keyboard, like the Mac menu bar icon; white for a dark taskbar.
func trayGlyph(darkTaskbar bool) uintptr {
	color := uint32(0xFF1B1B1B)
	if darkTaskbar {
		color = 0xFFFFFFFF
	}
	size := 16 * dpiOf(0) / 96
	return iconImage(size, func(c *canvas) {
		const left, top, right, bottom = 1.5, 3.5, 14.5, 12.5
		c.line(left, top, right, top, 1.2, color)
		c.line(right, top, right, bottom, 1.2, color)
		c.line(right, bottom, left, bottom, 1.2, color)
		c.line(left, bottom, left, top, 1.2, color)
		for _, y := range []float32{6, 8.3} {
			for _, x := range []float32{4.2, 6.7, 9.3, 11.8} {
				c.circle(x, y, 0.8, color)
			}
		}
		c.line(5.5, 10.6, 10.5, 10.6, 1.2, color)
	})
}

// MARK: - Menu

func (app *application) showMenu() {
	current := app.store.snapshot()
	menu, _, _ := procCreatePopupMenu.Call()
	if layoutFor(russianLanguage) == 0 {
		procAppendMenuW.Call(menu, mfString|mfGrayed, 0,
			uintptr(unsafe.Pointer(utf16Ptr("Добавьте русскую раскладку: Параметры → Время и язык"))))
		procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	}
	appendItem(menu, commandLetters, "Казахские буквы на правом Alt", current.RightAlt)
	appendItem(menu, commandAutofix, "Автозамена казахских слов", current.Autocorrect)
	appendItem(menu, commandWords, "Мои слова для автозамены…", false)
	if app.lastApp.exe != "" {
		appendItem(menu, commandExclude, "Не исправлять в «"+app.lastApp.name+"»", app.store.isExcluded(app.lastApp.exe))
	}
	if learned := len(current.Learned); learned > 0 {
		appendItem(menu, commandForget, "Забыть выученное ("+strconv.Itoa(learned)+")", false)
	}
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	appendItem(menu, commandLogin, "Запускать при входе в Windows", launchAtLogin())
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	appendItem(menu, commandQuit, "Выйти", false)

	var cursor point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	procSetForegroundWindow.Call(app.window)
	command, _, _ := procTrackPopupMenu.Call(menu, tpmReturnCmd|tpmRightButton|tpmBottomAlign|tpmRightAlign,
		uintptr(cursor.X), uintptr(cursor.Y), 0, app.window, 0)
	procPostMessageW.Call(app.window, wmNull, 0, 0) // so the next click on the icon opens the menu again
	procDestroyMenu.Call(menu)

	switch command {
	case commandLogin:
		setLaunchAtLogin(!launchAtLogin())
	case commandLetters:
		app.store.update(func(s *settings) { s.RightAlt = !s.RightAlt })
	case commandAutofix:
		app.store.update(func(s *settings) { s.Autocorrect = !s.Autocorrect })
	case commandWords:
		app.store.openUserWords()
	case commandExclude:
		app.store.toggleExclusion(app.lastApp.exe)
	case commandForget:
		app.store.forgetLearned()
	case commandQuit:
		procPostQuitMessage.Call(0)
	}
	app.applySettings()
}

func appendItem(menu uintptr, id int, title string, checked bool) {
	flags := uintptr(mfString)
	if checked {
		flags |= mfChecked
	}
	procAppendMenuW.Call(menu, flags, uintptr(id), uintptr(unsafe.Pointer(utf16Ptr(title))))
}

// MARK: - The app in front

// watchForeground remembers the last app used, so the menu can offer "don't correct in …"
// for it although clicking the tray brings the taskbar to the front.
func (app *application) watchForeground() {
	ignored := map[string]bool{"explorer.exe": true, "shellexperiencehost.exe": true, "startmenuexperiencehost.exe": true,
		"searchhost.exe": true, "textinputhost.exe": true, "lockapp.exe": true, "": true}
	own := uint32(os.Getpid())
	callback := syscall.NewCallback(func(hook, event, hwnd, object, child, thread, time uintptr) uintptr {
		var pid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == own {
			return 0
		}
		path := processPath(pid)
		exe := strings.ToLower(baseName(path))
		if !ignored[exe] {
			app.lastApp.exe, app.lastApp.name = exe, app.processes.displayName(path)
		}
		return 0
	})
	procSetWinEventHook.Call(3, 3, 0, callback, 0, 0, 0) // EVENT_SYSTEM_FOREGROUND, WINEVENT_OUTOFCONTEXT
}

func baseName(path string) string {
	if i := strings.LastIndexAny(path, `\/`); i >= 0 {
		return path[i+1:]
	}
	return path
}
