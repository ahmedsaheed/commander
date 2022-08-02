package main

import (
	"context"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

type response string
type errMsg struct{ err error }
type model struct {
	textInput textinput.Model
	progress  progress.Model
	response  string
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "I'm at yor service, what can I do for you?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.Prompt = "🐙 "

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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
		MaxTokens:        gpt3.IntPtr(100),
		Temperature:      gpt3.Float32Ptr(0),
		FrequencyPenalty: float32(0.2),
		PresencePenalty:  float32(0),
		TopP:             gpt3.Float32Ptr(1),
	})
	if err != nil {
		fmt.Println("Hmm, something ins't right.")
		log.Fatalln(err)
	}
	//fmt.Println(resp)
	return resp.Choices[0].Text

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.textInput.Value() != "" {
				m.response = getCommand(m.textInput.Value())
				println(getCommand(m.textInput.Value()))
				println()
				return m, tea.Quit
			} else {
				cmd = textinput.Blink
			}

		}

		//case response:
		//	m.response = string(msg)
		//	return m, tea.Quit
		//
		//// We handle errors just like any other message
		//case errMsg:
		//	m.err = msg.err
		//	return m, tea.Quit
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {

	if m.err != nil {
		return fmt.Sprintf("Error: %s", m.err)
	}
	return fmt.Sprintf(
		"Commander\n\n%s\n\n%s",
		m.textInput.View(),
		"enter confirm • esc quit",
	) + "\n"
}

////_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_
//
func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
