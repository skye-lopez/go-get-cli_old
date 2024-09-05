// Bespoke interaction class for the search feature since its a bit more complex
// NOTE: It may be worthwile to refactor the initial interaction class to support this if thats simple enough

package interaction

import (
	"fmt"

	"github.com/buger/goterm"
	"github.com/inancgumus/screen"
)

type SearchInteraction struct {
	Prompts        map[int]*Prompt // Pointer to prompts to render upon selection a package option.
	Trie           map[string][]*Option
	SearchInput    string // Text to filter
	CurrentIdx     int
	CursorIdx      int
	SearchSelected bool
	NextInsertIdx  int
}

func NewSearchInteraction() *SearchInteraction {
	return &SearchInteraction{
		SearchInput:    "",
		SearchSelected: true,
		Prompts:        make(map[int]*Prompt, 0),
		CursorIdx:      0,
		NextInsertIdx:  0,
		CurrentIdx:     0,
		Trie:           make(map[string][]*Option, 0),
	}
}

func (s *SearchInteraction) CreatePrompt(title string, description string, isPaginated bool) *Prompt {
	p := &Prompt{
		Title:       title,
		Description: description,
		Options:     make([][]*Option, 0),
		Idx:         s.NextInsertIdx,
		PageIdx:     0,
		IsPaginated: isPaginated,
	}

	firstPage := []*Option{}
	p.Options = append(p.Options, firstPage)
	s.Prompts[p.Idx] = p
	s.NextInsertIdx += 1

	return p
}

// On New insert of option build out Trie
func (s *SearchInteraction) UpdateTrie(o *Option) {
	insertTerm := ""
	for _, char := range o.Title {
		c := string(char)
		insertTerm += c

		_, exists := s.Trie[insertTerm]
		if !exists {
			s.Trie[insertTerm] = make([]*Option, 0)
		} else {
			// check if we already appended this - maybe a map would be better here since this is O(n) every time...
			for _, v := range s.Trie[insertTerm] {
				if v == o {
					return
				}
			}
			s.Trie[insertTerm] = append(s.Trie[insertTerm], o)
		}

	}
}

func (s *SearchInteraction) Open() {
	defer func() {
		fmt.Printf("\033[?25h")
	}()
	fmt.Printf("\033[?25l")

	s.Render()
	for {
		key := userInput()
		p := s.getCurrentPrompt()
		pLen := len(p.Options[p.PageIdx])
		switch s.SearchSelected {
		case true:
			switch key {
			case escape:
				return
			case r:
				s.SearchSelected = false
				s.Render()
			}
			fmt.Println("\n\n\n\n\n\n", key)
		case false:
			switch key {
			case escape:
				return
			case search:
				s.SearchSelected = true
				s.Render()
			case n:
				if p.PageIdx+1 < len(p.Options) {
					p.PageIdx += 1
					s.CursorIdx = 0
					s.Render()
				}
			case b:
				if p.PageIdx-1 >= 0 {
					p.PageIdx -= 1
					s.CursorIdx = 0
					s.Render()
				}
			case up:
				if s.CursorIdx-1 >= 0 {
					s.CursorIdx -= 1
					s.Render()
				}
			case down:
				if s.CursorIdx+1 < pLen {
					s.CursorIdx += 1
					s.Render()
				}
			case enter:
				selectedOption := p.Options[p.PageIdx][s.CurrentIdx]

				// If the option has children render that
				if selectedOption.PromptIdx > 0 {
					s.RenderNewPrompt(selectedOption.PromptIdx)
				}

				// Otherwise handle the option
				if selectedOption.Callback != nil {
					message, err := selectedOption.Callback()
					fmt.Println("\n\n\n", message, "\n\n\n", err)
				}
			case u: // naviagte up
				if p.ParentIdx >= 0 {
					s.RenderNewPrompt(p.ParentIdx)
				}
			}
		}
	}
}

func (s *SearchInteraction) RenderNewPrompt(newIdx int) {
	s.CurrentIdx = newIdx
	s.CursorIdx = 0
	s.Prompts[s.CurrentIdx].PageIdx = 0
	s.Render()
}

func (s *SearchInteraction) Render() {
	s.Clear()
	p := s.getCurrentPrompt()

	// Render Title
	fmt.Printf("\r%s\n%s\n", goterm.Color(goterm.Bold(p.Title), goterm.CYAN), goterm.Color(p.Description, goterm.MAGENTA))

	// If we are on the base prompt 0 we render the search bar
	if s.CurrentIdx == 0 {
		var searchDisplay string
		if s.SearchSelected {
			searchDisplay = goterm.Color(goterm.Bold("> Search: "+s.SearchInput), goterm.CYAN)
		} else {
			searchDisplay = goterm.Bold("Search >>" + s.SearchInput)
		}
		fmt.Println("-----------------------------------------------------------------------------------------------")
		fmt.Printf("\r %s \n", searchDisplay)
		fmt.Println("-----------------------------------------------------------------------------------------------")
	}

	optionsToRender := p.Options[p.PageIdx]
	// Because we are repainting, we need to ensure it gets cleared fully.
	linePadding := "                                                                               "
	nl := "\n"

	for j, v := range optionsToRender {
		if j == len(optionsToRender)-1 {
			nl = ""
		}
		switch j == s.CursorIdx && !s.SearchSelected {
		case true:
			fmt.Printf("\r%s%s%s%s",
				goterm.Color(goterm.Bold(">  "), goterm.YELLOW),
				goterm.Color(goterm.Bold(v.Title), goterm.YELLOW),
				goterm.Bold(" ("+v.Description+") "+linePadding),
				nl)
		case false:
			fmt.Printf("\r%s%s%s%s", "  ", v.Title, " ("+v.Description+") "+linePadding, "\n")
		}
	}
}

func (s *SearchInteraction) getCurrentPrompt() *Prompt {
	return s.Prompts[s.CurrentIdx]
}

func (s *SearchInteraction) Clear() {
	screen.Clear()
	screen.MoveTopLeft()
}
