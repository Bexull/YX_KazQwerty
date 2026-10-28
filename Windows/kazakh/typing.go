package kazakh

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ActionKind says what to do with the key that was just typed.
type ActionKind int

const (
	// Pass lets the key through untouched.
	Pass ActionKind = iota
	// Replace drops the key, deletes Delete characters (which should read Expect), types
	// Insert, then lets the key through.
	Replace
	// Undo drops the backspace, deletes Delete characters (Expect) and types Insert: undoes
	// the last replacement. It takes a second ⌫: the first one only erases the key that ended
	// the word, so a "!" can follow "Сәлеметсіз бе".
	Undo
)

// Action is the answer to one key.
type Action struct {
	Kind   ActionKind
	Delete int
	Insert string
	Expect string
}

// Language is the language of the current phrase, judged from its words so far.
type Language int

const (
	Unknown Language = iota
	Kazakh
	Russian
)

type word struct {
	typed string
	shown string
	// Left as typed because the phrase language wasn't known yet ("тур", names).
	waitingForContext bool
	// Corrected before the language was known, though it is a Russian word too ("Сени"):
	// a Russian word next takes the correction back.
	provisional bool
	// The user undid our change, or it is a finished pair: leave it alone.
	locked bool
	// Its vote in score, so it can be recounted when the word changes later.
	evidence int
}

type edit struct {
	previousBefore  *word
	separatorBefore string
	typed           string
	inserted        string
	boundary        string
	original        string
	// How much of boundary is still on screen.
	boundaryLeft int
}

// Typing is the autocorrect bookkeeping, separate from the keyboard plumbing so it can be
// tested on its own.
//
// It follows what is typed since the cursor last moved, as two words: the previous finished
// word and the one being typed. When a word ends it decides how that word should look, and
// may also fix the previous one now that it knows what follows ("сени жаксы" → "сені жақсы",
// "жолын болсын" → "жолың болсын", "ата ана" → "ата-ана").
type Typing struct {
	// Votes from the phrase so far: Kazakh letters count for Kazakh, plainly Russian words
	// for Russian. One both-language word ("он") can't flip a clearly Kazakh phrase.
	score         int
	learned       []string
	previous      *word
	separator     string
	current       string
	currentLocked bool
	lastEdit      *edit
}

// Language of the phrase being typed.
func (t *Typing) Language() Language {
	switch {
	case t.score > 0:
		return Kazakh
	case t.score < 0:
		return Russian
	}
	return Unknown
}

// TakeLearned returns the unknown Kazakh words typed since the last call (lowercase): the
// app learns a word typed twice.
func (t *Typing) TakeLearned() []string {
	words := t.learned
	t.learned = nil
	return words
}

// Reset: arrows, clicks, shortcuts, another window: the cursor may have moved.
func (t *Typing) Reset() {
	t.resetWindow()
	t.score = 0
}

// Backspace handles ⌫.
func (t *Typing) Backspace() Action {
	if t.lastEdit != nil {
		if t.lastEdit.boundaryLeft > 0 {
			// Only the space (or comma) after the corrected word goes.
			t.lastEdit.boundaryLeft--
			t.eraseOne()
			return Action{}
		}
		return t.undo(t.lastEdit)
	}
	t.eraseOne()
	return Action{}
}

// Text handles what a key typed.
func (t *Typing) Text(text string, vocabulary *Vocabulary) Action {
	t.lastEdit = nil
	if text != "" && allLettersOrNumbers(text) {
		t.current += text
		return Action{}
	}
	if t.current == "" {
		if strings.IndexFunc(text, isNewline) >= 0 {
			t.resetWindow()
			t.score = 0
		} else if t.previous != nil {
			t.separator += text
		}
		return Action{}
	}
	return t.finishWord(text, vocabulary)
}

// MARK: - Finishing a word

