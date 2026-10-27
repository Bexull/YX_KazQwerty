import AppKit

/// The keyboard's menu items, for a status item menu (YX_KazQwerty) or a submenu (YX_Tools).
/// Built again each time the menu opens, so the checkmarks and the app name are current.
@MainActor
final class KazakhKeyboardMenu: NSObject {
    private let keyboard: KazakhKeyboard

    init(keyboard: KazakhKeyboard) {
        self.keyboard = keyboard
    }

    func items() -> [NSMenuItem] {
        var items: [NSMenuItem] = []
        func add(_ title: String, _ action: Selector, on: Bool? = nil) {
            let item = NSMenuItem(title: title, action: action, keyEquivalent: "")
            item.target = self
            if let on { item.state = on ? .on : .off }
            items.append(item)
        }
        add(keyboard.layoutInstalled ? "Переустановить раскладку" : "Установить раскладку «Русская + Қазақша»",
            #selector(installLayout))
        add("Казахские буквы на правом ⌘", #selector(toggleRightCommand), on: keyboard.rightCommand)
        add("Автозамена казахских слов", #selector(toggleAutocorrect), on: keyboard.autocorrect)
        add("Мои слова для автозамены…", #selector(openUserWords))
        if let frontmost = keyboard.frontmostApp {
            add("Не исправлять в «\(frontmost.name)»", #selector(toggleExclusion), on: keyboard.isFrontmostExcluded)
        }
        if keyboard.learnedCount > 0 {
            add("Забыть выученное (\(keyboard.learnedCount))", #selector(forgetLearned))
        }
        if keyboard.needsAccess {
            items.append(.separator())
            add("Дать доступ к клавиатуре…", #selector(requestAccess))
        }
        return items
    }

    @objc private func installLayout() { keyboard.installLayout() }
    @objc private func toggleRightCommand() { keyboard.rightCommand.toggle() }
    @objc private func toggleAutocorrect() { keyboard.autocorrect.toggle() }
    @objc private func openUserWords() { keyboard.openUserWords() }
    @objc private func toggleExclusion() { keyboard.toggleFrontmostExclusion() }
    @objc private func forgetLearned() { keyboard.forgetLearned() }
    @objc private func requestAccess() { keyboard.requestAccess() }
}
