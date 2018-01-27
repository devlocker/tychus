package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configFile string
var port string
var debug bool

var rootCmd = &cobra.Command{
	Use:   "tychus",
	Short: "Restart and your application as you make changes",
	Run: func(cmd *cobra.Command, args []string) {
		start(args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".tychus.yml", "path to tychus config")
	rootCmd.Flags().StringVarP(&port, "port", "p", "", "Port to run on")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "print debug output")
}
