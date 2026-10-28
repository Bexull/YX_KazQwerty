// Command typetest types each line of a text file (and a space) through the autocorrect and
// prints what ends up on screen, one line per input line: to compare with the Mac app's
// engine or to try phrases by hand.
//
//	go run ./cmd/typetest ../Resources/kazakh phrases.txt
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"kazqwerty/kazakh"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: typetest DATA_DIR FILE")
		os.Exit(2)
	}
	vocabulary, err := kazakh.Load(os.DirFS(os.Args[1]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	vocabulary.Words = kazakh.Dictionary("", true)
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var typing kazakh.Typing
		screen := []rune{}
		for _, key := range line + " " {
			action := typing.Text(string(key), vocabulary)
			if action.Kind == kazakh.Replace {
				screen = append(screen[:max(0, len(screen)-action.Delete)], []rune(action.Insert)...)
			}
			screen = append(screen, key)
		}
		fmt.Fprintln(out, string(screen))
	}
}