func (t *Typing) finishWord(boundary string, vocabulary *Vocabulary) Action {
	typed := t.current
	finished := word{typed: typed, shown: typed, locked: t.currentLocked}
	evidence := 0
	known := false
	if t.currentLocked {
		if ContainsKazakhLetters(typed) {
			evidence = 2
		}
	} else {
		before, hasBefore := "", false
		if t.previous != nil {
			before, hasBefore = lastToken(t.previous.shown), true
		}
		decision := t.decide(typed, before, hasBefore, boundary, t.Language(), vocabulary)
		finished.shown = decision.shown
		finished.waitingForContext = decision.waitingForContext
		finished.provisional = decision.provisional
		evidence = decision.evidence
		finished.evidence = evidence
		known = decision.known
	}

	// The finished word may change how the previous one should look.
	newPrevious := copyWord(t.previous)
	joined, hasJoined := "", false
	linked := strings.IndexFunc(t.separator, isNewline) < 0 && allSpacesOrPunctuation(t.separator)
	if t.previous != nil && !t.previous.locked && linked {
		before := *t.previous
		beforePlain := Plain(before.typed)
		if form, ok := vocabulary.RightContext[beforePlain+"\t"+strings.ToLower(finished.shown)]; ok {
			before.shown = adapt(form, before.typed)
			before.waitingForContext = false
			before.provisional = false
		} else if before.waitingForContext && evidence > 0 {
			before.shown = t.decide(before.typed, "", false, t.separator, Kazakh, vocabulary).shown
			before.waitingForContext = false
			t.score += 2 - before.evidence
			before.evidence = 2
		} else if before.provisional && evidence < 0 {
			before.shown = before.typed
			before.provisional = false
			t.score += -1 - before.evidence
			before.evidence = -1
		}
		if t.separator == " " && t.Language() != Russian && !strings.Contains(before.shown, " ") &&
			!strings.Contains(finished.shown, " ") {
			if pair, ok := vocabulary.joined(before.typed, typed, before.shown, finished.shown); ok {
				joined, hasJoined = adapt(pair, before.typed), true
			}
		}
		newPrevious = &before
	}

	t.score = max(-3, min(3, t.score+evidence))
	if !known && !t.currentLocked && ContainsKazakhLetters(typed) {
		t.learned = append(t.learned, strings.ToLower(typed))
	}

	previousChanged := hasJoined || (t.previous != nil && newPrevious.shown != t.previous.shown)
	oldSpan, newSpan := typed, finished.shown
	if previousChanged {
		oldSpan = t.previous.shown + t.separator + typed
		newSpan = newPrevious.shown + t.separator + finished.shown
	}
	if hasJoined {
		newSpan = joined
	}

	pending := &edit{previousBefore: copyWord(t.previous), separatorBefore: t.separator, typed: typed,
		inserted: newSpan, boundary: boundary, original: oldSpan, boundaryLeft: utf8.RuneCountInString(boundary)}

	// Slide the window: the finished word becomes the previous one. A joined pair is one word now.
	switch {
	case hasJoined && t.previous != nil:
		t.previous = &word{typed: t.previous.typed + t.separator + typed, shown: joined, locked: true}
	case newPrevious != nil && newPrevious.provisional && !newPrevious.locked && linked && evidence == 0 &&
		(finished.shown == typed || finished.provisional) && !finished.waitingForContext &&
		len(tokens(newPrevious.typed)) < 3:
		// A word that says nothing either way ("да", "5", another name) after a provisional
		// fix: keep them together, so a Russian word next can still take the fix back
		// ("Нурлан, Ерлан, привет").
		held := *newPrevious
		held.typed += t.separator + typed
		held.shown += t.separator + finished.shown
		t.previous = &held
	default:
		t.previous = &finished
	}
	t.separator = boundary
	t.current = ""
	t.currentLocked = false
	if strings.IndexFunc(boundary, isNewline) >= 0 {
		t.resetWindow()
		t.score = 0
	}

	if oldSpan == newSpan {
		return Action{}
	}
	t.lastEdit = pending
	return Action{Kind: Replace, Delete: utf8.RuneCountInString(oldSpan), Insert: newSpan, Expect: oldSpan}
}

func (t *Typing) undo(pending *edit) Action {
	t.lastEdit = nil
	// Back to what was on screen before, with the cursor at the end of the typed word.
	t.previous = copyWord(pending.previousBefore)
	if t.previous != nil && pending.original != pending.typed {
		t.previous.locked = true
	}
	t.separator = pending.separatorBefore
	t.current = pending.typed
	t.currentLocked = true
	left := lastRunes(pending.boundary, pending.boundaryLeft)
	return Action{Kind: Undo, Delete: utf8.RuneCountInString(pending.inserted) + utf8.RuneCountInString(left),
		Insert: pending.original, Expect: pending.inserted + left}
}

