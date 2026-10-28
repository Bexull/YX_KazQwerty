package main

// Settings and the user's word lists in %APPDATA%\YX_KazQwerty, and starting with Windows.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"kazqwerty/kazakh"
)

// builtInExclusions: apps where autocorrect would get in the way: terminals, code editors,
// password managers, remote desktops.
var builtInExclusions = map[string]bool{
	"windowsterminal.exe": true, "cmd.exe": true, "powershell.exe": true, "pwsh.exe": true,
	"conhost.exe": true, "wsl.exe": true, "code.exe": true, "cursor.exe": true, "devenv.exe": true,
	"idea64.exe": true, "pycharm64.exe": true, "webstorm64.exe": true, "rider64.exe": true,
	"clion64.exe": true, "goland64.exe": true, "datagrip64.exe": true, "studio64.exe": true,
	"sublime_text.exe": true, "notepad++.exe": true, "dbeaver.exe": true, "ssms.exe": true,
	"keepass.exe": true, "keepassxc.exe": true, "1password.exe": true, "bitwarden.exe": true,
	"mstsc.exe": true, "putty.exe": true, "winscp.exe": true, "blender.exe": true,
}

type settings struct {
	RightAlt    bool              `json:"rightAlt"`
	Autocorrect bool              `json:"autocorrect"`
	Excluded    []string          `json:"excluded,omitempty"` // exe names, lowercase
	Included    []string          `json:"included,omitempty"` // built-in exclusions turned back on
	Learned     map[string]string `json:"learned,omitempty"`  // plain spelling -> word
	Seen        map[string]int    `json:"seen,omitempty"`     // unknown Kazakh word -> times typed
}

type store struct {
	mu        sync.Mutex
	settings  settings
	folder    string
	wordsTime time.Time
	userWords string
}

func newStore() *store {
	folder, err := os.UserConfigDir()
	if err != nil {
		folder = os.TempDir()
	}
	s := &store{folder: filepath.Join(folder, appName), settings: settings{RightAlt: true, Autocorrect: true}}
	if data, err := os.ReadFile(s.path()); err == nil {
		json.Unmarshal(data, &s.settings)
	}
	return s
}

func (s *store) path() string { return filepath.Join(s.folder, "settings.json") }

// WordsPath is the user's own word list.
func (s *store) wordsPath() string { return filepath.Join(s.folder, "kazakh-words.txt") }

// update changes the settings and saves them; the file is written outside the lock.
func (s *store) update(change func(*settings)) {
	s.mu.Lock()
	change(&s.settings)
	data, _ := json.MarshalIndent(s.settings, "", "  ")
	s.mu.Unlock()
	os.MkdirAll(s.folder, 0o755)
	os.WriteFile(s.path(), data, 0o644)
}

func (s *store) snapshot() settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

func (s *store) isExcluded(exe string) bool {
	current := s.snapshot()
	if contains(current.Excluded, exe) {
		return true
	}
	return builtInExclusions[exe] && !contains(current.Included, exe)
}

// exclusions: every app autocorrect stays out of, for the keyboard hook.
func (s *store) exclusions() map[string]bool {
	current := s.snapshot()
	set := map[string]bool{}
	for exe := range builtInExclusions {
		if !contains(current.Included, exe) {
			set[exe] = true
		}
	}
	for _, exe := range current.Excluded {
		set[exe] = true
	}
	return set
}

func (s *store) toggleExclusion(exe string) {
	excluded := s.isExcluded(exe)
	s.update(func(current *settings) {
		current.Excluded = without(current.Excluded, exe)
		current.Included = without(current.Included, exe)
		if !excluded {
			current.Excluded = append(current.Excluded, exe)
		} else if builtInExclusions[exe] {
			current.Included = append(current.Included, exe)
		}
	})
}

// remember counts unknown Kazakh words; one typed twice is learned. Reports whether the word
// list changed.
func (s *store) remember(words []string) bool {
	changed := false
	s.update(func(current *settings) {
		if current.Seen == nil {
			current.Seen = map[string]int{}
		}
		if current.Learned == nil {
			current.Learned = map[string]string{}
		}
		for _, word := range words {
			current.Seen[word]++
			if current.Seen[word] >= 2 {
				current.Learned[kazakh.Plain(word)] = word
				delete(current.Seen, word)
				changed = true
			}
		}
	})
	return changed
}

