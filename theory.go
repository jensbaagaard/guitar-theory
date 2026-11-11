package main

import (
	"fmt"
	"slices"
	"strings"
)

const letters = "CDEFGAB"

var (
	notes         = []string{"A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#"}
	stringNames   = [6]string{"e", "B", "G", "D", "A", "E"}
	openPitch     = [6]int{31, 26, 22, 17, 12, 7}
	intervalNames = []string{"P1", "m2", "M2", "m3", "M3", "P4", "TT", "P5", "m6", "M6", "m7", "M7", "P8"}
	intervalKeys  = map[string]int{"m2": 1, "2": 2, "m3": 3, "3": 4, "4": 5, "t": 6, "5": 7, "m6": 8, "6": 9, "m7": 10, "7": 11, "8": 12}

	chordRoots = []string{"C", "C#", "Db", "D", "D#", "Eb", "E", "F", "F#", "Gb", "G", "G#", "Ab", "A", "A#", "Bb", "B"}
	toneNames  = []string{"3rd", "5th", "7th"}
	chordTypes = []struct {
		symbol, name string
		tones        []int
	}{
		{"", "major", []int{4, 7}},
		{"m", "minor", []int{3, 7}},
		{"dim", "diminished", []int{3, 6}},
		{"aug", "augmented", []int{4, 8}},
		{"maj7", "major 7th", []int{4, 7, 11}},
		{"7", "dominant 7th", []int{4, 7, 10}},
		{"m7", "minor 7th", []int{3, 7, 10}},
		{"m7b5", "half-diminished", []int{3, 6, 10}},
	}

	majorKeys    = []string{"C", "G", "D", "A", "E", "B", "F#", "Db", "Ab", "Eb", "Bb", "F"}
	majorScale   = []int{0, 2, 4, 5, 7, 9, 11}
	numerals     = []string{"I", "ii", "iii", "IV", "V", "vi", "vii°"}
	degreeSuffix = []string{"", "m", "m", "", "", "m", "dim"}
)

type pos struct{ str, fret int }

func pitch(p pos) int {
	return openPitch[p.str] + p.fret
}

func noteAt(str, fret int) string {
	return notes[pitch(pos{str, fret})%12]
}

func noteFromKey(k string) (int, bool) {
	name := strings.ToUpper(k)
	if name == k {
		name += "#"
	}
	i := slices.Index(notes, name)
	return i, i >= 0
}

func pitchClass(name string) int {
	return (slices.Index(notes, name[:1]) + strings.Count(name, "#") - strings.Count(name, "b") + 12) % 12
}

func spell(root string, steps, semis int) string {
	i := (strings.IndexByte(letters, root[0]) + steps) % 7
	letter := letters[i : i+1]
	accidental := (pitchClass(root)+semis-pitchClass(letter)+18)%12 - 6
	if accidental < 0 {
		return letter + strings.Repeat("b", -accidental)
	}
	return letter + strings.Repeat("#", accidental)
}

func chordSpelling(root string, typ int) []string {
	spelling := []string{root}
	for i, semis := range chordTypes[typ].tones {
		spelling = append(spelling, spell(root, 2*(i+1), semis))
	}
	return spelling
}

func keyChords(key string) []string {
	chords := make([]string, len(majorScale))
	for degree, semis := range majorScale {
		chords[degree] = spell(key, degree, semis) + degreeSuffix[degree]
	}
	return chords
}

func fretboard(frets int, marks map[pos]string, names, numbers bool) string {
	var b strings.Builder
	if numbers {
		b.WriteString("  0 ")
		for f := 1; f <= frets; f++ {
			fmt.Fprintf(&b, "%3d   ", f)
		}
		b.WriteByte('\n')
	}
	for s, name := range stringNames {
		if !names {
			name = " "
		}
		open := marks[pos{s, 0}]
		if open == "" {
			open = " "
		}
		b.WriteString(name + " " + open + "║")
		for f := 1; f <= frets; f++ {
			mark := marks[pos{s, f}]
			if mark == "" {
				mark = "─"
			}
			b.WriteString("──" + mark + "──│")
		}
		b.WriteByte('\n')
	}
	b.WriteString("    ")
	for f := 1; f <= frets; f++ {
		switch f % 12 {
		case 0:
			b.WriteString(" • •  ")
		case 3, 5, 7, 9:
			b.WriteString("  •   ")
		default:
			b.WriteString("      ")
		}
	}
	return b.String()
}
