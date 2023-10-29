/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "0.4.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mongery version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
