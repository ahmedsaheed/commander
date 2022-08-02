package main

import (
	"context"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"log"
	"math"
	"strconv"
	"strings"
)

type uiState int

const (
	uiMainPage uiState = iota
	uiIsLoading
	uiLoaded
	progressBarWidth  = 71
	progressFullChar  = "█"
	progressEmptyChar = "░"
)

var (
	term          = termenv.EnvColorProfile()
	keyword       = makeFgStyle("211")
	subtle        = makeFgStyle("241")
	progressEmpty = subtle(progressEmptyChar)
	dot           = colorFg(" • ", "236")
	textStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render

	// Gradient colors we'll use for the progress bar
	ramp = makeRamp("#B14FFF", "#00FFA3", progressBarWidth)
)

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
func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

func makeRamp(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
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
func progressbar(width int, percent float64) string {
	w := float64(progressBarWidth)

	fullSize := int(math.Round(w * percent))
	var fullCells string
	for i := 0; i < fullSize; i++ {
		fullCells += termenv.String(progressFullChar).Foreground(term.Color(ramp[i])).String()
	}

	emptySize := int(w) - fullSize
	emptyCells := strings.Repeat(progressEmpty, emptySize)

	return fmt.Sprintf("%s%s %3.0f", fullCells, emptyCells, math.Round(percent*100))
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
	return resp.Choices[0].Text

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
				if m.uiState == uiIsLoading && msg.String() == "esc" {
					return m, tea.Quit
				}

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
		return docStyle.Render(fmt.Sprintf("\n %s%s%s\n\n", m.spinner.View(), " ", textStyle("Spinning...")))
	case uiLoaded:
		return docStyle.Render(m.response)
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
