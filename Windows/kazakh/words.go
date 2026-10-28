// Package kazakh restores Kazakh letters in words typed with plain Russian letters
// ("казак" -> "қазақ"). It is the Go twin of Sources/Kazakh*.swift in the Mac app: the same
// rules, the same data (Resources/kazakh) and the same tests.
package kazakh

import (
	"strings"
	"unicode"
)

// builtIn: common Kazakh words that people type with plain Russian letters. Each word is
// found by its "Russian-letter" spelling. Words whose plain spelling is an everyday Russian
// word are left out on purpose ("торт", "тур", "ага", "тенге", "казан"…).
const builtIn = `
сәлем сәлеметсіз қайырлы таң түн күн қалай қалайсың қалайсыз жақсы жақсымын қош кешіріңіз кешір
кешірім иә жоқ біз сіз сіздер менің сенің оның біздің сіздің қайда қашан қандай қанша қайсы
қайсысы қайдан мұнда мұндай қане бүгін бүгінгі ертең ертеңгі кешегі қазір кейін бұрын уақыт сағат
таңертең түс түнде үй үйде үйге жұмыс жұмыста оқы оқу оқушы оқытушы мұғалім сабақ қала қалада
қазақ қазақша қазақстан ағылшын ағылшынша тіл сөйле сөйлейсіз сөйлейсің сөйлеймін түсін түсіндім
түсінбедім түсінесің түсінік білем білмеймін білесің білесіз келдім келдің келдіңіз кеттім жүр
жүрміз бір екі үш жеті сегіз тоғыз қырық жетпіс тоқсан жүз мың бірге үшін сияқты тағы ғана және
бірақ сондықтан себебі бұл барлық көп үлкен кішкентай кішкене жаңа ескі жақын суық ыстық тамақ
әже әке қыз қызым қыздар ұлым іні сіңлі қарындас құдай сүйемін қымбат қымбаттым бақыт бақытты
денсаулық құтты құттықтаймын құттықтаймыз туған мерекеңіз болыңыз қажет мүмкін дұрыс ақша дүкен
көше қайтып қайта сенбі жексенбі дүйсенбі сейсенбі сәрсенбі бейсенбі жұма қаңтар ақпан сәуір
шілде қыркүйек қараша желтоқсан сізге саған маған оған бізге сіздерге оларға қалдым қалды
айтыңыз жазыңыз оқыңыз келіңіз отырыңыз сұраймын сұрақ қызық қызықты ұнайды ұнамайды көрем
көремін тәуелсіздік елім халық кісі әйел жігіт ғой шығар шығып кіріп беріңіз кеңес кеңсе көмек
көмектесіңіз өте оңай қиын оқасы ештеңе
`

var plainLetters = map[rune]rune{
	'ә': 'а', 'і': 'и', 'ң': 'н', 'ғ': 'г', 'қ': 'к', 'ө': 'о', 'ү': 'у', 'ұ': 'у', 'һ': 'х',
	'i': 'и', // a Latin i typed for the Kazakh і
}

// Plain is the spelling a word gets when typed with Russian letters only.
func Plain(word string) string {
	letters := []rune(strings.ToLower(word))
	for i, letter := range letters {
		if plain, ok := plainLetters[letter]; ok {
			letters[i] = plain
		}
	}
	return string(letters)
}

func isKazakhLetter(letter rune) bool {
	lower := unicode.ToLower(letter)
	_, ok := plainLetters[lower]
	return ok && lower != 'i'
}

// ContainsKazakhLetters reports whether text has any of ә і ң ғ қ ө ү ұ һ.
func ContainsKazakhLetters(text string) bool {
	return strings.IndexFunc(text, isKazakhLetter) >= 0
}

func isCyrillicLetter(letter rune) bool {
	return letter >= 0x0400 && letter <= 0x04FF && unicode.IsLetter(letter)
}

// harmony: "b" if the word's last vowel is back (а о ұ ы), "f" if front (ә е ө ү і), "-" if unknown.
func harmony(word string) string {
	letters := []rune(strings.ToLower(word))
	for i := len(letters) - 1; i >= 0; i-- {
		if strings.ContainsRune("аоұыя", letters[i]) {
			return "b"
		}
		if strings.ContainsRune("әеөүіэ", letters[i]) {
			return "f"
		}
	}
	return "-"
}

