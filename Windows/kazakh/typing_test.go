package kazakh

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// The same checks as Tools/TypingTests/main.swift in the Mac app.

type fakeCorpus map[string]*Entry

func (c fakeCorpus) Entry(plain string) *Entry { return c[plain] }

func entry(forms ...any) *Entry {
	e := &Entry{}
	for i := 0; i < len(forms); i++ {
		switch value := forms[i].(type) {
		case string:
			e.Forms = append(e.Forms, Form{Text: value, Count: forms[i+1].(int)})
			i++
		case flag:
			switch value {
			case russian:
				e.IsRussianToo = true
			case name:
				e.IsName = true
			case mostlyKazakh:
				e.IsMostlyKazakh = true
			}
		}
	}
	return e
}

type flag int

const (
	russian flag = iota
	name
	mostlyKazakh
)

func fakeVocabulary() *Vocabulary {
	return &Vocabulary{
		Words: Dictionary("рақмет\n# comment\nсәлем # inline", true),
		Corpus: fakeCorpus{
			"агайын": entry("ағайын", 500), "туыс": entry("туыс", 900),
			"сиздерди": entry("сіздерді", 300), "шакырамыз": entry("шақырамыз", 200),
			"жакка": entry("жаққа", 700), "баргым": entry("барғым", 90), "келип": entry("келіп", 9000),
			"тур": entry("тұр", 11000, "тур", 140, "түр", 85, russian),
			"мен": entry("мен", 90000, russian), "бир": entry("бір", 50000),
			"он": entry("он", 5000, "оң", 2300, russian), "сени": entry("сені", 730, russian),
			"жаксы": entry("жақсы", 8000), "коремин": entry("көремін", 400),
			"айгерим": entry("әйгерім", 159, "айгерім", 26, name), "кайда": entry("қайда", 3000),
			"сал":  entry("сәл", 1200, "сал", 234, russian),
			"озин": entry("өзін", 4400, "өзің", 590), "кунин": entry("күнін", 790, "күнің", 50, russian),
			"туган": entry("туған", 6000), "жолын": entry("жолын", 1400, "жолың", 50),
			"болсын": entry("болсын", 3000), "айтасын": entry("айтасың", 120, "айтасын", 40),
			"ата": entry("ата", 4000), "ана": entry("ана", 3000, russian),
			"шынымен": entry("шынымен", 400), "ба": entry("ба", 900, russian),
			"корсет": entry("көрсет", 300, russian), "бар": entry("бар", 20000, russian),
			"барма": entry("барма", 30), "кешке": entry("кешке", 500),
			"нурлан": entry("нұрлан", 1610, name), "ерлан": entry("ерлан", 1164, name),
			"узынагаш": entry("ұзынағаш", 45, name),
			"аска":     entry("асқа", 205, russian, mostlyKazakh), "жатам": entry("жатам", 10),
			"ойын": entry("ойын", 9000, "ойың", 120), "ойындар": entry("ойындар", 900),
			"салеметсиз": entry("сәлеметсіз", 300),
		},
		LeftContext:  map[string]string{"туған\tкунин": "күнің"},
		RightContext: map[string]string{"жолын\tболсын": "жолың"},
		Pairs:        map[string]string{"ата\tана": "ата-ана"},
		Endings: map[string][]Form{
			"f\tши": {{"ші", 215}},
			"b\tка": {{"қа", 720}, {"ка", 79}},
		},
	}
}

// typeKeys types keys ("⌫" backspace, "⟲" reset) into a pretend text field and returns its
// text. Every edit must find on screen exactly the text it expects to delete.
func typeKeys(t *testing.T, keys string, vocabulary *Vocabulary) string {
	var typing Typing
	text := ""
	for _, key := range strings.Split(keys, "") {
		var action Action
		switch key {
		case "⌫":
			action = typing.Backspace()
		case "⟲":
			typing.Reset()
		default:
			action = typing.Text(key, vocabulary)
		}
		switch action.Kind {
		case Pass:
			if key == "⌫" {
				text = dropLastRunes(text, 1)
			} else if key != "⟲" {
				text += key
			}
		case Replace, Undo:
			if !strings.HasSuffix(text, action.Expect) || utf8.RuneCountInString(action.Expect) != action.Delete {
				t.Errorf("edit expected «%s» but screen ends «%s»", action.Expect, lastRunes(text, action.Delete))
			}
			text = dropLastRunes(text, action.Delete) + action.Insert
			if action.Kind == Replace {
				text += key
			}
		}
	}
	return text
}

