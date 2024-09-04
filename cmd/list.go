package cmd

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/skye-lopez/go-get-cli/interaction"
	"github.com/spf13/cobra"
)

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "display a list via filters",
	Long: `List packages with certain filters
    Usage examples:

    ~~~List Packages by category~~~
    go-get-cli list -c`,

	Run: list,
}

func init() {
	rootCmd.AddCommand(listCommand)
	listCommand.Flags().BoolP("categories", "c", false, "List all available categories")
}

func list(cmd *cobra.Command, args []string) {
	categories, _ := cmd.Flags().GetBool("categories")

	if categories {
		sort.Slice(data.Categories, func(i, j int) bool {
			return data.Categories[i].Name < data.Categories[j].Name
		})

		i := interaction.NewInteraction()
		homePrompt := i.CreatePrompt("Available packages by category:", "[n] Next | [b] Last | [esc] Exit | [enter] Select", true)
		fmt.Println(homePrompt.Idx)

		for _, v := range data.Categories {
			// TODO: Some are empty fsr...
			if v.Name == "" {
				continue
			}
			option := homePrompt.AddOption(v.Name, v.Description, v)

			categoryPrompt := i.CreatePrompt(v.Name+" - Packages ("+v.Description+") ", "[n] Next | [b] Last | [esc] Exit | [enter] Select", true)
			categoryPrompt.AttachParent(homePrompt.Idx)

			option.AttachPrompt(categoryPrompt.Idx)

			for _, ov := range v.Entries {
				if len(ov.Name) < 2 {
					continue
				}
				catOption := categoryPrompt.AddOption(ov.Name, ov.Description, ov)

				entryPrompt := i.CreatePrompt(ov.Name, ov.Description, false)
				entryPrompt.AttachParent(categoryPrompt.Idx)

				catOption.AttachPrompt(entryPrompt.Idx)

				installOption := entryPrompt.AddOption("Install via go get", "go get "+ov.Link, ov)

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
