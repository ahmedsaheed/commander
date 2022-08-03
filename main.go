package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
	// border    = lipgloss.NewStyle().
	// 		BorderStyle(lipgloss.RoundedBorder()).
	// 		BorderForeground(lipgloss.Color("228")).
	// 		BorderTop(true).
	// 		BorderLeft(true).
	// 		BorderRight(true).BorderBottom(true).Render
	color     = termenv.EnvColorProfile().Color
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
	quitting  bool
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

func getCommand(word string) string {

	apiKey := "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"
	if apiKey == "" {
		log.Fatalln("Missing API KEY")
	}

	ctx := context.Background()
	client := gpt3.NewClient(apiKey)

	resp, err := client.CompletionWithEngine(ctx, "text-davinci-002", gpt3.CompletionRequest{
		Prompt:           []string{word},
		MaxTokens:        gpt3.IntPtr(450),
		Temperature:      gpt3.Float32Ptr(0),
		FrequencyPenalty: float32(0.2),
		PresencePenalty:  float32(0),
		TopP:             gpt3.Float32Ptr(1),
	})
	if err != nil {
		fmt.Println("Alas, something went wrong.")
		log.Fatalln(err)
		return "Error" + err.Error()
	}
	s := ""
	for i := 0; i < len(resp.Choices); i++ {
		s = s + resp.Choices[i].Text + "\n"
	}
	return s

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.uiState {
	case uiMainPage:
		var cmd tea.Cmd
		switch msg := msg.(type) {
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
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)

		switch msg := msg.(type) {

		case tea.WindowSizeMsg:
			headerHeight := lipgloss.Height(m.headerView())
			footerHeight := lipgloss.Height(m.footerView())
			verticalMarginHeight := headerHeight + footerHeight

			if !m.isReady {
				m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
				m.viewport.YPosition = headerHeight
				m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
				m.viewport.SetContent(textStyle(m.response))
				m.isReady = true
				m.viewport.YPosition = headerHeight + 1
			} else {
				m.viewport.Width = msg.Width
				m.viewport.Height = msg.Height - verticalMarginHeight
			}

			if useHighPerformanceRenderer {
				cmds = append(cmds, viewport.Sync(m.viewport))
			}
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
func (m model) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
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
				helpStyle("enter: confirm exit • esc: exit\n"),
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
