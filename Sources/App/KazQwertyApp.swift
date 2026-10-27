import AppKit
import ServiceManagement

@main
enum KazQwertyApp {
    static func main() {
        MainActor.assumeIsolated {
            let app = NSApplication.shared
            let delegate = AppDelegate()
            app.delegate = delegate
            app.setActivationPolicy(.accessory)
            withExtendedLifetime(delegate) {
                app.run()
            }
        }
    }
}

/// A keyboard icon in the menu bar; its menu holds every setting.
@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate, NSMenuDelegate {
    private let keyboard = KazakhKeyboard()
    private lazy var keyboardMenu = KazakhKeyboardMenu(keyboard: keyboard)
    private var statusItem: NSStatusItem?
    private var dumpSignal: DispatchSourceSignal?

    func applicationDidFinishLaunching(_ notification: Notification) {
        let arguments = ProcessInfo.processInfo.arguments
        if arguments.contains("--install-keyboard-layout") || arguments.contains("--check-keyboard-layout") {
            if arguments.contains("--install-keyboard-layout") {
                do {
                    try KazakhLayout.install().forEach { print($0) }
                } catch {
                    print("install failed: \(error)")
                }
            }
            KazakhLayout.check().forEach { print($0) }
            print("current input source: \(KazakhLayout.currentInputSourceID ?? "-")")
            fflush(stdout)
            NSApp.terminate(nil)
            return
        }

        let item = NSStatusBar.system.statusItem(withLength: NSStatusItem.squareLength)
        let image = NSImage(systemSymbolName: "keyboard", accessibilityDescription: "YX_KazQwerty")
        image?.isTemplate = true
        item.button?.image = image
        item.button?.toolTip = "YX_KazQwerty"
        let menu = NSMenu()
        menu.delegate = self
        item.menu = menu
        statusItem = item
        keyboard.start()

        // `kill -USR1 <pid>` writes the keyboard state to ~/Library/Logs/YX_KazQwerty/state.txt.
        signal(SIGUSR1, SIG_IGN)
        let dump = DispatchSource.makeSignalSource(signal: SIGUSR1, queue: .main)
        dump.setEventHandler { [weak self] in
            MainActor.assumeIsolated { self?.writeStateDump() }
        }
        dump.resume()
        dumpSignal = dump
    }

    func menuNeedsUpdate(_ menu: NSMenu) {
        menu.removeAllItems()
        keyboardMenu.items().forEach(menu.addItem)
        menu.addItem(.separator())
        let login = NSMenuItem(title: "Открывать при входе в систему", action: #selector(toggleLaunchAtLogin), keyEquivalent: "")
        login.target = self
        login.state = SMAppService.mainApp.status == .enabled ? .on : .off
        menu.addItem(login)
        menu.addItem(.separator())
        let quit = NSMenuItem(title: "Выйти", action: #selector(quit), keyEquivalent: "q")
        quit.target = self
        menu.addItem(quit)
    }

    @objc private func toggleLaunchAtLogin() {
        do {
            if SMAppService.mainApp.status == .enabled {
                try SMAppService.mainApp.unregister()
            } else {
                try SMAppService.mainApp.register()
            }
        } catch {
            NSSound.beep()
        }
    }

    @objc private func quit() {
        NSApp.terminate(nil)
    }

    private func writeStateDump() {
        let directory = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Logs/YX_KazQwerty")
        try? FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        try? (keyboard.diagnostics() + "\n").write(to: directory.appendingPathComponent("state.txt"), atomically: true, encoding: .utf8)
    }
}
