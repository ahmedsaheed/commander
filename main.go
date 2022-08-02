package main

import (
	"context"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
)

const (
	uiMainPage uiState = iota
	uiIsLoading
	uiLoaded
)

var textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	textInput textinput.Model
	uiState   uiState
	response  string
	err       error
	spinner   spinner.Model
	mounted   bool
	isReady   bool
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
	ti.Prompt = "🐙 "

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
		MaxTokens:        gpt3.IntPtr(300),
		Temperature:      gpt3.Float32Ptr(0),
		FrequencyPenalty: float32(0.2),
		PresencePenalty:  float32(0),
		TopP:             gpt3.Float32Ptr(1),
	})
	if err != nil {
		fmt.Println("Hmm, something ins't right.")
		log.Fatalln(err)
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
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit

			case "enter":
				m.uiState = uiIsLoading
				m.spinner = spinner.New()
				m.spinner.Spinner = spinner.Pulse
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

				var cmd tea.Cmd
				m.spinner, cmd = m.spinner.Update(msg)
				if m.textInput.Value() != "" {
					m.response = getCommand(m.textInput.Value())
					m.isReady = true
					m.uiState = uiLoaded

				}

				return m, cmd
			}
		}

		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case uiLoaded:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			}
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {

	switch m.uiState {
	case uiMainPage:
		return fmt.Sprintf(
			"Search: \n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	case uiIsLoading:
		return fmt.Sprintf("\n %s%s%s\n\n", m.spinner.View(), " ", textStyle("Thinking..."))
	case uiLoaded:
		return fmt.Sprintf(
			"\n 🔎 Searched: "+m.textInput.Value()+"\n\n%s\n\n%s",

			m.response,

			"Esc to quit",
		) + "\n"

	default:
		return ""
	}
}

//_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_

func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
