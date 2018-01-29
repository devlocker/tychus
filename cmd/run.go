package cmd

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/devlocker/tychus/tychus"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Reloads your application as you make changes to source files.",
	Run: func(cmd *cobra.Command, args []string) {
		start(args)
	},
}

func start(args []string) {
	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	// Load configuration and use default logger
	c := &tychus.Configuration{}
	err := c.Load(configFile)
	if err != nil {
		c.Logger.Fatal(err.Error())
	}

	c.Logger = tychus.NewLogger(debug)

	// If PORT is set, use that instead of AppPort. For things like foreman
	// where ports are automatically assigned.
	port, ok := os.LookupEnv("PORT")
	if ok {
		if appPort, err := strconv.Atoi(port); err == nil {
			c.AppPort = appPort
		}
	}

	o := tychus.New(args, c)

	// Run tychus
	go func() {
		err = o.Start()
		if err != nil {
			o.Stop()
			c.Logger.Fatal(err.Error())
		}
	}()

	<-stop
	// Have to call `Stop`
	o.Stop()
}

func isEnabledStr(b bool) string {
	if b {
		return color.GreenString("enabled")
	}

	return color.YellowString("disabled")
}
