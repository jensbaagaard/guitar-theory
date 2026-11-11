package main

import (
	"slices"
	"strings"
	"testing"
)

func TestNoteAt(t *testing.T) {
	cases := []struct {
		name      string
		str, fret int
		want      string
	}{
		{"open low E", 5, 0, "E"},
		{"low E fret 3", 5, 3, "G"},
		{"A string fret 2", 4, 2, "B"},
		{"A string fret 11", 4, 11, "G#"},
		{"D string fret 4", 3, 4, "F#"},
		{"G string fret 4", 2, 4, "B"},
		{"B string fret 1", 1, 1, "C"},
		{"high e fret 1", 0, 1, "F"},
		{"low E fret 12", 5, 12, "E"},
		{"high e fret 24", 0, 24, "E"},
	}
	for _, c := range cases {
		if got := noteAt(c.str, c.fret); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestSpell(t *testing.T) {
	cases := []struct {
		name         string
		root         string
		steps, semis int
		want         string
	}{
		{"major 3rd of A", "A", 2, 4, "C#"},
		{"minor 3rd of C", "C", 2, 3, "Eb"},
		{"4th of F", "F", 3, 5, "Bb"},
		{"major 7th of F#", "F#", 6, 11, "E#"},
		{"minor 3rd of Ab", "Ab", 2, 3, "Cb"},
	}
	for _, c := range cases {
		if got := spell(c.root, c.steps, c.semis); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestKeyChords(t *testing.T) {
	got := strings.Join(keyChords("F"), " ")

	if want := "F Gm Am Bb C Dm Edim"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestStringNamesSetting(t *testing.T) {
	if !strings.Contains(fretboard(12, nil, true, true), "E  ║") {
		t.Error("string names missing when enabled")
	}
	if strings.ContainsAny(fretboard(12, nil, false, true), "eBGDAE") {
		t.Error("string names shown when disabled")
	}
}

func TestFretNumbersSetting(t *testing.T) {
	if !strings.Contains(fretboard(12, nil, true, true), "12") {
		t.Error("fret numbers missing when enabled")
	}
	if strings.ContainsAny(fretboard(12, nil, true, false), "0123456789") {
		t.Error("fret numbers shown when disabled")
	}
}

func TestQuestionsStayWithinSelection(t *testing.T) {
	s := settings{enabled: [6]bool{1: true, 4: true}, lo: 5, hi: 7}
	g := &nameNote{}

	for range 200 {
		g.next(s)
		if !s.enabled[g.at.str] || g.at.fret < 5 || g.at.fret > 7 {
			t.Fatalf("got string %d fret %d, outside strings 2/5 frets 5-7", g.at.str+1, g.at.fret)
		}
	}
}

func TestNoQuestionWithoutStrings(t *testing.T) {
	g := &nameNote{}

	if g.next(settings{lo: 0, hi: 12}) {
		t.Error("got a question with no strings enabled")
	}
}

func TestNameNoteNaturalGuess(t *testing.T) {
	g := &nameNote{at: pos{5, 3}}

	answered, correct, _ := g.key("g", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestNameNoteShiftedLetterGuessesSharp(t *testing.T) {
	g := &nameNote{at: pos{3, 4}}

	answered, correct, _ := g.key("F", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestNameNoteWrongGuess(t *testing.T) {
	g := &nameNote{at: pos{5, 3}}

	answered, correct, _ := g.key("a", settings{})

	if !answered || correct {
		t.Errorf("got answered=%v correct=%v, want answered but wrong", answered, correct)
	}
}

func TestNameNoteIgnoresNonNoteKey(t *testing.T) {
	g := &nameNote{at: pos{5, 3}}

	answered, _, _ := g.key("B", settings{})

	if answered {
		t.Error("B# counted as an answer")
	}
}

func TestFindNoteAcceptsMarkerMovedOntoTheNote(t *testing.T) {
	s := settings{lo: 0, hi: 12}
	g := &findNote{target: pos{5, 3}, cursor: 2}

	g.key("right", s)
	answered, correct, _ := g.key("enter", s)

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestFindNoteAcceptsSameNoteAnOctaveUp(t *testing.T) {
	s := settings{lo: 0, hi: 12}
	g := &findNote{target: pos{5, 0}, cursor: 12}

	_, correct, _ := g.key("enter", s)

	if !correct {
		t.Error("fret 12 rejected for an open-string note")
	}
}

func TestFindNoteRejectsOtherFret(t *testing.T) {
	s := settings{lo: 0, hi: 12}
	g := &findNote{target: pos{5, 3}, cursor: 4}

	answered, correct, _ := g.key("enter", s)

	if !answered || correct {
		t.Errorf("got answered=%v correct=%v, want answered but wrong", answered, correct)
	}
}

func TestFindNoteMarkerStaysInFretRange(t *testing.T) {
	s := settings{lo: 3, hi: 5}
	g := &findNote{target: pos{5, 5}, cursor: 5}

	g.key("right", s)

	if g.cursor != 5 {
		t.Errorf("marker at fret %d, want 5", g.cursor)
	}
}

func TestFindNoteTypedFretMovesMarker(t *testing.T) {
	s := settings{lo: 0, hi: 12}
	g := &findNote{target: pos{5, 3}, cursor: 9}

	g.key("3", s)
	answered, correct, _ := g.key("enter", s)

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestFindNotePlainDigitsDoNotCombine(t *testing.T) {
	s := settings{lo: 0, hi: 12}
	g := &findNote{target: pos{5, 3}, cursor: 0}

	g.key("1", s)
	g.key("2", s)

	if g.cursor != 2 {
		t.Errorf("marker at fret %d, want 2", g.cursor)
	}
}

func TestFindNoteShiftedDigitsCountOnFromTen(t *testing.T) {
	cases := []struct {
		key  string
		want int
	}{
		{"!", 10},
		{`"`, 11},
		{"#", 12},
		{"¤", 13},
		{"%", 14},
		{"&", 15},
		{"/", 16},
		{"(", 17},
		{")", 18},
		{"=", 19},
		{"shift+1", 10},
		{"shift+0", 19},
	}
	for _, c := range cases {
		g := &findNote{target: pos{5, 3}, cursor: 0}

		g.key(c.key, settings{lo: 0, hi: 24})

		if g.cursor != c.want {
			t.Errorf("%s: marker at fret %d, want %d", c.key, g.cursor, c.want)
		}
	}
}

func TestFindNoteTypedFretOutsideRangeIsIgnored(t *testing.T) {
	s := settings{lo: 3, hi: 5}
	g := &findNote{target: pos{5, 5}, cursor: 4}

	g.key("9", s)

	if g.cursor != 4 {
		t.Errorf("marker at fret %d, want 4", g.cursor)
	}
}

func TestFindNoteInOrderAsksOneNoteOnEveryStringLowToHigh(t *testing.T) {
	s := settings{enabled: [6]bool{true, true, true, true, true, true}, lo: 0, hi: 12, inOrder: true}
	g := &findNote{}

	var strs []int
	asked := map[string]bool{}
	for range 6 {
		g.next(s)
		strs = append(strs, g.target.str)
		asked[noteAt(g.target.str, g.target.fret)] = true
	}

	if want := []int{5, 4, 3, 2, 1, 0}; !slices.Equal(strs, want) {
		t.Errorf("got strings %v, want %v", strs, want)
	}
	if len(asked) != 1 {
		t.Errorf("got %d different notes in one pass, want 1", len(asked))
	}
}

func TestFindNoteInOrderSkipsDisabledStrings(t *testing.T) {
	s := settings{enabled: [6]bool{1: true, 4: true}, lo: 0, hi: 12, inOrder: true}
	g := &findNote{}

	g.next(s)
	first := g.target.str
	g.next(s)
	second := g.target.str

	if first != 4 || second != 1 {
		t.Errorf("got strings %d then %d, want 4 then 1", first, second)
	}
}

func TestFindNoteInOrderMovesToNewNoteAfterLastString(t *testing.T) {
	s := settings{enabled: [6]bool{1: true, 4: true}, lo: 0, hi: 12, inOrder: true}
	g := &findNote{}

	g.next(s)
	before := noteAt(g.target.str, g.target.fret)
	g.next(s)
	g.next(s)
	after := noteAt(g.target.str, g.target.fret)

	if before == after {
		t.Errorf("still asking for %s after covering every string", after)
	}
}

func TestNameIntervalPerfectFourth(t *testing.T) {
	g := &nameInterval{pair: [2]pos{{5, 3}, {4, 3}}}

	answered, correct, _ := g.key("4", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestNameIntervalMinorPrefix(t *testing.T) {
	g := &nameInterval{pair: [2]pos{{5, 3}, {4, 1}}}

	g.key("m", settings{})
	answered, correct, _ := g.key("3", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestNameIntervalTritone(t *testing.T) {
	g := &nameInterval{pair: [2]pos{{5, 3}, {4, 4}}}

	answered, correct, _ := g.key("t", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
}

func TestNameIntervalWrongGuess(t *testing.T) {
	g := &nameInterval{pair: [2]pos{{5, 3}, {4, 3}}}

	answered, correct, feedback := g.key("t", settings{})

	if !answered || correct {
		t.Errorf("got answered=%v correct=%v, want answered but wrong", answered, correct)
	}
	if want := "TT — that was P4, 5 semitones (string 6 fret 3 → string 5 fret 3)"; feedback != want {
		t.Errorf("got feedback %q, want %q", feedback, want)
	}
}

func TestNameIntervalNeedsTwoPositions(t *testing.T) {
	g := &nameInterval{}

	if g.next(settings{enabled: [6]bool{5: true}, lo: 3, hi: 3}) {
		t.Error("got an interval question from a single position")
	}
}

func TestChordToneFlatAnsweredWithSharpKey(t *testing.T) {
	g := &chordTone{q: chordQuestion{root: "C", typ: 1, tone: 0}}

	answered, correct, feedback := g.key("D", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
	if want := "Cm = C Eb G"; feedback != want {
		t.Errorf("got feedback %q, want %q", feedback, want)
	}
}

func TestChordToneWrongGuess(t *testing.T) {
	g := &chordTone{q: chordQuestion{root: "C", typ: 1, tone: 0}}

	answered, correct, _ := g.key("e", settings{})

	if !answered || correct {
		t.Errorf("got answered=%v correct=%v, want answered but wrong", answered, correct)
	}
}

func TestKeyChordAnsweredWithRoot(t *testing.T) {
	g := &keyChord{q: keyQuestion{key: "Eb", degree: 5}}

	answered, correct, feedback := g.key("c", settings{})

	if !answered || !correct {
		t.Errorf("got answered=%v correct=%v, want both true", answered, correct)
	}
	if want := "vi in Eb major = Cm"; feedback != want {
		t.Errorf("got feedback %q, want %q", feedback, want)
	}
}

func TestKeyChordWrongGuess(t *testing.T) {
	g := &keyChord{q: keyQuestion{key: "Eb", degree: 5}}

	answered, correct, _ := g.key("d", settings{})

	if !answered || correct {
		t.Errorf("got answered=%v correct=%v, want answered but wrong", answered, correct)
	}
}
