// Simple data manager using JSON files.
// This is not ideal but ensures a user does not need sqlite or such to use this CLI tool.
// if this sucks or performance does I will change it.

package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func Init() {
	store := FetchAndParseMD()
	if _, err := os.Stat("store.json"); errors.Is(err, os.ErrNotExist) {
		jsonString, err := json.Marshal(store)
		if err != nil {
			panic(err)
		}
		os.WriteFile("store.json", jsonString, os.ModePerm)
	}
	// TODO: We need to add functionality to basically refresh this list.
}

func ReadFile(target *Store) {
	file, _ := os.ReadFile("store.json")
	json.Unmarshal(file, &target)
}

type Entry struct {
	Category    string
	Name        string
	Link        string
	Description string
}

type Category struct {
	Name        string
	Description string
	Entries     []Entry
}

type Store struct {
	Entries    []Entry
	Categories []Category
}

func FetchAndParseMD() Store {
	store := Store{
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

		c := Category{
			Name:        "",
			Description: "",
			Entries:     []Entry{},
		}

		j := 0
		for j < len(sectionLines) {
			line := sectionLines[j]

			// Case 1 - We find a line that contains a Category (## or ###)
			if strings.Contains(line, "##") {
				// If we find a SuperClass, we skip it if it has children classes
				search := j + 1
				for sectionLines[search] == "" {
					search += 1
				}

				// If it has a children class we just want to skip this line.
				if strings.Contains(sectionLines[search], "###") {
					j += 1
					continue
				}

				normalizedName := strings.Replace(line, "#", "", 3)
				c.Name = normalizedName

				// Otherwise we see if it has a Description
				if strings.Contains(sectionLines[search], "._") {
					c.Description = sectionLines[search]
					j = search + 1
					continue
				}
			}

			// Case 2 - The line is a link or package essentially
			// format example:
			// - [bingo](https://github.com/iancmcc/bingo) - Fast, zero-allocation, lexicographical-order-preserving packing of native types to bytes.
			if strings.Contains(line, "(http") {
				e := Entry{
					Category:    c.Name,
					Name:        "",
					Link:        "",
					Description: "",
				}

				title := match(line, "[", "]")
				link := match(line, "(", ")")
				e.Name = title
				e.Link = link

				// Attempt to find a description
				entrySections := strings.Split(line, ") -")
				if len(entrySections) > 1 {
					e.Description = entrySections[1]
				}

				c.Entries = append(c.Entries, e)
				store.Entries = append(store.Entries, e)
			}
			j += 1
		}
		store.Categories = append(store.Categories, c)
	}

	return store
}

func match(s string, openingBracket string, closingBracket string) string {
	i := strings.Index(s, openingBracket)
	if i >= 0 {
		j := strings.Index(s, closingBracket)
		if j >= 0 {
			return s[i+1 : j]
		}
	}
	return ""
}
