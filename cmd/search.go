package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCommand = &cobra.Command{
	Use:   "search",
	Short: "Search for a specific package with options",
	Long: `Search for a specific package with filter options!
    Usage examples:

    ~~~Search By Term~~~
    go-get-cli search -t <YOUR_SEARCH_TERM_HERE>`,

	Run: search,
}

func init() {
	rootCmd.AddCommand(searchCommand)
	searchCommand.Flags().StringP("term", "t", "default", "The term to search the index by")
}

func search(cmd *cobra.Command, args []string) {
	term, _ := cmd.Flags().GetString("term")
	fmt.Println(term)
}
