package kazakh

import (
	"bytes"
	"io/fs"
	"strconv"
	"strings"
)

// Form is one way a plain spelling can be written in Kazakh, with how often the corpus saw it.
type Form struct {
	Text  string
	Count int
}

// Entry is what a word typed with Russian letters may stand for.
type Entry struct {
	// Forms, most frequent first. May include the plain spelling itself ("туыс", "сал").
	Forms []Form
	// The plain spelling is also a Russian word ("тур"): only correct it inside Kazakh text.
	IsRussianToo bool
	// A name ("Айгерим"): Russian text keeps the Russian spelling.
	IsName bool
	// The plain spelling is far more common in Russian ("для", "он"): a vote for Russian.
	IsMostlyRussian bool
	// Russian has the plain spelling only as a rare word next to a common Kazakh one
	// ("аска": асқа): correct it without waiting for Kazakh text.
	IsMostlyKazakh bool
	// Nearly always written in capitals ("ӘКК", "АҚШ"): an abbreviation.
	IsAcronym bool
	// The best form is a hand-picked chat word ("тұрасың"): always worth correcting to.
	IsChat bool
	// Hand-picked, the user's own or learned word: always applies.
	IsTrusted bool
}

func (e *Entry) best() string { return e.Forms[0].Text }

func (e *Entry) countOf(text string) int {
	for _, form := range e.Forms {
		if form.Text == text {
			return form.Count
		}
	}
	return 0
}

// Lexicon finds the entry for a plain spelling.
type Lexicon interface {
	Entry(plain string) *Entry
}

// Vocabulary is everything autocorrect knows. Treat it as read-only once in use: to change
// the word list, make a copy (it is shared with the keyboard hook).
type Vocabulary struct {
	// Plain spelling -> word: hand-picked list, the user's words, learned words.
	Words  map[string]string
	Corpus Lexicon
	// "prev<TAB>plain" -> form chosen after that word; "plain<TAB>next" -> form chosen before it.
	LeftContext  map[string]string
	RightContext map[string]string
	// Plain "first<TAB>second" -> "first-second" or "firstsecond"; "first<TAB>secoX*" -> "-"
	// for whole families ("бір" + "бір…" always takes a hyphen).
	Pairs map[string]string
	// "<b|f|-><TAB>plain ending" -> endings, most frequent first ("b" / "f": back / front stem).
	Endings map[string][]Form
	// Russian words: never guessed as Kazakh.
	Russian *SortedLines
}

// CorpusSize is the number of plain spellings in the corpus, for diagnostics.
func (v *Vocabulary) CorpusSize() int {
	if corpus, ok := v.Corpus.(*Corpus); ok {
		return corpus.lines.Count()
	}
	return 0
}

// WithWords returns a copy that uses another hand-picked / user / learned word list.
func (v *Vocabulary) WithWords(words map[string]string) *Vocabulary {
	copied := *v
	copied.Words = words
	return &copied
}

func (v *Vocabulary) entry(plain string) *Entry {
	if word, ok := v.Words[plain]; ok {
		return &Entry{Forms: []Form{{Text: word, Count: 1_000_000}}, IsTrusted: true}
	}
	if v.Corpus == nil {
		return nil
	}
	return v.Corpus.Entry(plain)
}

// joined: how two typed words should be joined, if they belong together.
func (v *Vocabulary) joined(first, second, firstShown, secondShown string) (string, bool) {
	firstPlain, secondPlain := Plain(first), Plain(second)
	if pair, ok := v.Pairs[firstPlain+"\t"+secondPlain]; ok {
		return pair, true
	}
	if letters := []rune(secondPlain); len(letters) >= 4 && v.Pairs[firstPlain+"\t"+string(letters[:4])+"*"] == "-" {
		return strings.ToLower(firstShown) + "-" + strings.ToLower(secondShown), true
	}
	return "", false
}

// SortedLines is a text file of lines sorted by their first tab-separated field (UTF-8
// bytes), searched by binary search.
type SortedLines struct {
	data   []byte
	starts []int32
}

// NewSortedLines indexes the lines of data.
func NewSortedLines(data []byte) *SortedLines {
	lines := &SortedLines{data: data, starts: make([]int32, 0, len(data)/20)}
	lineStart := 0
	for i, b := range data {
		if b == '\n' {
			if i > lineStart {
				lines.starts = append(lines.starts, int32(lineStart))
			}
			lineStart = i + 1
		}
	}
	if lineStart < len(data) {
		lines.starts = append(lines.starts, int32(lineStart))
	}
	return lines
}

// Count is the number of lines.
func (s *SortedLines) Count() int { return len(s.starts) }

