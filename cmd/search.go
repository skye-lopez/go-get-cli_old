package cmd

import (
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
	searchCommand.Flags().StringP("term", "t", "", "The term to search the index by")
}

func search(cmd *cobra.Command, args []string) {
	term, _ := cmd.Flags().GetString("term")

	// What I want
	// A search bar
	// Results that get filtered per key press in...
	//
	// We may need to implement a custom interaction
	if len(term) > 0 {
	}
}
