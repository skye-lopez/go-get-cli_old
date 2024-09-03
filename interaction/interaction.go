package interaction

import (
	"fmt"

	"github.com/buger/goterm"
	"github.com/pkg/term"
)

type Interaction struct {
	Prompts           map[int]*Prompt
	CurrentIdx        int
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
}

func NewInteraction() *Interaction {
	return &Interaction{
		Prompts:           make(map[int]*Prompt, 0),
		CurrentIdx:        0,
		CursorIdx:         0,
		LinesOnLastRender: 0,
	}
}

func (i *Interaction) CreatePrompt(title string, description string, isPaginated bool) *Prompt {
	i.CurrentIdx += 1
	p := &Prompt{
		Title:       title,
		Description: description,
		Options:     make([][]*Option, 0),
		Idx:         i.CurrentIdx,
		PageIdx:     0,
		IsPaginated: isPaginated,
	}

	firstPage := []*Option{}
	p.Options = append(p.Options, firstPage)

	i.Prompts[i.CurrentIdx] = p
	return p
}

func (p *Prompt) AddOption(title string, description string, packet any) {
	o := &Option{
		Title:       title,
		Description: description,
		Packet:      packet,
	}

	// Get idx of the last page
	lastPage := len(p.Options) - 1

	// Fill to 10
	if len(p.Options[lastPage]) < 10 {
		p.Options[lastPage] = append(p.Options[lastPage], o)
		return
	}

	// Create new page if already at 10
	newPage := []*Option{o}
	p.Options = append(p.Options, newPage)
}

func (i *Interaction) Open() {
	// Render initial state
	i.Render()
	p := i.Prompts[i.CurrentIdx]
	for {
		key := userInput()

		switch key {
		case escape:
			return
		case n:
			if p.PageIdx+1 < len(p.Options) {
				p.PageIdx += 1
				i.Render()
			}
		case b:
			if p.PageIdx-1 >= 0 {
				p.PageIdx -= 1
				i.Render()
			}
		}
	}
}

// Render is called on any user input action and by default repaints the current Prompt
func (i *Interaction) Render() {
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

	// Draw prompts options
	optionsToRender := p.Options[p.PageIdx]
	// Because we are repainting, we need to ensure it gets cleared fully.
	linePadding := "                                            "

	iters := 0
	for _, v := range optionsToRender {
		fmt.Printf("\r%s%s", v.Title+linePadding, "\n")
		i.LinesOnLastRender += 1
		iters += 1
	}

	// For paginated prompts we want to ensure they are drawn properly.
	if p.IsPaginated {
		for iters < 10 {
			fmt.Printf("\r%s%s", linePadding+linePadding, "\n")
			iters += 1
			i.LinesOnLastRender += 1
		}
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
