package cmd

import (
	"os"

	"github.com/myyrakle/mongery/internal/run"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mongery",
	Short: "generate mongery codes",
	Run: func(cmd *cobra.Command, args []string) {
		run.Generate()
	},
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
