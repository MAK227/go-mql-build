package Common

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type FilePicker struct {
	renderer  *glamour.TermRenderer
	treeState string
	Mode      string
	Files     []File
	CurrIndex int
	width     int
	height    int
}

const THEME = "dracula"

func (m *FilePicker) ReadFiles(force bool) {
	start := m.CurrIndex - 1
	if start < 0 {
		start = 0
	}
	end := m.CurrIndex + 1
	if end > len(m.Files)-1 {
		end = len(m.Files)
	}

	for i := start; i < end; i++ {
		// read file if it hasn't been read yet
		if m.Files[i].Content == "" || force {
			dat, _ := os.ReadFile(m.Files[i].Path)
			content := string(dat)
			if content == "" {
				content = "#Empty file"
			} else {
				content = fmt.Sprintf("```cpp\n%s\n```", content)
				headLines := strings.Split(content, "\n")
				content = strings.Join(headLines[:min(m.height-2, len(headLines))], "\n")
			}

			m.Files[i].Content, _ = m.renderer.Render(content)

		}
	}
}

func (m FilePicker) Init() (tea.Model, tea.Cmd) {
	files := getFiles(".")
	m.Files = files
	m.CurrIndex = 0
	m.width, m.height, _ = term.GetSize(0)
	m.Rerender(true)
	return m, nil
}

func (m *FilePicker) Rerender(force bool) {
	if force {
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithStandardStyle(THEME),
			glamour.WithWordWrap(m.width-lipgloss.Width(m.treeState)),
		)
	}
	m.treeState = lipgloss.NewStyle().Height(m.height).Render(m.buildTree(true))
	m.ReadFiles(force)
}

func (m FilePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "down":
			// select the next file
			for i := range m.Files {
				if m.Files[i].Selected && i < len(m.Files)-1 {
					m.Files[i].Selected = false
					m.Files[i+1].Selected = true
					m.CurrIndex = i + 1
					m.Rerender(false)
					break
				}
			}
		case "up":
			// select the previous file
			for i := range m.Files {
				if m.Files[i].Selected && i > 0 {
					m.Files[i].Selected = false
					m.Files[i-1].Selected = true
					m.CurrIndex = i - 1
					m.Rerender(false)
					break
				}
			}
		case "enter", "c":
			m.Mode = "compile"
			return m, tea.Quit
		case "s":
			m.Mode = "syntax"
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 3
		m.Rerender(true)
	}

	return m, nil
}

func (m FilePicker) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		helpView(),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.treeState,
			m.Files[m.CurrIndex].Content,
		),
	)
}
