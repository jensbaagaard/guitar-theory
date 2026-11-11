package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
)

const (
	helpNoteKeys        = "a–g natural · shift+letter sharp"
	helpFlatKeys        = "a–g natural · shift+letter sharp (Bb = shift+a)"
	helpFindKeys        = "←/→ move · 0–9 fret · shift+1…0 frets 10–19 · enter confirm"
	danishShiftedDigits = `=!"#¤%&/()`
)

var fretKeys = func() map[string]int {
	keys := map[string]int{}
	for digit, symbol := range []rune(danishShiftedDigits) {
		shifted := 10 + (digit+9)%10
		keys[strconv.Itoa(digit)] = digit
		keys["shift+"+strconv.Itoa(digit)] = shifted
		keys[string(symbol)] = shifted
	}
	return keys
}()

type settings struct {
	enabled [6]bool
	lo, hi  int
	names   bool
	numbers bool
	inOrder bool
}

func (s settings) positions() []pos {
	var ps []pos
	for str, on := range s.enabled {
		for f := s.lo; on && f <= s.hi; f++ {
			ps = append(ps, pos{str, f})
		}
	}
	return ps
}

func (s settings) board(marks map[pos]string) string {
	b := fretboard(max(s.hi, 12), marks, s.names, s.numbers)
	b = strings.ReplaceAll(b, "●", markerStyle.Render("●"))
	return strings.ReplaceAll(b, "○", rootStyle.Render("○"))
}

type game interface {
	next(s settings) bool
	view(s settings) string
	key(k string, s settings) (answered, correct bool, feedback string)
}

var games = []struct {
	name, desc, help string
	fretboard        bool
	stringOrder      bool
	game             game
}{
	{"Name the note", "name the marked note", helpNoteKeys, true, false, &nameNote{}},
	{"Find the note", "move the marker to the note asked for", helpFindKeys, true, true, &findNote{}},
	{"Name the interval", "name the distance between two marked notes", "2–8 major/perfect · m then 2/3/6/7 minor · t tritone", true, false, &nameInterval{}},
	{"Chord tones", "name the 3rd, 5th or 7th of a chord", helpFlatKeys, false, false, &chordTone{}},
	{"Chords in a key", "name the chord on a degree of a major key", helpFlatKeys, false, false, &keyChord{}},
}

func pick[T comparable](options []T, prev T) T {
	for {
		o := options[rand.IntN(len(options))]
		if o != prev || len(options) == 1 {
			return o
		}
	}
}

type nameNote struct{ at pos }

func (g *nameNote) next(s settings) bool {
	ps := s.positions()
	if len(ps) == 0 {
		return false
	}
	g.at = pick(ps, g.at)
	return true
}

func (g *nameNote) view(s settings) string {
	return s.board(map[pos]string{g.at: "●"}) + "\n\nWhich note?"
}

func (g *nameNote) key(k string, _ settings) (bool, bool, string) {
	guess, ok := noteFromKey(k)
	if !ok {
		return false, false, ""
	}
	answer := noteAt(g.at.str, g.at.fret)
	if notes[guess] == answer {
		return true, true, answer
	}
	return true, false, fmt.Sprintf("%s — string %d, fret %d was %s", notes[guess], g.at.str+1, g.at.fret, answer)
}

type findNote struct {
	target   pos
	cursor   int
	note     string
	queue    []pos
	settings settings
}

func (g *findNote) next(s settings) bool {
	ps := s.positions()
	if len(ps) == 0 {
		return false
	}
	if s != g.settings {
		g.queue, g.settings = nil, s
	}
	if !s.inOrder {
		g.target = pick(ps, g.target)
	} else {
		if len(g.queue) == 0 {
			g.note, g.queue = stringSweep(ps, g.note)
		}
		g.target, g.queue = g.queue[0], g.queue[1:]
	}
	g.cursor = min(max(g.cursor, s.lo), s.hi)
	return true
}

func stringSweep(ps []pos, prev string) (string, []pos) {
	var names []string
	for _, p := range ps {
		if n := noteAt(p.str, p.fret); !slices.Contains(names, n) {
			names = append(names, n)
		}
	}
	note := pick(names, prev)
	var sweep []pos
	for _, p := range ps {
		visited := len(sweep) > 0 && sweep[len(sweep)-1].str == p.str
		if noteAt(p.str, p.fret) == note && !visited {
			sweep = append(sweep, p)
		}
	}
	slices.Reverse(sweep)
	return note, sweep
}

func (g *findNote) view(s settings) string {
	board := s.board(map[pos]string{{g.target.str, g.cursor}: "●"})
	return fmt.Sprintf("%s\n\nFind %s on this string", board, noteAt(g.target.str, g.target.fret))
}

func (g *findNote) key(k string, s settings) (bool, bool, string) {
	switch k {
	case "left", "h":
		g.cursor = max(g.cursor-1, s.lo)
	case "right", "l":
		g.cursor = min(g.cursor+1, s.hi)
	case "enter":
		want := noteAt(g.target.str, g.target.fret)
		got := noteAt(g.target.str, g.cursor)
		if got == want {
			return true, true, fmt.Sprintf("%s at fret %d", want, g.cursor)
		}
		return true, false, fmt.Sprintf("fret %d is %s — %s is at fret %d", g.cursor, got, want, g.target.fret)
	default:
		if fret, ok := fretKeys[k]; ok && fret >= s.lo && fret <= s.hi {
			g.cursor = fret
		}
	}
	return false, false, ""
}

