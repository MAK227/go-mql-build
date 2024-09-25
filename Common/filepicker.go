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
			}

			content, _ = m.renderer.Render(content)

			headlines := strings.Split(content, "\n")
			content = strings.Join(headlines[:min(m.height-2, len(headlines))], "\n")

			m.Files[i].Content = content

		}
	}
}

func (m FilePicker) Init() (tea.Model, tea.Cmd) {
	files, err := getFiles(".")
	if err != nil {
		return m, tea.Quit
	}
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
	m.treeState = m.buildTree(true)
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

		case "ctrl+d":
			// select the next file
			for i := range m.Files {
				if m.Files[i].Selected && i < len(m.Files)-1 {

					fileIndex := min(i+5, len(m.Files)-1)

					m.Files[i].Selected = false
					m.Files[fileIndex].Selected = true
					m.CurrIndex = fileIndex
					m.Rerender(false)
					break
				}
			}
		case "ctrl+u":
			// select the previous file
			for i := range m.Files {
				if m.Files[i].Selected && i > 0 {

					fileIndex := max(i-5, 0)

					m.Files[i].Selected = false
					m.Files[fileIndex].Selected = true
					m.CurrIndex = fileIndex
					m.Rerender(false)
					break
				}
			}

		case "shift+d":
			// select the next file
			for i := range m.Files {
				if m.Files[i].Selected && i < len(m.Files)-1 {
					m.Files[i].Selected = false
					m.Files[len(m.Files)-1].Selected = true
					m.CurrIndex = len(m.Files) - 1
					m.Rerender(false)
					break
				}
			}
		case "shift+u":
			// select the previous file
			for i := range m.Files {
				if m.Files[i].Selected && i > 0 {
					m.Files[i].Selected = false
					m.Files[0].Selected = true
					m.CurrIndex = 0
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
		m.height = msg.Height - 1
		m.Rerender(true)
	}

	return m, nil
}

func (m FilePicker) View() string {
	var content string

	if len(m.Files) == 0 {
		content = "No files found"
	} else {
		content = m.Files[m.CurrIndex].Content
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		helpView(),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.treeState,
			content,
		),
	)
}