// eraseOne: an ordinary backspace, one character before the cursor goes.
func (t *Typing) eraseOne() {
	if t.current != "" {
		t.current = dropLastRunes(t.current, 1)
	} else if t.separator != "" {
		t.separator = dropLastRunes(t.separator, 1)
		if t.separator == "" && t.previous != nil {
			// Back inside the previous word (the last one, if several are kept together).
			previous := t.previous
			held := !previous.locked && strings.IndexFunc(previous.typed, isSeparatorLetter) >= 0
			if held {
				t.current = lastToken(previous.shown)
			} else {
				t.current = previous.shown
			}
			t.currentLocked = previous.locked
			t.previous = nil
		}
	}
}

func (t *Typing) resetWindow() {
	t.previous = nil
	t.separator = ""
	t.current = ""
	t.currentLocked = false
	t.lastEdit = nil
}

// MARK: - Deciding one word

type decision struct {
	shown             string
	waitingForContext bool
	provisional       bool
	// +2 Kazakh letters, +1 a Kazakh-only word, -1 mostly Russian, -2 a word only Russian knows.
	evidence int
	known    bool
}

// decide: how typed should look, given the word before it and the phrase language so far.
func (t *Typing) decide(typed, before string, hasBefore bool, boundary string, language Language,
	vocabulary *Vocabulary) decision {
	if strings.IndexFunc(typed, isCyrillicLetter) < 0 {
		return decision{shown: typed}
	}
	plain := Plain(typed)
	lower := strings.ToLower(typed)
	isRussian := vocabulary.Russian != nil && vocabulary.Russian.Contains(lower)
	asked := strings.Contains(boundary, "?")

	// A particle typed on its own takes the form the previous word calls for ("шынымен бе").
	if hasBefore && language != Russian && strings.IndexFunc(before, isCyrillicLetter) >= 0 {
		if form, ok := separateParticle(plain, before); ok {
			return decision{shown: adapt(form, typed), evidence: 2, known: true}
		}
	}

	if entry := vocabulary.entry(plain); entry != nil {
		// An abbreviation ("ӘКК", "АҚШ") only when typed in capitals: in a chat "акк" is an
		// account and "сим" a SIM card.
		if entry.IsAcronym && !(utf8.RuneCountInString(typed) > 1 && typed == strings.ToUpper(typed)) {
			return decision{shown: typed, known: true}
		}
		// Written together even in the corpus ("білесіңбе", "жоқпа"), or asked with "?" ("барма?").
		if !entry.IsTrusted && language != Russian {
			if split, ok := splitParticle(plain, typed, entry.Forms[0].Count, asked, language, vocabulary); ok {
				return decision{shown: split, evidence: 2, known: true}
			}
		}

		form := entry.best()
		byContext := false
		chosen, found := "", false
		if hasBefore {
			chosen, found = vocabulary.LeftContext[strings.ToLower(before)+"\t"+plain]
		}
		if found {
			form, byContext = chosen, true
		} else {
			if second, ok := preferSecondPerson(entry, before, hasBefore, vocabulary); ok {
				form = second
			}
			// The typed spelling is itself a common Kazakh word ("сал"): keep it.
			if !entry.IsTrusted && !entry.IsChat && form != lower && !differsOnlyByPerson(form, lower) &&
				entry.countOf(lower) >= max(1, entry.Forms[0].Count*15/100) {
				form = lower
			}
		}
		ambiguousWithRussian := entry.IsRussianToo || entry.IsName || entry.IsAcronym
		// Fixed before the language is known, and taken back if a Russian word follows
		// ("Нурлан, ты где?"): names, which say nothing about the language, and words Russian
		// has only as rare ones ("аска"). A name Russian also has as a word ("Канат", "Мука")
		// only with a capital and if the Kazakh name is common.
		tentative := language == Unknown && !entry.IsTrusted && !byContext && (entry.IsMostlyKazakh || entry.IsAcronym ||
			entry.IsName && (!entry.IsRussianToo || startsUpper(typed) && entry.Forms[0].Count >= 50))
		waits := ambiguousWithRussian && !entry.IsTrusted && !byContext && language != Kazakh && !tentative
		shown := typed
		if !waits {
			shown = adapt(form, typed)
		}
		provisional := shown != typed && (isRussian || tentative) && language == Unknown && !byContext && !entry.IsTrusted
		evidence := 1
		switch {
		case provisional || tentative:
			evidence = 0 // a guess about the language can't count as a vote for it ("Сени как дела")
		case ContainsKazakhLetters(shown):
			evidence = 2
		case entry.IsMostlyRussian:
			evidence = -1
		case ambiguousWithRussian || isRussian:
			evidence = 0
		}
		return decision{shown: shown, waitingForContext: waits && form != lower, provisional: provisional,
			evidence: evidence, known: true}
	}

	// Not a word the corpus knows. Russian words are left alone.
	if isRussian {
		return decision{shown: typed, evidence: -2, known: true}
	}
	if split, ok := splitParticle(plain, typed, 0, asked, language, vocabulary); ok {
		return decision{shown: split, evidence: 2, known: true}
	}
	capitalized := startsUpper(typed)
	if g, ok := compose(plain, capitalized, language, vocabulary); ok && g.text != lower {
		// A name with a Kazakh ending ("Узынагашка"), but the phrase may still turn out
		// Russian ("Нурланов сказал").
		provisional := g.fromName && language == Unknown
		evidence := 2
		if provisional {
			evidence = 0
		}
		return decision{shown: adapt(g.text, typed), provisional: provisional, evidence: evidence, known: true}
	}
	// Written small ("узынагашка"), a name waits for the phrase to show it is Kazakh.
	if language == Unknown && !capitalized {
		if g, ok := compose(plain, false, Kazakh, vocabulary); ok && g.fromName && g.text != lower {
			return decision{shown: typed, waitingForContext: true, known: true}
		}
	}
	if ContainsKazakhLetters(typed) {
		return decision{shown: typed, evidence: 2}
	}
	return decision{shown: typed, evidence: -1}
}

