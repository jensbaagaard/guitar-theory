package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	maxFret    = 24
	fromRow    = 6
	toRow      = 7
	namesRow   = 8
	numbersRow = 9
	orderRow   = 10
	helpMenu   = "↑/↓ move · enter select · q quit"
	helpSetup  = "↑/↓ move · space toggle · ←/→ change fret · enter start · esc games · q quit"
	noQuestion = "Nothing to ask with this selection — enable more strings or widen the fret range"
)

const (
	menuScreen = iota
	setupScreen
	playScreen
)

var (
	pageStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle  = lipgloss.NewStyle().Bold(true)
	markerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	rootStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	badStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle    = lipgloss.NewStyle().Faint(true)
)

type model struct {
	screen   int
	choice   int
	cursor   int
	settings settings
	right    int
	total    int
	feedback string
	err      string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	k := key.String()
	if k == "ctrl+c" || k == "q" {
		return m, tea.Quit
	}
	switch m.screen {
	case menuScreen:
		return m.updateMenu(k), nil
	case setupScreen:
		return m.updateSetup(k), nil
	default:
		return m.updatePlay(k), nil
	}
}

func (m model) updateMenu(k string) model {
	switch k {
	case "up", "k":
		m.choice = max(m.choice-1, 0)
	case "down", "j":
		m.choice = min(m.choice+1, len(games)-1)
	case "enter":
		if games[m.choice].fretboard {
			m.screen, m.cursor = setupScreen, 0
			return m
		}
		return m.start()
	}
	return m
}

func (m model) updateSetup(k string) model {
	m.err = ""
	s := &m.settings
	switch k {
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, m.lastRow())
	case "space":
		switch {
		case m.cursor < fromRow:
			s.enabled[m.cursor] = !s.enabled[m.cursor]
		case m.cursor == namesRow:
			s.names = !s.names
		case m.cursor == numbersRow:
			s.numbers = !s.numbers
		case m.cursor == orderRow:
			s.inOrder = !s.inOrder
		}
	case "left", "h":
		m.adjustFret(-1)
	case "right", "l":
		m.adjustFret(1)
	case "enter":
		return m.start()
	case "esc":
		m.screen = menuScreen
	}
	return m
}

func (m model) lastRow() int {
	if games[m.choice].stringOrder {
		return orderRow
	}
	return numbersRow
}

func (m *model) adjustFret(d int) {
	s := &m.settings
	switch m.cursor {
	case fromRow:
		s.lo = min(max(s.lo+d, 0), s.hi)
	case toRow:
		s.hi = min(max(s.hi+d, s.lo), maxFret)
	}
}

func (m model) start() model {
	m.right, m.total, m.feedback = 0, 0, ""
	if games[m.choice].game.next(m.settings) {
		m.screen = playScreen
	} else {
		m.err = noQuestion
	}
	return m
}

func (m model) updatePlay(k string) model {
	g := games[m.choice]
	if k == "esc" {
		m.screen = menuScreen
		if g.fretboard {
			m.screen = setupScreen
		}
		return m
	}
	answered, correct, feedback := g.game.key(k, m.settings)
	if !answered {
		return m
	}
	m.total++
	if correct {
		m.right++
		m.feedback = okStyle.Render("✓ " + feedback)
	} else {
		m.feedback = badStyle.Render("✗ " + feedback)
	}
	g.game.next(m.settings)
	return m
}

func (m model) View() tea.View {
	title := "Guitar Theory · " + games[m.choice].name
	var body string
	switch m.screen {
	case menuScreen:
		title, body = "Guitar Theory", m.menuView()
	case setupScreen:
		body = m.setupView()
	default:
		body = m.playView()
	}
	v := tea.NewView(pageStyle.Render(titleStyle.Render(title) + "\n\n" + body))
	v.AltScreen = true
	return v
}

func pointer(on bool) string {
	if on {
		return ">"
	}
	return " "
}

func checkbox(on bool) string {
	if on {
		return "[x]"
	}
	return "[ ]"
}

func (m model) menuView() string {
	var b strings.Builder
	for i, g := range games {
		fmt.Fprintf(&b, "%s %-19s%s\n", pointer(m.choice == i), g.name, dimStyle.Render(g.desc))
	}
	b.WriteString("\n" + dimStyle.Render(helpMenu))
	return b.String()
}

func (m model) setupView() string {
	s := m.settings
	var b strings.Builder
	b.WriteString("  Strings\n")
	for i, name := range stringNames {
		fmt.Fprintf(&b, "%s %s %d %s\n", pointer(m.cursor == i), checkbox(s.enabled[i]), i+1, name)
	}
	fmt.Fprintf(&b, "\n%s From fret  %d\n%s To fret    %d\n", pointer(m.cursor == fromRow), s.lo, pointer(m.cursor == toRow), s.hi)
	fmt.Fprintf(&b, "\n  Show\n%s %s String names\n%s %s Fret numbers\n\n", pointer(m.cursor == namesRow), checkbox(s.names), pointer(m.cursor == numbersRow), checkbox(s.numbers))
	if games[m.choice].stringOrder {
		order := "random"
		if s.inOrder {
			order = "one by one, low to high"
		}
		fmt.Fprintf(&b, "%s String order  %s\n\n", pointer(m.cursor == orderRow), order)
	}
	if m.err != "" {
		b.WriteString(badStyle.Render(m.err) + "\n")
	}
	b.WriteString(dimStyle.Render(helpSetup))
	return b.String()
}

func (m model) playView() string {
	g := games[m.choice]
	return fmt.Sprintf("%s\n\n%d/%d  %s\n\n%s", g.game.view(m.settings), m.right, m.total, m.feedback, dimStyle.Render(g.help+" · esc back · q quit"))
}

func main() {
	m := model{settings: settings{enabled: [6]bool{true, true, true, true, true, true}, hi: 12, names: true, numbers: true}}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
