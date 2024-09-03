// Simple data manager using JSON files.
// This is not ideal but ensures a user does not need sqlite or such to use this CLI tool.
// if this sucks or performance does I will change it.

package store

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func Init() {
	FetchAndParseMD()
}

type Entry struct {
	Category    string
	Name        string
	Link        string
	Description string
}

type Category struct {
	Name    string
	Entries []Entry
}

type Store struct {
	Entries    []Entry
	Categories []Category
}

func FetchAndParseMD() {
	store := &Store{
		Entries:    []Entry{},
		Categories: []Category{},
	}

	resp, err := http.Get("https://raw.githubusercontent.com/avelino/awesome-go/main/README.md")
	if err != nil {
		fmt.Println("Error getting readme")
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading readme body")
		panic(err)
	}

	b := string(body)

	// Step 1 - parse each section
	sections := strings.Split(b, "**[⬆ back to top](#contents)**")

	for _, section := range sections {
		sectionLines := strings.Split(section, "\n")

		i := 0
		for i < len(sectionLines) {
			line := sectionLines[i]
			// We have found the "Parent" node
			// Its possible it has children with the ### item.
			if strings.Contains(line, "##") && !strings.Contains(line, "###") {
				normalizedName := strings.Replace(line, "## ", "", 3)
				c := &Category{
					Name:    normalizedName,
					Entries: []Entry{},
				}
				fmt.Println(c, store)
				break
			}
			i += 1
		}
	}
}