// usableStem: a stem good enough to build on: a Kazakh word, not a Russian look-alike. A name
// only with a Kazakh case ending ("Ұзынағаш" + "қа"), and written small only inside a Kazakh phrase.
func usableStem(entry *Entry, ending string, capitalized bool, language Language) bool {
	if entry.IsName && (!capitalized && language != Kazakh || !nameEndings[ending]) {
		return false
	}
	return !entry.IsRussianToo || language == Kazakh || ContainsKazakhLetters(entry.best())
}

// splitParticle: "айтасынба" → "айтасың ба", "биргана" → "бір ғана": a particle written
// together with its word. mergedCount > 0 when the merged spelling itself is in the corpus:
// then the stem must be far more common, or the word must end with a question mark.
func splitParticle(plain, typed string, mergedCount int, asked bool, language Language,
	vocabulary *Vocabulary) (string, bool) {
	for _, p := range attachable {
		length := utf8.RuneCountInString(p.plain)
		if !strings.HasSuffix(plain, p.plain) || utf8.RuneCountInString(plain)-length < 3 {
			continue
		}
		stemPlain := dropLastRunes(plain, length)
		stemTyped := dropLastRunes(typed, length)
		capitalized := startsUpper(typed)
		var stem string
		if entry := vocabulary.entry(stemPlain); entry != nil && (!entry.IsName || capitalized || language == Kazakh) {
			// A merged spelling the corpus knows well is a word of its own ("барма": don't go).
			if mergedCount > 0 && !(asked && p.isQuestion) && (mergedCount >= 20 || entry.Forms[0].Count < mergedCount*5) {
				continue
			}
			stem = entry.best()
			if second, ok := preferSecondPerson(entry, "", false, vocabulary); ok {
				stem = second
			}
		} else if mergedCount == 0 {
			g, ok := compose(stemPlain, capitalized, language, vocabulary)
			if !ok {
				continue
			}
			stem = g.text
		} else {
			continue
		}
		if !p.fits(stem) && !(asked && p.isQuestion && mergedCount == 0) {
			continue
		}
		// Each part keeps its own case: "САЛЕМЕТСИЗБЕ" → "СӘЛЕМЕТСІЗ БЕ".
		return adapt(stem, stemTyped) + " " + adapt(p.form(stem), lastRunes(typed, length)), true
	}
	return "", false
}

type guess struct {
	text string
	// Built on a name ("Ұзынағаш" + "қа").
	fromName bool
}