func TestTyping(t *testing.T) {
	fake := fakeVocabulary()
	checks := []struct{ name, keys, expected string }{
		{"hand-picked word", "казак ", "қазақ "},
		{"capitalized", "Казак, ", "Қазақ, "},
		{"all caps", "КАЗАКСТАН!", "ҚАЗАҚСТАН!"},
		{"already right", "қазақ ", "қазақ "},
		{"Russian untouched", "привет как дела ", "привет как дела "},
		{"backspace takes back the space only", "казак ⌫", "қазақ"},
		{"then punctuation", "казак ⌫!", "қазақ!"},
		{"second backspace undoes", "казак ⌫⌫", "казак"},
		{"undo then keep", "казак ⌫⌫ ", "казак "},
		{"split particle, then !", "салеметсизбе ⌫!", "сәлеметсіз бе!"},
		{"split particle in capitals", "САЛЕМЕТСИЗБЕ ", "СӘЛЕМЕТСІЗ БЕ "},
		{"reset mid-word", "каз⟲ак ", "казак "},
		{"user word", "ракмет ", "рақмет "},
		{"inflected words", "агайын туыс, сиздерди шакырамыз ", "ағайын туыс, сіздерді шақырамыз "},
		{"Kazakh phrase fixes тур", "мен бир жакка баргым келип тур ", "мен бір жаққа барғым келіп тұр "},
		{"Russian phrase keeps тур", "поедем в тур ", "поедем в тур "},
		{"тур alone stays", "тур ", "тур "},
		{"first word fixed by the next", "сени жаксы коремин ", "сені жақсы көремін "},
		{"name in Russian", "Привет, Айгерим ", "Привет, Айгерим "},
		{"name in Kazakh", "Айгерим кайда ", "Әйгерім қайда "},
		{"name alone", "Айгерим ", "Әйгерім "},
		{"name taken back by Russian", "Айгерим, привет ", "Айгерим, привет "},
		{"names held for a take-back", "Нурлан, Ерлан, привет ", "Нурлан, Ерлан, привет "},
		{"back into held names", "Нурлан, Ерлан ⌫⌫⌫ан ", "Нұрлан, Ерлан "},
		{"name + case ending", "Узынагашка ", "Ұзынағашқа "},
		{"no Russian case ending", "Нурлана ", "Нурлана "},
		{"small name waits for Kazakh", "узынагашка жаксы ", "ұзынағашқа жақсы "},
		{"small name alone stays", "узынагашка ", "узынагашка "},
		{"rare Russian look-alike alone", "аска ", "асқа "},
		{"rare look-alike taken back", "аска, привет ", "аска, привет "},
		{"common Russian look-alike waits", "тур ", "тур "},
		{"spoken ғой, merged", "жатамго ", "жатам ғой "},
		{"spoken ғой, separate", "жатам го ", "жатам ғой "},
		{"slang го untouched", "го ", "го "},
		{"common word kept", "мен сал ", "мен сал "},
		{"your, not his", "озин ", "өзің "},
		{"a word of its own, not his", "ойын ", "ойын "},
		{"his after оның", "оның озин ", "оның өзін "},
		{"left context", "туган кунин ", "туған күнің "},
		{"right context", "жолын болсын ", "жолың болсын "},
		{"merged question", "айтасынба ", "айтасың ба "},
		{"question mark splits", "барма?", "бар ма?"},
		{"no split without ?", "барма ", "барма "},
		{"particle form", "шынымен ба ", "шынымен бе "},
		{"hyphen pair", "ата ана ", "ата-ана "},
		{"stem + ending", "корсетши ", "көрсетші "},
		{"undo a two-word fix", "сени жаксы ⌫⌫", "сени жаксы"},
		{"return resets", "мен\rтур ", "мен\rтур "},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if got := typeKeys(t, check.keys, fake); got != check.expected {
				t.Errorf("got «%s», expected «%s»", got, check.expected)
			}
		})
	}
}

func TestLearnsUnknownKazakhWords(t *testing.T) {
	var typing Typing
	fake := fakeVocabulary()
	for _, key := range strings.Split("үйттің ", "") {
		typing.Text(key, fake)
	}
	if learned := typing.TakeLearned(); len(learned) != 1 || learned[0] != "үйттің" {
		t.Errorf("learned %v", learned)
	}
}

// The phrase sets and the real data of the Mac app: Tools/TypingTests/*.tsv, Resources/kazakh.
func TestPhraseSets(t *testing.T) {
	root := filepath.Join("..", "..")
	vocabulary, err := Load(os.DirFS(filepath.Join(root, "Resources", "kazakh")))
	if err != nil {
		t.Skip("no Resources/kazakh:", err)
	}
	vocabulary.Words = Dictionary("", true)
	sets := []struct {
		file   string
		strict bool // Russian text must never change
	}{
		{"chat-phrases.tsv", false}, {"russian-phrases.tsv", true},
		{"holdout-kazakh.tsv", false}, {"holdout-russian.tsv", true}, {"colloquial.tsv", false},
	}
	for _, set := range sets {
		data, err := os.ReadFile(filepath.Join(root, "Tools", "TypingTests", set.file))
		if err != nil {
			t.Errorf("%s: %v", set.file, err)
			continue
		}
		right, total := 0, 0
		var misses []string
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			columns := strings.Split(line, "\t")
			typed, expected := columns[0], columns[0]
			if len(columns) > 1 {
				expected = columns[1]
			}
			total++
			got := strings.TrimSpace(typeKeys(t, typed+" ", vocabulary))
			if got == expected {
				right++
			} else {
				misses = append(misses, "«"+typed+"» → «"+got+"» (надо «"+expected+"»)")
			}
		}
		sort.Strings(misses)
		t.Logf("%s: %d of %d right %s", set.file, right, total, strings.Join(misses, "; "))
		if set.strict && right != total {
			t.Errorf("%s: Russian text changed: %s", set.file, strings.Join(misses, "; "))
		}
	}
}
