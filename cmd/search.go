package cmd

import (
	"os/exec"
	"strings"

	"github.com/skye-lopez/go-get-cli/interaction"
	"github.com/spf13/cobra"
)

var searchCommand = &cobra.Command{
	Use:   "search",
	Short: "Search for a specific package with options",
	Long: `Search for a specific package
    Usage examples:

    ~~~Search By Text~~~
    go-get-cli search -t <YOUR_SEARCH_TERM_HERE>`,

	Run: search,
}

func init() {
	rootCmd.AddCommand(searchCommand)
	searchCommand.Flags().BoolP("search", "", true, "Start a search session.")
}

func search(cmd *cobra.Command, args []string) {
	search, _ := cmd.Flags().GetBool("search")

	if !search {
		return
	}

	s := interaction.NewSearchInteraction()
	homePrompt := s.CreatePrompt("Search for a pacakge. Result will filter as you type.", "[n] Next Page | [b] Last Page | [+] Select search bar | [r] Select results | [enter] Select prompt | [esc] Exit", true)

	for _, v := range data.Entries {
		entryOption := homePrompt.AddOption(v.Name, v.Description+" [Category: "+v.Category+"]", v)
		s.UpdateTrie(entryOption)

		entryPrompt := s.CreatePrompt(v.Name+"( "+v.Description+" )", "[enter] Select | [u] Back to list | [esc] Exit", true)
		entryPrompt.AttachParent(homePrompt.Idx)
		entryOption.AttachPrompt(entryPrompt.Idx)

		installOption := entryPrompt.AddOption("Install via go get (gitlab/github package only)", "go get "+v.Link, v)

		installFunc := func(...any) (string, error) {
			goPath, err := exec.LookPath("go")
			// This likely means the user does not have a go PATH set to $PATH
			if err != nil {
				panic(err)
			}

			// Format link for install
			// /usr/local/go/bin/go go get  https://github.com/guptarohit/asciigraph
			var installCandidate string

			if strings.Contains(v.Link, "https://") {
				installCandidate = strings.Split(v.Link, "https://")[1]
			}

			install := exec.Command(goPath, "get", installCandidate)
			err = install.Run()
			if err != nil {
				return "Error installing the selected package.", err
			}
			return "Package installed! Have fun :)", nil
		}
		installOption.AddCallback(installFunc)
	}

	s.Open()

	// What I want
	// A search bar
	// Results that get filtered per key press in...
	//
	// We may need to implement a custom interaction
}