// fitsHarmony: Kazakh writes ғ and қ with back vowels, г and к with front ones; an ending
// must agree with its stem.
func fitsHarmony(stem, ending string) bool {
	stemHarmony := harmony(stem)
	if stemHarmony == "-" {
		return true
	}
	for _, letter := range ending {
		if strings.ContainsRune("аоұығқ", letter) {
			return stemHarmony == "b"
		}
		if strings.ContainsRune("әеөүігк", letter) {
			return stemHarmony == "f"
		}
	}
	return true
}

// differsOnlyByPerson: same word but "your" (ң) instead of "his" (н) near the end: "күнің" / "күнін".
func differsOnlyByPerson(first, second string) bool {
	a, b := []rune(first), []rune(second)
	if len(a) != len(b) || first == second {
		return false
	}
	for i := range a {
		if a[i] == b[i] {
			continue
		}
		pair := (a[i] == 'н' && b[i] == 'ң') || (a[i] == 'ң' && b[i] == 'н')
		if !pair || i < len(a)-4 {
			return false
		}
	}
	return true
}

// isSecondPersonEnding: a real "your"/"you" ending: -сың, -ың, -ыңыз, -ыңды, -ың + case...
// but not "-мың" (1st person).
func isSecondPersonEnding(word string) bool {
	lower := strings.ToLower(word)
	cases := []string{"", "ыз", "із", "ды", "ді", "ты", "ті", "а", "е", "нан", "нен", "мен", "да", "де", "та", "те",
		"ға", "ге", "қа", "ке", "дар", "дер"}
	for _, suffix := range cases {
		if !strings.HasSuffix(lower, "ң"+suffix) {
			continue
		}
		stem := []rune(strings.TrimSuffix(lower, "ң"+suffix))
		if len(stem) == 0 {
			return false
		}
		vowel := stem[len(stem)-1]
		if strings.ContainsRune("ыі", vowel) {
			return len(stem) < 2 || stem[len(stem)-2] != 'м'
		}
		return strings.ContainsRune("аеоөұүә", vowel)
	}
	return false
}

// thirdPersonCues: words after which "-ын/-ін" (his, hers) is meant rather than "-ың/-ің" (your).
var thirdPersonCues = set("оның", "олардың", "өзінің", "соның", "мұның", "бұның", "кімнің")

// nameEndings: endings a name or place takes in Kazakh ("Ұзынағашқа", "Қарлығашпен"), plain
// spelling. Russian case endings ("Нурлана", "Нурланом") are not here.
var nameEndings = set(
	"га", "ге", "ка", "ке", "на", "не", // to
	"ны", "ни", "ды", "ди", "ты", "ти", // object
	"нын", "нин", "дын", "дин", "тын", "тин", // of
	"да", "де", "та", "те", "дагы", "деги", "тагы", "теги",
	"дан", "ден", "тан", "тен", "нан", "нен", // from
	"мен", "бен", "пен", "сыз", "сиз", // with, without
	"дай", "дей", "тай", "тей", "дики", "тики", "ники",
	"лар", "лер", "дар", "дер", "тар", "тер",
)

// Dictionary maps plain spellings to correct words. The user's list (one word or several per
// line, # for comments) comes last, so it can add words and override built-in ones.
func Dictionary(userList string, includeBuiltIn bool) map[string]string {
	result := map[string]string{}
	if includeBuiltIn {
		for _, correct := range strings.Fields(builtIn) {
			key := Plain(correct)
			if _, taken := result[key]; key != correct && !taken {
				result[key] = correct
			}
		}
	}
	for _, line := range strings.FieldsFunc(userList, isNewline) {
		text, _, _ := strings.Cut(line, "#")
		for _, word := range strings.Fields(text) {
			correct := strings.ToLower(word)
			if key := Plain(correct); key != correct {
				result[key] = correct
			}
		}
	}
	return result
}

// UserListTemplate starts the user's own word list.
const UserListTemplate = `# Свои казахские слова для автозамены — по одному или несколько в строке, в правильном написании.
# Автозамена сама поймёт, как их набирают русскими буквами: «рақмет» сработает на «ракмет».
# Изменения подхватываются сразу после сохранения.

`

func set(words ...string) map[string]bool {
	result := make(map[string]bool, len(words))
	for _, word := range words {
		result[word] = true
	}
	return result
}

func isNewline(letter rune) bool {
	switch letter {
	case '\n', '\r', '\v', '\f', 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}
