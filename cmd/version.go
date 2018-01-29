package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "master"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Tychus",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Tychus Version: %v\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
