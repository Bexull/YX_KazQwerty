package kazakh

import (
	"strings"
	"unicode/utf8"
)

// particle is a Kazakh particle people write together with the previous word.
type particle struct {
	plain      string
	isQuestion bool
	form       func(word string) string
	// fits: whether the word before it can take this particle at all.
	fits func(word string) bool
}

// attachable: particles people write together with the previous word.
var attachable = func() []particle {
	var list []particle
	for _, plain := range []string{"ма", "ме", "ба", "бе", "па", "пе"} {
		list = append(list, particle{plain: plain, isQuestion: true, form: question, fits: isPredicate})
	}
	isPronoun := func(word string) bool { return pronouns[strings.ToLower(word)] }
	isLimited := func(word string) bool { return limited[strings.ToLower(word)] }
	return append(list,
		particle{plain: "ше", form: func(string) string { return "ше" }, fits: isPronoun},
		particle{plain: "гой", form: emphatic, fits: isPredicate},
		particle{plain: "кой", form: emphatic, fits: isPredicate},
		// Spoken "ғой": "келе жатамго" is "келе жатам ғой".
		particle{plain: "го", form: emphatic, fits: isPredicate},
		particle{plain: "ко", form: emphatic, fits: isPredicate},
		particle{plain: "гана", form: only, fits: isLimited},
		particle{plain: "кана", form: only, fits: isLimited},
	)
}()

var (
	pronouns       = set("мен", "сен", "ол", "біз", "сіз", "сендер", "сіздер", "олар")
	limited        = set("бір", "аз", "сен", "мен", "сіз", "біз", "ол", "осы", "сол", "тек", "сәл")
	predicateWords = set(
		"жақсы", "дұрыс", "рас", "бар", "жоқ", "мүмкін", "керек", "дайын", "аман", "сау", "шын", "шынымен",
		"солай", "осылай", "бос", "ыстық", "суық", "қымбат", "арзан", "мұнда", "сонда", "осында", "бүгін",
		"ертең", "қазір", "сен", "сіз", "сендер", "сіздер", "ол", "олар", "мен", "біз", "сенікі", "менікі",
		"жатыр", "отыр", "тұр", "жүр",
	)
	predicateEndings = []string{
		"сың", "сің", "сыз", "сіз", "мын", "мін", "бын", "бін", "пын", "пін", "мыз", "міз", "быз", "біз",
		"пыз", "піз", "дар", "дер", "тар", "тер", "дың", "дің", "тың", "тің", "ңыз", "ңіз", "ды", "ді",
		"ты", "ті", "ған", "ген", "қан", "кен", "атын", "етін", "йтын", "йтін", "ады", "еді", "йды", "йді",
	}
	// shortEndings: "I did", "we did" ("айттым", "келдік"), and spoken "I do" ("жатам",
	// "келем", "ойлайм" for "жатырмын", "келемін", "ойлаймын"). Only after a few letters:
	// "тік" alone is a verb.
	shortEndings = []string{"дым", "дім", "тым", "тім", "дық", "дік", "тық", "тік", "ам", "ем", "йм"}
)

// isPredicate: a finished statement that a question particle can follow ("айтасың", "болды",
// "жақсы"), as opposed to a bare verb stem ("кешік" + "пе" is "don't be late").
func isPredicate(word string) bool {
	lower := strings.ToLower(word)
	if predicateWords[lower] {
		return true
	}
	for _, ending := range predicateEndings {
		if strings.HasSuffix(lower, ending) {
			return true
		}
	}
	for _, ending := range shortEndings {
		if utf8.RuneCountInString(lower) >= utf8.RuneCountInString(ending)+2 && strings.HasSuffix(lower, ending) {
			return true
		}
	}
	return false
}

// separateParticle: the form a particle typed as a word of its own should take ("ба" after
// "шынымен" is "бе").
func separateParticle(plain, after string) (string, bool) {
	switch plain {
	case "ма", "ме", "ба", "бе", "па", "пе":
		return question(after), true
	case "гой":
		return emphatic(after), true
	case "го", "ко":
		// Spoken "ғой" after a statement ("жатам го"), not slang "го" (let's go) or Russian "ко".
		if isPredicate(after) {
			return emphatic(after), true
		}
		return "", false
	case "гана":
		return only(after), true
	}
	return "", false
}

// question: ма/ме after vowels and л, р, у, й; ба/бе after м, н, ң, з, ж; па/пе after the
// rest. The vowel follows the word's last vowel.
func question(after string) string {
	lower := strings.ToLower(after)
	last, _ := utf8.DecodeLastRuneInString(lower)
	if lower == "" {
		return "ма"
	}
	consonant := "п"
	switch {
	case strings.ContainsRune("аәеоөұүыіиэюяуйлр", last):
		consonant = "м"
	case strings.ContainsRune("мнңзж", last):
		consonant = "б"
	}
	if harmony(lower) == "f" {
		return consonant + "е"
	}
	return consonant + "а"
}

// emphatic: ғой after voiced sounds, қой after voiceless ones.
func emphatic(after string) string {
	if isVoiceless(after) {
		return "қой"
	}
	return "ғой"
}

func only(after string) string {
	if isVoiceless(after) {
		return "қана"
	}
	return "ғана"
}

func isVoiceless(word string) bool {
	last, _ := utf8.DecodeLastRuneInString(strings.ToLower(word))
	return word != "" && strings.ContainsRune("кқпстфхцчшщ", last)
}