type nameInterval struct {
	pair  [2]pos
	minor bool
}

func (g *nameInterval) next(s settings) bool {
	ps := s.positions()
	var pairs [][2]pos
	for _, root := range ps {
		for _, top := range ps {
			semis := pitch(top) - pitch(root)
			reach := max(top.fret-root.fret, root.fret-top.fret)
			if semis >= 1 && semis <= 12 && (root.str == top.str || reach <= 4) {
				pairs = append(pairs, [2]pos{root, top})
			}
		}
	}
	if len(pairs) == 0 {
		return false
	}
	g.pair, g.minor = pick(pairs, g.pair), false
	return true
}

func (g *nameInterval) prefix() string {
	if g.minor {
		return "m"
	}
	return ""
}

func (g *nameInterval) view(s settings) string {
	board := s.board(map[pos]string{g.pair[0]: "○", g.pair[1]: "●"})
	return fmt.Sprintf("%s\n\nWhich interval from ○ up to ●?  %s", board, g.prefix())
}

func (g *nameInterval) key(k string, _ settings) (bool, bool, string) {
	if k == "m" {
		g.minor = !g.minor
		return false, false, ""
	}
	guess, ok := intervalKeys[g.prefix()+k]
	if !ok {
		return false, false, ""
	}
	root, top := g.pair[0], g.pair[1]
	answer := pitch(top) - pitch(root)
	if guess == answer {
		return true, true, intervalNames[answer]
	}
	unit := "semitones"
	if answer == 1 {
		unit = "semitone"
	}
	return true, false, fmt.Sprintf("%s — that was %s, %d %s (string %d fret %d → string %d fret %d)", intervalNames[guess], intervalNames[answer], answer, unit, root.str+1, root.fret, top.str+1, top.fret)
}

type chordQuestion struct {
	root      string
	typ, tone int
}

var chordQuestions = func() []chordQuestion {
	var qs []chordQuestion
	for _, root := range chordRoots {
		for typ, c := range chordTypes {
			spelling := strings.Join(chordSpelling(root, typ), " ")
			if strings.Contains(spelling, "##") || strings.Contains(spelling, "bb") {
				continue
			}
			for tone := range c.tones {
				qs = append(qs, chordQuestion{root, typ, tone})
			}
		}
	}
	return qs
}()

type chordTone struct{ q chordQuestion }

func (g *chordTone) next(settings) bool {
	g.q = pick(chordQuestions, g.q)
	return true
}

func (g *chordTone) view(settings) string {
	c := chordTypes[g.q.typ]
	return fmt.Sprintf("Chord: %s %s (%s%s)\n\nWhich note is the %s?", g.q.root, c.name, g.q.root, c.symbol, toneNames[g.q.tone])
}

func (g *chordTone) key(k string, _ settings) (bool, bool, string) {
	guess, ok := noteFromKey(k)
	if !ok {
		return false, false, ""
	}
	chord := g.q.root + chordTypes[g.q.typ].symbol
	spelling := chordSpelling(g.q.root, g.q.typ)
	answer := spelling[g.q.tone+1]
	if guess == pitchClass(answer) {
		return true, true, fmt.Sprintf("%s = %s", chord, strings.Join(spelling, " "))
	}
	return true, false, fmt.Sprintf("%s — the %s of %s is %s (%s)", notes[guess], toneNames[g.q.tone], chord, answer, strings.Join(spelling, " "))
}

type keyQuestion struct {
	key    string
	degree int
}

var keyQuestions = func() []keyQuestion {
	var qs []keyQuestion
	for _, key := range majorKeys {
		for degree := 1; degree < len(majorScale); degree++ {
			qs = append(qs, keyQuestion{key, degree})
		}
	}
	return qs
}()

type keyChord struct{ q keyQuestion }

func (g *keyChord) next(settings) bool {
	g.q = pick(keyQuestions, g.q)
	return true
}

func (g *keyChord) view(settings) string {
	return fmt.Sprintf("Key of %s major\n\nWhich chord is %s?  (press its root note)", g.q.key, numerals[g.q.degree])
}

func (g *keyChord) key(k string, _ settings) (bool, bool, string) {
	guess, ok := noteFromKey(k)
	if !ok {
		return false, false, ""
	}
	chords := keyChords(g.q.key)
	numeral, chord := numerals[g.q.degree], chords[g.q.degree]
	if guess == pitchClass(spell(g.q.key, g.q.degree, majorScale[g.q.degree])) {
		return true, true, fmt.Sprintf("%s in %s major = %s", numeral, g.q.key, chord)
	}
	return true, false, fmt.Sprintf("%s — %s in %s major is %s (%s)", notes[guess], numeral, g.q.key, chord, strings.Join(chords, " "))
}
