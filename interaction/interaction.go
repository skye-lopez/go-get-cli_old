package interaction

import (
	"fmt"

	"github.com/buger/goterm"
	"github.com/inancgumus/screen"
	"github.com/pkg/term"
)

type Interaction struct {
	Prompts           map[int]*Prompt
	CurrentIdx        int
	NextInsertIdx     int
	CursorIdx         int
	LinesOnLastRender int
}

type Prompt struct {
	Description string
	Title       string
	Options     [][]*Option
	Idx         int
	PageIdx     int
	IsPaginated bool
}

type Option struct {
	Packet      any
	Title       string
	Description string
	PromptIdx   int
	Callback    func(...any)
}

func NewInteraction() *Interaction {
	return &Interaction{
		Prompts:           make(map[int]*Prompt, 0),
		CurrentIdx:        0,
		NextInsertIdx:     0,
		CursorIdx:         0,
		LinesOnLastRender: 0,
	}
}

func (i *Interaction) CreatePrompt(title string, description string, isPaginated bool) *Prompt {
	i.NextInsertIdx += 1
	p := &Prompt{
		Title:       title,
		Description: description,
		Options:     make([][]*Option, 0),
		Idx:         i.NextInsertIdx,
		PageIdx:     0,
		IsPaginated: isPaginated,
	}

	firstPage := []*Option{}
	p.Options = append(p.Options, firstPage)

	// check if we are making the first prompt
	_, ok := i.Prompts[0]
	if !ok {
		i.Prompts[0] = p
	} else {
		i.Prompts[i.NextInsertIdx] = p
	}
	return p
}

func (p *Prompt) AddOption(title string, description string, packet any) *Option {
	o := &Option{
		Title:       title,
		Description: description,
		Packet:      packet,
	}

	// Skip paginating
	if !p.IsPaginated {
		p.Options[0] = append(p.Options[0], o)
		return o
	}

	// Get idx of the last page
	lastPage := len(p.Options) - 1

	// Fill to 10
	if len(p.Options[lastPage]) < 10 {
		p.Options[lastPage] = append(p.Options[lastPage], o)
		return o
	}

	// Create new page if already at 10
	newPage := []*Option{o}
	p.Options = append(p.Options, newPage)

	return o
}

func (o *Option) AttachPrompt(promptIdx int) {
	o.PromptIdx = promptIdx
}

func (o *Option) AddCallback(cb func(...any)) {
	o.Callback = cb
}

func (i *Interaction) Open() *Option {
	// Hide cursor and return it on close
	defer func() {
		fmt.Printf("\033[?25h")
	}()
	fmt.Printf("\033[?25l")

	// Render initial state
	i.Render()
	p := i.Prompts[i.CurrentIdx]
	pLen := len(p.Options[p.PageIdx])
	for {
		key := userInput()

		switch key {
		case escape:
			return &Option{}
		case n:
			if p.PageIdx+1 < len(p.Options) {
				p.PageIdx += 1
				i.CursorIdx = 0
				i.Render()
			}
		case b:
			if p.PageIdx-1 >= 0 {
				p.PageIdx -= 1
				i.CursorIdx = 0
				i.Render()
			}
		case up:
			if i.CursorIdx-1 >= 0 {
				i.CursorIdx -= 1
				i.Render()
			}
		case down:
			if i.CursorIdx+1 < pLen {
				i.CursorIdx += 1
				i.Render()
			}
		case enter:
			cprompt := i.Prompts[i.CurrentIdx]
			cpromptOptions := cprompt.Options[cprompt.PageIdx]
			selectedOption := cpromptOptions[i.CursorIdx]

			// If the option has children render that
			if selectedOption.PromptIdx > 0 {
				i.CurrentIdx = selectedOption.PromptIdx
				i.CursorIdx = 0
				i.Prompts[i.CurrentIdx].PageIdx = 0
				i.Render()
			}

			// Otherwise handle the option
			if selectedOption.Callback != nil {
				selectedOption.Callback()
			}
		}
	}
}

// This makes our UI really easy to work with but im not a huge fan of clearing any past context...
func (i *Interaction) HardFlushScreen() {
	screen.Clear()
	screen.MoveTopLeft()
}

// Render is called on any user input action and by default repaints the current Prompt
func (i *Interaction) Render() {
	// For now we are just going to hard clear. can come back to this later.
	i.HardFlushScreen()
	p := i.Prompts[i.CurrentIdx]
	// Redraw
	// This *MOSTLY* works but Im sure will need to be changed a bit.
	if i.LinesOnLastRender > 1 {
		fmt.Printf("\033[%dA", i.LinesOnLastRender)
	}
	i.LinesOnLastRender = 0

	// Draw prompts title
	fmt.Printf("\r%s\n%s\n", goterm.Color(goterm.Bold(p.Title), goterm.CYAN), goterm.Color(p.Description, goterm.MAGENTA))
	i.LinesOnLastRender += 2

	optionsToRender := p.Options[p.PageIdx]
	// Because we are repainting, we need to ensure it gets cleared fully.
	linePadding := "                                                                               "
	nl := "\n"

	for j, v := range optionsToRender {
		if j == len(optionsToRender)-1 {
			nl = ""
		}
		switch j == i.CursorIdx {
		case true:
			fmt.Printf("\r%s%s%s%s",
				goterm.Color(goterm.Bold(">  "), goterm.YELLOW),
				goterm.Color(goterm.Bold(v.Title), goterm.YELLOW),
				goterm.Bold(" ("+v.Description+") "+linePadding),
				nl)
		case false:
			fmt.Printf("\r%s%s%s%s", "  ", v.Title, v.Description+linePadding, "\n")
		}
		i.LinesOnLastRender += 1
	}
}

// Raw input keycodes
var (
	up     byte = 65
	down   byte = 66
	escape byte = 27
	enter  byte = 13
	n      byte = 110
	b      byte = 98
)

func userInput() byte {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)

	// Read in 3 bytes at a time
	var action int
	bytes := make([]byte, 3)
	action, _ = t.Read(bytes)

	t.Restore()
	t.Close()

	// Arrow keys have a <esc>[ prefix so the 3rd byte is actually what we want
	// Otherwise we just want the initial byte
	if action == 3 {
		return bytes[2]
	} else {
		return bytes[0]
	}
}