// Lookup returns the rest of the line after key and its tab, if a line starts with key.
func (s *SortedLines) Lookup(key string) (string, bool) {
	low, high := 0, len(s.starts)-1
	for low <= high {
		middle := (low + high) / 2
		start := int(s.starts[middle])
		order := s.compare(key, start)
		if order == 0 {
			from := start + len(key)
			if from < len(s.data) && s.data[from] == '\t' {
				from++
			}
			end := from
			for end < len(s.data) && s.data[end] != '\n' {
				end++
			}
			return string(s.data[from:end]), true
		}
		if order < 0 {
			high = middle - 1
		} else {
			low = middle + 1
		}
	}
	return "", false
}

// Contains reports whether a line is exactly word (or starts with word and a tab).
func (s *SortedLines) Contains(word string) bool {
	_, ok := s.Lookup(word)
	return ok
}

// compare compares key with the first field of the line starting at start.
func (s *SortedLines) compare(key string, start int) int {
	for i := 0; ; i++ {
		position := start + i
		lineByte := byte('\n')
		if position < len(s.data) {
			lineByte = s.data[position]
		}
		lineEnded := lineByte == '\t' || lineByte == '\n'
		if i == len(key) {
			if lineEnded {
				return 0
			}
			return -1
		}
		if lineEnded {
			return 1
		}
		if key[i] != lineByte {
			if key[i] < lineByte {
				return -1
			}
			return 1
		}
	}
}

// Corpus is the word-form list built by Tools/make-kazakh-dictionary.py: one line per plain
// spelling, "plain<TAB>form:count|form:count<TAB>flags", sorted by plain spelling (UTF-8
// bytes). "=" is the plain spelling itself; flags: r, R, k, n, a, c, s (see the builder).
type Corpus struct {
	lines *SortedLines
}

// Entry implements Lexicon.
func (c *Corpus) Entry(plain string) *Entry {
	rest, ok := c.lines.Lookup(plain)
	if !ok {
		return nil
	}
	fields := strings.Split(rest, "\t")
	forms := parseForms(fields[0], plain)
	if len(forms) == 0 {
		return nil
	}
	flags := ""
	if len(fields) > 1 {
		flags = fields[1]
	}
	return &Entry{
		Forms:           forms,
		IsRussianToo:    strings.Contains(flags, "r"),
		IsName:          strings.Contains(flags, "n"),
		IsMostlyRussian: strings.Contains(flags, "R"),
		IsMostlyKazakh:  strings.Contains(flags, "k"),
		IsAcronym:       strings.Contains(flags, "a"),
		IsChat:          strings.Contains(flags, "c"),
	}
}

// parseForms: "form:count|=:count" -> forms; "=" is the plain spelling.
func parseForms(field, plain string) []Form {
	var forms []Form
	for _, item := range strings.Split(field, "|") {
		colon := strings.LastIndex(item, ":")
		if item == "" || colon < 0 {
			continue
		}
		count, err := strconv.Atoi(item[colon+1:])
		if err != nil {
			continue
		}
		text := item[:colon]
		if text == "=" {
			text = plain
		}
		forms = append(forms, Form{Text: text, Count: count})
	}
	return forms
}

// Load reads everything from a folder of files built by Tools/make-kazakh-dictionary.py.
func Load(folder fs.FS) (*Vocabulary, error) {
	vocabulary := &Vocabulary{
		Words:        map[string]string{},
		LeftContext:  map[string]string{},
		RightContext: map[string]string{},
		Pairs:        map[string]string{},
		Endings:      map[string][]Form{},
	}
	dictionary, err := readFile(folder, "dictionary.tsv")
	if err != nil {
		return nil, err
	}
	vocabulary.Corpus = &Corpus{lines: NewSortedLines(dictionary)}
	if russian, err := readFile(folder, "russian.txt"); err == nil {
		vocabulary.Russian = NewSortedLines(russian)
	}
	forEachLine(folder, "context.tsv", func(fields []string) {
		if len(fields) != 4 {
			return
		}
		switch fields[0] {
		case "L":
			vocabulary.LeftContext[fields[1]+"\t"+fields[2]] = fields[3]
		case "R":
			vocabulary.RightContext[fields[1]+"\t"+fields[2]] = fields[3]
		}
	})
	forEachLine(folder, "pairs.tsv", func(fields []string) {
		if len(fields) == 3 {
			vocabulary.Pairs[fields[0]+"\t"+fields[1]] = fields[2]
		}
	})
	forEachLine(folder, "endings.tsv", func(fields []string) {
		if len(fields) == 3 {
			vocabulary.Endings[fields[0]+"\t"+fields[1]] = parseForms(fields[2], fields[1])
		}
	})
	return vocabulary, nil
}

// readFile reads a data file with LF line ends, also when a Windows checkout gave it CRLF.
func readFile(folder fs.FS, name string) ([]byte, error) {
	data, err := fs.ReadFile(folder, name)
	if err == nil && bytes.IndexByte(data, '\r') >= 0 {
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	}
	return data, err
}

func forEachLine(folder fs.FS, name string, body func([]string)) {
	data, err := readFile(folder, name)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			body(strings.Split(line, "\t"))
		}
	}
}
