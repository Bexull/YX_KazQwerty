<div align="center">

<img src="docs/images/icon.png" width="128" alt="Иконка YX_KazQwerty">

# YX_KazQwerty

**Казахская клавиатура для macOS и Windows**<br>
пишешь русскими буквами, получается правильный казахский

[![Скачать](https://img.shields.io/github/v/release/Bexull/YX_KazQwerty?label=%D1%81%D0%BA%D0%B0%D1%87%D0%B0%D1%82%D1%8C&color=1a8fc2)](https://github.com/Bexull/YX_KazQwerty/releases/latest)
![macOS 26+](https://img.shields.io/badge/macOS-26%2B-000000?logo=apple)
![Apple silicon](https://img.shields.io/badge/Apple%20silicon-arm64-000000)
![Windows 10/11](https://img.shields.io/badge/Windows-10%2F11%20%C2%B7%20%D0%B1%D0%B5%D1%82%D0%B0-0078D4)
[![Тесты](https://github.com/Bexull/YX_KazQwerty/actions/workflows/tests.yml/badge.svg)](https://github.com/Bexull/YX_KazQwerty/actions/workflows/tests.yml)
[![MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

| Набираешь | Получается |
|---|---|
| сиздерди шакырамыз | сіздерді шақырамыз |
| Салеметсизбе | Сәлеметсіз бе |
| мен бир жакка баргым келип тур | мен бір жаққа барғым келіп тұр |
| айтасынба | айтасың ба |
| ата ана | ата-ана |
| нягып уйттин | неге үйттің |
| Карлыгаш, Узынагашка барамын | Қарлығаш, Ұзынағашқа барамын |
| Привет, Айгерим, как дела? | Привет, Айгерим, как дела? *(русский текст не трогается)* |

Часть набора [YX_Tools](https://github.com/Bexull/YX_Tools).

</div>

## Что умеет

- **Раскладка «Русская + Қазақша»**: обычная «Русская – ПК» с цифрами плюс казахские буквы.
  Хватает двух языков: ABC и этой раскладки.
- **Казахские буквы на правом ⌘** (левый ⌘ остаётся для сочетаний клавиш) или на ⌥.
- **Автозамена**: набираешь без казахских букв, и слова исправляются сами. Русский текст,
  имена и названия в русских фразах остаются как есть.
- **Казахский словарь проверки орфографии** для программ, которые подчёркивают ошибки.

Всё работает на самом Маке. В интернет ничего не уходит.

## Установка

Нужны **macOS 26 Tahoe** и Mac на **Apple silicon** (M1 и новее).

1. Скачайте `YX_KazQwerty-1.0.zip` со страницы
   [последнего выпуска](https://github.com/Bexull/YX_KazQwerty/releases/latest) и откройте архив.
2. Перетащите YX_KazQwerty в папку «Программы».
3. Разрешите первый запуск. Приложение не нотаризовано Apple, поэтому macOS его остановит.
   Откройте YX_KazQwerty, нажмите «Готово», затем «Системные настройки» → «Конфиденциальность
   и безопасность» → внизу «Всё равно открыть». Или одной командой в Терминале:

   ```bash
   xattr -dr com.apple.quarantine /Applications/YX_KazQwerty.app
   ```

4. В строке меню появится значок клавиатуры. В его меню:
   - «Установить раскладку «Русская + Қазақша»». Она встанет вместо «Русская – ПК».
   - «Дать доступ к клавиатуре…». Правому ⌘ и автозамене нужен «Универсальный доступ»:
     Системные настройки → Конфиденциальность и безопасность → Универсальный доступ → YX_KazQwerty.
5. Перезапустите программы из App Store, которые были открыты (Teams, Telegram, WhatsApp,
   Outlook), иначе в них новая раскладка не выбирается. Проще всего выйти из системы и войти
   снова: заодно включится казахский словарь проверки орфографии.

Чтобы YX_KazQwerty запускался сам: в его меню «Открывать при входе в систему».

### Windows

Бета для Windows 10 и 11: казахские буквы на правом Alt в обычной русской раскладке и та же
автозамена. Один файл `YX_KazQwerty.exe`, устанавливать ничего не нужно. Как скачать, запустить
и проверить: [Windows/README.md](Windows/README.md).

> [!TIP]
> Нужна ещё громкость каждого приложения? [YX_Tools](https://github.com/Bexull/YX_Tools) — это
> YX_KazQwerty и [YX_Dybys](https://github.com/Bexull/YX_Dybys) в одном значке. Вместе с YX_Tools
> YX_KazQwerty не запускайте: каждое слово исправлялось бы дважды.

## Казахские буквы

Правый ⌘ (или ⌥) и клавиша русской буквы; с ⇧ получаются заглавные.

| Клавиша | а | и | н | г | к | о | у | ц | х |
|---|---|---|---|---|---|---|---|---|---|
| **с правым ⌘** | ә | і | ң | ғ | қ | ө | ү | ұ | һ |

## Автозамена

- Слово исправляется, когда нажимаете пробел или знак препинания.
- Сначала определяется язык фразы: русские слова и имена в русском тексте остаются как есть.
- Разговорное пишется правильно: «нестимиз» → «не істейміз», «нягып» → «неге»,
  «білмимін» → «білмеймін». Разговорные окончания остаются: «жатсың», «боп», «қып».
- Аббревиатуры («ӘКК», «АҚШ») исправляются, только если набраны заглавными: «акк» в
  переписке остаётся аккаунтом.
- **Отменить:** ⌫ два раза сразу после исправления. Первый ⌫ стирает только пробел, так что
  можно дописать «!» или «?».
- **Обучение:** если дважды набрать казахскими буквами слово, которого нет в словаре,
  YX_KazQwerty его запомнит. Сбросить: «Забыть выученное».
- **Свои слова:** «Мои слова для автозамены…» открывает файл
  `~/Library/Application Support/YX_KazQwerty/kazakh-words.txt`, куда можно вписать свои написания.
- В терминалах, редакторах кода и менеджерах паролей автозамена выключена. В любой другой
  программе её можно выключить пунктом «Не исправлять в «…»».

Слово исправилось неправильно? [Расскажите](https://github.com/Bexull/YX_KazQwerty/issues/new?template=word.yml),
так автозамена становится точнее.

## Частые вопросы

<details>
<summary>macOS пишет, что приложение повреждено или его нельзя открыть</summary>

Так Gatekeeper встречает приложения, не нотаризованные Apple. Выполните в Терминале
`xattr -dr com.apple.quarantine /Applications/YX_KazQwerty.app` или разрешите запуск в
«Конфиденциальность и безопасность».
</details>

<details>
<summary>В Teams или Telegram печатается только по-английски</summary>

Программы из App Store хранят свою копию списка раскладок. Если она сделана до установки
«Русская + Қазақша», новой раскладки в ней нет. YX_KazQwerty удаляет такие копии сам; закройте
программу полностью (⌘Q) и откройте снова.
</details>

<details>
<summary>Автозамена не работает</summary>

Проверьте, что выбрана раскладка «Русская + Қазақша» (автозамена работает только в ней),
что у YX_KazQwerty есть «Универсальный доступ» и что программа не в списке «Не исправлять в…».
</details>

<details>
<summary>Как удалить</summary>

Выйдите из YX_KazQwerty (в его меню «Выйти») и удалите его из «Программ». Чтобы убрать
раскладку, словарь и настройки:

```bash
rm -rf ~/Library/Keyboard\ Layouts/RussianKazakh.bundle ~/Library/Spelling/kk.dic ~/Library/Spelling/kk.aff
rm -rf ~/Library/Application\ Support/YX_KazQwerty
defaults delete com.bexul.kazqwerty
```

Затем верните «Русская – ПК»: Системные настройки → Клавиатура → Источники ввода.
</details>

## Сборка из исходников

Нужны macOS 26 и Xcode 26.

```bash
git clone https://github.com/Bexull/YX_KazQwerty.git
cd YX_KazQwerty
./build.sh --install
```

- `./build.sh` только собирает `build/YX_KazQwerty.app`, `--install` ставит в «Программы»,
  `--release` ещё и пакует `build/YX_KazQwerty-<версия>.zip`.
- Без сертификата сборка подписывается ad hoc, и после каждой пересборки macOS заново
  спрашивает «Универсальный доступ». Чтобы не спрашивала, положите хэш своего сертификата
  Apple Development (из `security find-identity -v -p codesigning`) в файл `.sign-identity`
  или в переменную `MIXER_SIGN_IDENTITY`.

## Как устроено

Версия для Windows лежит в `Windows/` и описана в [её README](Windows/README.md).

- `Sources/Keyboard/KazakhKeyboard.swift`: перехват клавиатуры (правый ⌘ и автозамена),
  обучение, исключённые программы.
- `Sources/Keyboard/KazakhTyping.swift`: сама автозамена. Решает, как должно выглядеть каждое
  слово, по словарю, соседним словам и языку фразы. Не зависит от AppKit, поэтому тестируется
  отдельно.
- `Sources/Keyboard/KazakhLexicon.swift`, `Sources/Keyboard/KazakhWords.swift`: словарь в памяти
  и встроенный список частых слов.
- `Sources/Keyboard/KazakhLayout.swift`, `Resources/RussianKazakh.bundle`: раскладка и её установка.
- `Sources/Keyboard/KazakhKeyboardMenu.swift`, `Sources/App/KazQwertyApp.swift`: значок в строке
  меню и его меню.
- `Resources/kazakh/`: словарь автозамены, собранный из корпусов скриптом
  `Tools/make-kazakh-dictionary.py`.
- `Resources/AppIcon.icon`: иконка Liquid Glass (открывается в Icon Composer).

Как исследовались и проверялись правила автозамены:
[docs/kazakh-autocorrect-research.md](docs/kazakh-autocorrect-research.md),
[docs/kazakh-colloquial-analysis.md](docs/kazakh-colloquial-analysis.md).

## Разработка

**Тесты автозамены** (они же запускаются на GitHub при каждом изменении):

```bash
mkdir -p build
swiftc -O -o build/test-typing Tools/TypingTests/main.swift Sources/Keyboard/KazakhTyping.swift \
  Sources/Keyboard/KazakhWords.swift Sources/Keyboard/KazakhLexicon.swift
build/test-typing Resources/kazakh Tools/TypingTests/chat-phrases.tsv Tools/TypingTests/russian-phrases.tsv
```

`holdout-*.tsv` — фразы, на которых правила не настраивались, `colloquial.tsv` — разговорные
формы. Любое изменение русского текста считается ошибкой. Проверка на 3000 настоящих отзывах:
`Tools/make-review-test.py` делает набор из тестовой части KazSAnDRA, `Tools/ReviewTest`
считает, сколько слов восстановлено.

**Пересборка словаря.** Скачайте корпуса из раздела «Данные» и выполните:

```bash
python3 Tools/make-kazakh-dictionary.py Resources/kazakh \
    --words kaz_newscrawl_2016_1M-words.txt --sentences kaz_newscrawl_2016_1M-sentences.txt \
    --russian ru_full.txt Tools/russian-kazakhstan.txt --subtitles kk_full.txt \
    --informal 01_pc_train_ib.csv 04_pc_valid.csv \
    --chat Tools/kazakh-chat-words.txt --phrases Tools/kazakh-phrases.txt --joins Tools/kazakh-joins.txt \
    --spoken Tools/kazakh-spoken.txt
```

Ручные списки в `Tools/`: разговорные слова, устойчивые фразы, слитные слова, русские слова
Казахстана, разговорные формы. Раскладка пересобирается так:
`swiftc -o build/make-keyboard-layout Tools/make-keyboard-layout.swift && build/make-keyboard-layout Resources/RussianKazakh.bundle`.

**Отладка.** `pkill -USR1 -f YX_KazQwerty.app/Contents/MacOS/YX_KazQwerty` записывает состояние
клавиатуры в `~/Library/Logs/YX_KazQwerty/state.txt`.

## Данные

Словарь автозамены собран из открытых корпусов:

- [Leipzig Corpora Collection](https://wortschatz.uni-leipzig.de/en/download), `kaz_newscrawl_2016_1M`:
  казахские новости, CC BY (Goldhahn, Eckart, Quasthoff, LREC 2012).
- [FrequencyWords](https://github.com/hermitdave/FrequencyWords) (Hermit Dave): частотные
  списки по субтитрам OpenSubtitles, CC BY-SA 4.0.
- [KazSAnDRA](https://github.com/IS2AI/KazSAnDRA) (ISSAI): неформальные казахские отзывы, CC BY 4.0.

Спасибо авторам этих корпусов.

## Лицензия

Код распространяется по лицензии [MIT](LICENSE), данные автозамены в `Resources/kazakh` — по
[CC BY-SA 4.0](Resources/kazakh/LICENSE.md).

---

### In English

**YX_KazQwerty** is a Kazakh keyboard for macOS 26 on Apple silicon, with a Windows beta in
[Windows/](Windows/README.md). It installs a combined
Russian + Kazakh layout, puts the Kazakh letters on the right ⌘ key, and runs an on-device
autocorrect that turns Kazakh typed with plain Russian letters into proper Kazakh spelling
("салеметсизбе" → "Сәлеметсіз бе"), while leaving Russian text alone. Download the latest build
from [Releases](https://github.com/Bexull/YX_KazQwerty/releases/latest). The app is not notarized,
so run `xattr -dr com.apple.quarantine /Applications/YX_KazQwerty.app` after installing. It is
part of [YX_Tools](https://github.com/Bexull/YX_Tools). Code: MIT. Autocorrect data: CC BY-SA 4.0.
