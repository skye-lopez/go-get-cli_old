// Bespoke interaction class for the search feature since its a bit more complex
// NOTE: It may be worthwile to refactor the initial interaction class to support this if thats simple enough

package interaction

import (
	"fmt"

	"github.com/buger/goterm"
	"github.com/inancgumus/screen"
)

type SearchInteraction struct {
	Prompts        map[int]*Prompt     // Pointer to prompts to render upon selection a package option.
	StoredOptions  map[int][][]*Option // This is a temp reference to the non-searched options after they have been paginated.
	Trie           map[string][][]*Option
	SearchInput    string // Text to filter
	CurrentIdx     int
	CursorIdx      int
	NextInsertIdx  int
	SearchSelected bool
}

func NewSearchInteraction() *SearchInteraction {
	return &SearchInteraction{
		SearchInput:    "",
		SearchSelected: true,
		Prompts:        make(map[int]*Prompt, 0),
		CursorIdx:      0,
		NextInsertIdx:  0,
		CurrentIdx:     0,
		Trie:           make(map[string][][]*Option, 0),
		StoredOptions:  make(map[int][][]*Option),
	}
}

func (s *SearchInteraction) StoreOptionsFromPrompt(p *Prompt) {
	s.StoredOptions[p.Idx] = p.Options
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

func (s *SearchInteraction) UpdateTrie(o *Option) {
	terms := make([]string, 0)
	runningTerm := ""
	for _, char := range o.Title {
		c := string(char)
		runningTerm += c
		terms = append(terms, runningTerm)
	}

	for _, term := range terms {
		_, exists := s.Trie[term]

		// Init new section if needed
		if !exists {
			newState := [][]*Option{{o}}
			s.Trie[term] = newState
			continue
		}

		// Check if this option is already stored in the term (no duplicates)
		skip := false
		for _, page := range s.Trie[term] {
			for _, v := range page {
				// if we find the current value we dont want to do anything.
				if v == o {
					skip = true
				}
			}
		}

		if skip {
			continue
		}

		lastPageIdx := len(s.Trie[term]) - 1
		if len(s.Trie[term][lastPageIdx]) < 10 {
			s.Trie[term][lastPageIdx] = append(s.Trie[term][lastPageIdx], o)
		} else {
			newPage := []*Option{o}
			s.Trie[term] = append(s.Trie[term], newPage)
		}
	}
}

func (s *SearchInteraction) UpdateOnSearch() {
	if s.SearchInput == "" {
		s.Prompts[0].Options = s.StoredOptions[0]
		return
	}

	_, avail := s.Trie[s.SearchInput]
	if !avail {
		emptyOption := &Option{
			Title:       "No Search results!",
			Description: "Try another search term",
		}
		s.Prompts[0].Options = [][]*Option{{emptyOption}}
		return
	}
	s.Prompts[0].Options = s.Trie[s.SearchInput]
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
			case results:
				s.SearchSelected = false
				s.Render()
			case backspace:
				if len(s.SearchInput) > 0 {
					s.SearchInput = s.SearchInput[:len(s.SearchInput)-1]
				}
				s.UpdateOnSearch()
				s.Render()
			default:
				str := string(key)
				if str != " " {
					s.SearchInput += str
					s.UpdateOnSearch()
					s.Render()
				}
			}
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
				selectedOption := p.Options[p.PageIdx][s.CursorIdx]

				// If the option has children render that
				if selectedOption.PromptIdx > 0 {
					s.RenderNewPrompt(selectedOption.PromptIdx)
				}

				// Otherwise handle the option
				if selectedOption.Callback != nil {
					fmt.Println("Index we are looking for:")
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
	if s.CurrentIdx == 0 {
		var keyOptions string
		if s.SearchSelected {
			keyOptions = "[=] Select Results | [esc] Exit (Spaces are excluded)"
		} else {
			keyOptions = "[+] Select Search | [n] Next Page | [b] Last Page | [enter] Select Package | [esc] Exit"
		}
		fmt.Printf("\r%s\n%s\n",
			goterm.Color(goterm.Bold("Search for a package!"), goterm.CYAN),
			goterm.Color(keyOptions, goterm.MAGENTA))
	} else {
		fmt.Printf("\r%s\n%s\n", goterm.Color(goterm.Bold(p.Title), goterm.CYAN), goterm.Color(p.Description, goterm.MAGENTA))
	}

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
