package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/devlocker/tychus/tychus"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var with string

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&with, "with", "w", "", "extra options to run with")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Reloads your application as you make changes to source files.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(with) > 1 {
			args = append(args, strings.Fields(with)...)
		}

		start(args)
	},
}

func start(args []string) {
	c := &tychus.Configuration{}
	c.Logger = tychus.NewLogger(debug)

	err := c.Load(configFile)
	if err != nil {
		c.Logger.Fatal(err.Error())
	}

	c.Logger.Printf(
		"Starting: build [%v], proxy [%v]",
		isEnabledStr(c.Build.Enabled),
		isEnabledStr(c.Proxy.Enabled),
	)

	// If PORT is set, use that instead of AppPort. For things like foreman
	// where ports are automatically assigned.
	port := os.Getenv("PORT")
	if len(port) > 0 {
		appPort, err := strconv.Atoi(port)
		if err == nil {
			c.Proxy.AppPort = appPort
		}
	}

	err = tychus.Start(args, c)
	if err != nil {
		c.Logger.Fatal(err.Error())
	}
}

func isEnabledStr(b bool) string {
	if b {
		return color.GreenString("enabled")
	}

	return color.YellowString("disabled")
}
