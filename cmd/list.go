package cmd

import (
	"os/exec"
	"sort"
	"strings"

	"github.com/skye-lopez/go-get-cli/interaction"
	"github.com/spf13/cobra"
)

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "display a list of packages via filters",
	Long: ` List and install packages.
    ~~~~List Packages by category~~~~
    go-get-cli list -c
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

    ~~~List all available packages~~~
    go-get-cli list -a 

    NOTE: see go-get-cli search if you are looking for a specific package.
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
    `,

	Run: list,
}

func init() {
	rootCmd.AddCommand(listCommand)
	listCommand.Flags().BoolP("categories", "c", false, "List all available categories and their subprojects")
	listCommand.Flags().BoolP("all", "a", false, "List all available packages")
}

// TODO: SHOW CURRENT PAGE DURING PAGINATED REQUESTS
func list(cmd *cobra.Command, args []string) {
	categories, _ := cmd.Flags().GetBool("categories")
	all, _ := cmd.Flags().GetBool("all")
	sort.Slice(data.Categories, func(i, j int) bool {
		return data.Categories[i].Name < data.Categories[j].Name
	})

	if all && !categories {
		i := interaction.NewInteraction()
		homePrompt := i.CreatePrompt("All packages (sorted by name)", "[n] Next Page | [b] Last Page | [esc] Exit | [enter] Select", true)

		for _, v := range data.Entries {
			if len(v.Name) <= 0 {
				continue
			}

			option := homePrompt.AddOption(v.Name+" [category: "+v.Category+"]", v.Description, v)
			entryPrompt := i.CreatePrompt(v.Name+"( "+v.Description+" )", "[enter] Select | [u] Back to list | [esc] Exit", false)
			entryPrompt.AttachParent(homePrompt.Idx)
			option.AttachPrompt(entryPrompt.Idx)

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

		i.Open()
	}

	if categories && !all {
		i := interaction.NewInteraction()
		homePrompt := i.CreatePrompt("Available packages by category:", "[n] Next page | [b] Last page | [esc] Exit | [enter] Select", true)

		for _, v := range data.Categories {
			// TODO: Some are empty fsr...
			if v.Name == "" {
				continue
			}
			option := homePrompt.AddOption(v.Name, v.Description, v)

			categoryPrompt := i.CreatePrompt(v.Name+" - Packages ("+v.Description+") ", "[n] Next page | [b] Last page | [enter] Select | [u] Back to categories | [esc] Exit", true)
			categoryPrompt.AttachParent(homePrompt.Idx)

			option.AttachPrompt(categoryPrompt.Idx)

			for _, ov := range v.Entries {
				if len(ov.Name) < 2 {
					continue
				}
				catOption := categoryPrompt.AddOption(ov.Name, ov.Description, ov)

				entryPrompt := i.CreatePrompt(ov.Name+"( "+ov.Description+" )", "[enter] Select | [u] Back to category | [esc] Exit", false)
				entryPrompt.AttachParent(categoryPrompt.Idx)

				catOption.AttachPrompt(entryPrompt.Idx)

				installOption := entryPrompt.AddOption("Install via go get (gitlab/github package only)", "go get "+ov.Link, ov)

				installFunc := func(...any) (string, error) {
					goPath, err := exec.LookPath("go")
					// This likely means the user does not have a go PATH set to $PATH
					if err != nil {
						panic(err)
					}

					// Format link for install
					// /usr/local/go/bin/go go get  https://github.com/guptarohit/asciigraph
					var installCandidate string

					if strings.Contains(ov.Link, "https://") {
						installCandidate = strings.Split(ov.Link, "https://")[1]
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
		}

		i.Open()
	}
}
