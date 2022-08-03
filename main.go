package main

import (
	"context"
	"fmt"
	"log"

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
)

var (
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("175")).Bold(true).Render
	docStyle  = lipgloss.NewStyle().Padding(3).Render
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	border    = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("228")).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).BorderBottom(true).Render
	color     = termenv.EnvColorProfile().Color
	ruler     = lipgloss.NewStyle().BorderBottom(true).BorderTop(true).BorderBackground(lipgloss.Color("228")).Foreground(lipgloss.Color("175")).Render
	greyText  = termenv.Style{}.Foreground(color("241")).Styled
	MainRuler = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), true, false).Render
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
	ti.Placeholder = "I'm at yor service, what can I do for you?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.Prompt = "🔍 "
	vp := viewport.New(78, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return model{
		uiState:   uiMainPage,
		textInput: ti,
		viewport:  vp,
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

				m.uiState = uiIsLoading
				m.spinner = spinner.New()
				m.spinner.Spinner = spinner.Pulse
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

				m.spinner, cmd = m.spinner.Update(msg)
				if m.textInput.Value() != "" {
					m.response = getCommand(m.textInput.Value())
					m.isReady = true
					m.uiState = uiLoaded

				} else {
					m.isReady = false
					m.uiState = uiMainPage

				}

				return m, cmd
			}
		}

		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case uiLoaded:
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "ctrl+n":
				m.uiState = uiMainPage
				m.textInput.Focus()
				m.textInput.SetValue("")

			case "esc":
				m.uiState = uiMainPage
				m.textInput.Blink()
				m.textInput.SetValue(m.textInput.Value())

			case "ctrl+q":
				cmd = tea.Quit
			}

		}

		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}
func (m model) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
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
		state = fmt.Sprintf(
			"\n "+textStyle("\n\n🔍 Searched:")+" "+m.textInput.Value()+"\n\n%s\n\n%s",

			MainRuler(ruler(m.response)),

			helpStyle("ctrl + n: new search modes • q: exit esc: back\n"),
		) + "\n"
	}
	return state
}

//_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
