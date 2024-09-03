package cmd

import (
	"fmt"
	"sort"

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
		display := NewMenu("Available categories - [N] Next Page | [B] Last Page | [ENTER] Select | [ESC] Exit")
		for _, v := range data.Categories {
			if v.Name == "" {
				continue
			}
			displayName := v.Name
			if v.Description != "" {
				displayName += " (" + v.Description + ")                                                       "
			}
			display.AddItem(displayName, v.Name, v)
		}

		result := display.Display()
		fmt.Println(result)
	}
}
