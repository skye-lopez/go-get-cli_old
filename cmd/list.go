package cmd

import (
	"sort"

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

		for _, v := range data.Categories {
			// TODO: Some are empty fsr...
			if v.Name == "" {
				continue
			}
			option := homePrompt.AddOption(v.Name, v.Description, v)
			categoryPrompt := i.CreatePrompt(v.Name+" - Packages ("+v.Description+") ", "[n] Next | [b] Last | [esc] Exit | [enter] Select", true)
			option.AttachPrompt(categoryPrompt.Idx)

			for _, ov := range v.Entries {
				if ov.Name == "" {
					continue
				}
				categoryPrompt.AddOption(ov.Name, ov.Description, ov)
			}
		}

		i.Open()
	}
}