func (s *store) forgetLearned() {
	s.update(func(current *settings) {
		current.Learned = nil
		current.Seen = nil
	})
}

// words: hand-picked, then learned, then the user's own list (which wins).
func (s *store) words() map[string]string {
	words := kazakh.Dictionary("", true)
	for plain, word := range s.snapshot().Learned {
		words[plain] = word
	}
	for plain, word := range kazakh.Dictionary(s.userWords, false) {
		words[plain] = word
	}
	return words
}

// reloadUserWords reads the user's list again if it changed on disk.
func (s *store) reloadUserWords() bool {
	info, err := os.Stat(s.wordsPath())
	if err != nil {
		changed := s.userWords != ""
		s.userWords = ""
		return changed
	}
	if info.ModTime().Equal(s.wordsTime) {
		return false
	}
	s.wordsTime = info.ModTime()
	data, _ := os.ReadFile(s.wordsPath())
	s.userWords = string(data)
	return true
}

// openUserWords opens the user's list in the default editor, creating it first.
func (s *store) openUserWords() {
	if _, err := os.Stat(s.wordsPath()); err != nil {
		os.MkdirAll(s.folder, 0o755)
		os.WriteFile(s.wordsPath(), []byte(strings.ReplaceAll(kazakh.UserListTemplate, "\n", "\r\n")), 0o644)
	}
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("open"))), uintptr(unsafe.Pointer(utf16Ptr(s.wordsPath()))),
		0, 0, swShowNormal)
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func without(list []string, value string) []string {
	var result []string
	for _, item := range list {
		if item != value {
			result = append(result, item)
		}
	}
	sort.Strings(result)
	return result
}

// MARK: - Starting with Windows (HKCU\...\Run)

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValue = appName

func exePath() string {
	buffer := make([]uint16, 1024)
	length, _, _ := procGetModuleFileNameW.Call(0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return syscall.UTF16ToString(buffer[:length])
}

func launchAtLogin() bool {
	var key uintptr
	if status, _, _ := procRegOpenKeyExW.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(utf16Ptr(runKey))), 0, keyRead,
		uintptr(unsafe.Pointer(&key))); status != 0 {
		return false
	}
	defer procRegCloseKey.Call(key)
	buffer := make([]uint16, 1024)
	size := uint32(len(buffer) * 2)
	status, _, _ := procRegQueryValueExW.Call(key, uintptr(unsafe.Pointer(utf16Ptr(runValue))), 0, 0,
		uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)))
	return status == 0 && strings.Contains(strings.ToLower(syscall.UTF16ToString(buffer)), strings.ToLower(exePath()))
}

func setLaunchAtLogin(on bool) {
	var key uintptr
	if status, _, _ := procRegOpenKeyExW.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(utf16Ptr(runKey))), 0,
		keySetValue, uintptr(unsafe.Pointer(&key))); status != 0 {
		return
	}
	defer procRegCloseKey.Call(key)
	if !on {
		procRegDeleteValueW.Call(key, uintptr(unsafe.Pointer(utf16Ptr(runValue))))
		return
	}
	value, _ := syscall.UTF16FromString(`"` + exePath() + `"`)
	procRegSetValueExW.Call(key, uintptr(unsafe.Pointer(utf16Ptr(runValue))), 0, regSz,
		uintptr(unsafe.Pointer(&value[0])), uintptr(len(value)*2))
}

// systemUsesDarkTheme: apps (AppsUseLightTheme) or the taskbar (SystemUsesLightTheme).
func systemUsesDarkTheme(value string) bool {
	var key uintptr
	if status, _, _ := procRegOpenKeyExW.Call(hkeyCurrentUser,
		uintptr(unsafe.Pointer(utf16Ptr(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`))), 0, keyRead,
		uintptr(unsafe.Pointer(&key))); status != 0 {
		return false
	}
	defer procRegCloseKey.Call(key)
	var light uint32
	size := uint32(4)
	status, _, _ := procRegQueryValueExW.Call(key, uintptr(unsafe.Pointer(utf16Ptr(value))), 0, 0,
		uintptr(unsafe.Pointer(&light)), uintptr(unsafe.Pointer(&size)))
	return status == 0 && light == 0
}
