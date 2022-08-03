package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	uiMainPage uiState = iota
	uiIsLoading
	uiLoaded
	useHighPerformanceRenderer = false
)

var (
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("175")).Bold(true).Render
	docStyle  = lipgloss.NewStyle().Padding(3).Render
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	MainRuler = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), true, false).Render
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type model struct {
	textInput textinput.Model
	uiState   uiState
	response  string
	err       error
	spinner   spinner.Model
	isReady   bool
	viewport  viewport.Model
}

type uiState int

func (m model) Init() tea.Cmd {
	switch m.uiState {
	case uiMainPage:
		return textinput.Blink
	case uiIsLoading:
		return nil

	case uiLoaded:
		return nil

	}
	return nil
}

func initialModel() model {

	ti := textinput.New()
	ti.Placeholder = "Let me convert your English into code?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.Prompt = "🔍 "

	return model{
		uiState:   uiMainPage,
		textInput: ti,
		err:       nil,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch m.uiState {
	case uiMainPage:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			headerHeight := lipgloss.Height(m.headerView())
			footerHeight := lipgloss.Height(m.footerView())
			verticalMarginHeight := headerHeight + footerHeight

			if !m.isReady {
				m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
				m.viewport.YPosition = headerHeight
				m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
				// m.viewport.SetContent(textStyle(m.response))
				// m.isReady = true
				m.viewport.YPosition = headerHeight + 1
				//panic(msg.Height - verticalMarginHeight)
			} else {
				m.viewport.Width = msg.Width
				m.viewport.Height = msg.Height - verticalMarginHeight
			}

			// if useHighPerformanceRenderer {
			// 	cmds = append(cmds, viewport.Sync(m.viewport))
			// }
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				return m, tea.Quit

			case "enter":
				m.response = getCommand(m.textInput.Value())
				m.uiState = uiLoaded
				return m, cmd
			}

		}

		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case uiLoaded:

		if !m.isReady {
			m.viewport.SetContent(textStyle(m.response))
			m.isReady = true
		} else {
			m.viewport.SetContent(" ")
		}

		if useHighPerformanceRenderer {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "ctrl+n":
				m.isReady = false
				m.response = ""
				m.uiState = uiMainPage
				m.textInput.Focus()
				m.textInput.SetValue("")

			case "esc":
				m.isReady = false
				m.response = ""
				m.uiState = uiMainPage
				m.textInput.Blink()
				m.textInput.SetValue(m.textInput.Value())

			}

		}

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) helpView(help string) string {
	return helpStyle("\n  " + help + " \n")
}

func (m model) headerView() string {
	title := titleStyle.Render(m.textInput.Value())
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) View() string {
	var state string

	switch m.uiState {
	case uiMainPage:

		state =
			docStyle(fmt.Sprintf(
				textStyle("Commander")+"\n\n%s\n\n\n%s",
				m.textInput.View(),
				m.helpView("enter: confirm exit • esc: exit\n"),
			) + "\n")
	case uiIsLoading:
		state = fmt.Sprintf("\n %s%s%s\n\n", m.spinner.View(), " ", textStyle("Thinking..."))
	case uiLoaded:

		state = fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	}
	return state
}

//_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