// compose: a word the corpus lacks: a known stem plus a common ending ("көрсет" + "ші").
func compose(plain string, capitalized bool, language Language, vocabulary *Vocabulary) (guess, bool) {
	letters := []rune(plain)
	if len(letters) < 5 || language == Russian {
		return guess{}, false
	}
	for cut := len(letters) - 1; cut >= 3; cut-- {
		stemPlain, endingPlain := string(letters[:cut]), string(letters[cut:])
		entry := vocabulary.entry(stemPlain)
		if entry == nil || !usableStem(entry, endingPlain, capitalized, language) {
			continue
		}
		stem := entry.best()
		endings := vocabulary.Endings[harmony(stem)+"\t"+endingPlain]
		if len(endings) == 0 {
			continue
		}
		ending := endings[0].Text
		for _, second := range endings[1:] {
			if differsOnlyByPerson(ending, second.Text) && isSecondPersonEnding(stem+second.Text) &&
				second.Count*20 >= endings[0].Count {
				ending = second.Text
				break
			}
		}
		if !fitsHarmony(stem, ending) {
			continue
		}
		return guess{text: stem + ending, fromName: entry.IsName}, true
	}
	return guess{}, false
}

// preferSecondPerson: chat is mostly "you": between "-ың/-ің" (your) and "-ын/-ін" (his) take
// "-ың" unless the previous word points at a third person ("оның").
func preferSecondPerson(entry *Entry, before string, hasBefore bool, vocabulary *Vocabulary) (string, bool) {
	best := entry.Forms[0]
	if len(entry.Forms) <= 1 || utf8.RuneCountInString(best.Text) < 4 {
		return "", false
	}
	if hasBefore && thirdPersonCues[strings.ToLower(before)] {
		return "", false
	}
	threshold := max(5, best.Count/100)
	for _, form := range entry.Forms[1:] {
		if form.Count < threshold {
			continue
		}
		if differsOnlyByPerson(best.Text, form.Text) && isSecondPersonEnding(form.Text) &&
			strings.Count(form.Text, "ң") > strings.Count(best.Text, "ң") &&
			!isWordOfItsOwn(best.Text, form.Text, vocabulary) {
			return form.Text, true
		}
	}
	return "", false
}

// isWordOfItsOwn: "Ойын" is a game, not "ой" + "-ын" (his thought): the corpus has its plural
// "ойындар", and a possessive ending never takes one.
func isWordOfItsOwn(word, other string, vocabulary *Vocabulary) bool {
	a, b := []rune(word), []rune(other)
	index := -1
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			index = i
			break
		}
	}
	if index < 0 {
		return false
	}
	stem := string(a[:index+1])
	for _, plural := range []string{"дар", "дер", "тар", "тер", "лар", "лер"} {
		if entry := vocabulary.entry(Plain(stem + plural)); entry != nil && entry.countOf(stem+plural) > 0 {
			return true
		}
	}
	return false
}

// adapt gives form the typed word's capitalization.
func adapt(form, typed string) string {
	letters := []rune{}
	for _, letter := range typed {
		if unicode.IsLetter(letter) {
			letters = append(letters, letter)
		}
	}
	text := string(letters)
	if len(letters) > 1 && text == strings.ToUpper(text) && text != strings.ToLower(text) {
		return strings.ToUpper(form)
	}
	if startsUpper(typed) && form != "" {
		first, size := utf8.DecodeRuneInString(form)
		return string(unicode.ToUpper(first)) + form[size:]
	}
	return form
}

// MARK: - Helpers

func copyWord(w *word) *word {
	if w == nil {
		return nil
	}
	copied := *w
	return &copied
}

func startsUpper(text string) bool {
	first, _ := utf8.DecodeRuneInString(text)
	return text != "" && unicode.IsUpper(first)
}

func allLettersOrNumbers(text string) bool {
	for _, letter := range text {
		if !unicode.IsLetter(letter) && !unicode.IsNumber(letter) {
			return false
		}
	}
	return true
}

func allSpacesOrPunctuation(text string) bool {
	for _, letter := range text {
		if letter != ' ' && !unicode.IsPunct(letter) {
			return false
		}
	}
	return true
}

func isSeparatorLetter(letter rune) bool {
	return !unicode.IsLetter(letter) && !unicode.IsNumber(letter)
}

func tokens(text string) []string {
	return strings.FieldsFunc(text, isSeparatorLetter)
}

func lastToken(text string) string {
	if words := tokens(text); len(words) > 0 {
		return words[len(words)-1]
	}
	return text
}

func dropLastRunes(text string, count int) string {
	for ; count > 0 && text != ""; count-- {
		_, size := utf8.DecodeLastRuneInString(text)
		text = text[:len(text)-size]
	}
	return text
}

func lastRunes(text string, count int) string {
	letters := []rune(text)
	if count >= len(letters) {
		return text
	}
	return string(letters[len(letters)-count:])
}
