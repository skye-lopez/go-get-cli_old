package cmd

import (
	"fmt"
	"log"

	"github.com/buger/goterm"
	"github.com/pkg/term"
)

// Raw input keycodes
var (
	up     byte = 65
	down   byte = 66
	escape byte = 27
	enter  byte = 13
	next   byte = 110
	back   byte = 98
	keys        = map[byte]bool{
		up:   true,
		down: true,
	}
)

type Menu struct {
	Prompt       string
	CursorPos    int
	MenuItems    []*MenuItem
	AllMenuItems []*MenuItem
	StartIdx     int
	EndIdx       int
	Paginate     bool
}

type MenuItem struct {
	Text string
	ID   string
	Data any
}

func NewMenu(prompt string, paginate bool) *Menu {
	return &Menu{
		Prompt:       prompt,
		MenuItems:    make([]*MenuItem, 0),
		AllMenuItems: make([]*MenuItem, 0),
		StartIdx:     0,
		EndIdx:       10,
		Paginate:     paginate,
	}
}

// AddItem will add a new menu option to the menu list
func (m *Menu) AddItem(option string, id string, data any) *Menu {
	menuItem := &MenuItem{
		Text: option,
		ID:   id,
		Data: data,
	}

	if m.Paginate {
		if len(m.MenuItems) < 10 {
			m.MenuItems = append(m.MenuItems, menuItem)
		}
		m.AllMenuItems = append(m.AllMenuItems, menuItem)
	} else {
		m.MenuItems = append(m.MenuItems, menuItem)
	}
	return m
}

// renderMenuItems prints the menu item list.
// Setting redraw to true will re-render the options list with updated current selection.
func (m *Menu) renderMenuItems(redraw bool) {
	if redraw {
		// Move the cursor up n lines where n is the number of options, setting the new
		// location to start printing from, effectively redrawing the option list
		//
		// This is done by sending a VT100 escape code to the terminal
		// @see http://www.climagic.org/mirrors/VT100_Escape_Codes.html
		fmt.Printf("\033[%dA", (len(m.MenuItems))-1)
	}

	for index, menuItem := range m.MenuItems {
		newline := "\n"
		if index == len(m.MenuItems)-1 {
			// Adding a new line on the last option will move the cursor position out of range
			// For out redrawing
			newline = ""
		}

		menuItemText := menuItem.Text
		cursor := "  "
		if index == m.CursorPos {
			cursor = goterm.Color("> ", goterm.YELLOW)
			menuItemText = goterm.Color(menuItemText, goterm.YELLOW)
		}

		fmt.Printf("\r %s %s%s", cursor, menuItemText, newline)
	}
}

// Display will display the current menu options and awaits user selection
// It returns the users selected choice
func (m *Menu) Display() *MenuItem {
	defer func() {
		// Show cursor again.
		fmt.Printf("\033[?25h")
	}()

	fmt.Printf("%s\n", goterm.Color(goterm.Bold(m.Prompt)+":", goterm.CYAN))

	m.renderMenuItems(false)

	// Turn the terminal cursor off
	fmt.Printf("\033[?25l")

	for {
		keyCode := getInput()
		if keyCode == escape {
			return &MenuItem{}
		} else if keyCode == enter {
			menuItem := m.MenuItems[m.CursorPos]
			fmt.Println("\r")
			return menuItem
		} else if keyCode == up {
			m.CursorPos = (m.CursorPos + len(m.MenuItems) - 1) % len(m.MenuItems)
			m.renderMenuItems(true)
		} else if keyCode == down {
			m.CursorPos = (m.CursorPos + 1) % len(m.MenuItems)
			m.renderMenuItems(true)
		} else if keyCode == next {
			if m.Paginate {

				start := m.StartIdx
				end := m.EndIdx
				step := 10

				for end < len(m.AllMenuItems) && step > 0 {
					end += 1
					start += 1
					step -= 1
				}

				m.StartIdx = start
				m.EndIdx = end

				m.MenuItems = m.AllMenuItems[m.StartIdx:m.EndIdx]
				m.CursorPos = 0
				m.renderMenuItems(true)
			}
		} else if keyCode == back {
			if m.Paginate {

				start := m.StartIdx
				end := m.EndIdx
				step := 10

				for start > 0 && step > 0 {
					end -= 1
					start -= 1
					step -= 1
				}

				m.StartIdx = start
				m.EndIdx = end
				m.MenuItems = m.AllMenuItems[m.StartIdx:m.EndIdx]
				m.CursorPos = 0
				m.renderMenuItems(true)
			}
		}
	}
}

// getInput will read raw input from the terminal
// It returns the raw ASCII value inputted
func getInput() byte {
	t, _ := term.Open("/dev/tty")

	err := term.RawMode(t)
	if err != nil {
		log.Fatal(err)
	}

	var read int
	readBytes := make([]byte, 3)
	read, err = t.Read(readBytes)

	t.Restore()
	t.Close()

	// Arrow keys are prefixed with the ANSI escape code which take up the first two bytes.
	// The third byte is the key specific value we are looking for.
	// For example the left arrow key is '<esc>[A' while the right is '<esc>[C'
	// See: https://en.wikipedia.org/wiki/ANSI_escape_code
	if read == 3 {
		if _, ok := keys[readBytes[2]]; ok {
			return readBytes[2]
		}
	} else {
		return readBytes[0]
	}

	return 0
}
