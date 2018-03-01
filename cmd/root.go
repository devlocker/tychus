package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/devlocker/tychus/tychus"
	"github.com/spf13/cobra"
)

var version = "0.6.2"

var appPort int
var debug bool
var ignored []string
var noProxy bool
var proxyPort int
var timeout int
var wait bool

var rootCmd = &cobra.Command{
	Use:   "tychus",
	Short: "Live reload utility + proxy",
	Long: `Tychus is a command line utility for live reloading applications.
Tychus serves your application through a proxy. Anytime the proxy receives
an HTTP request will automatically rerun your command if the filesystem has
changed.
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		start(args)
	},
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&appPort, "app-port", "a", 3000, "port your application runs on, overwritten by ENV['PORT']")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "print debug output")
	rootCmd.Flags().StringSliceVarP(&ignored, "ignore", "x", []string{"node_modules", "log", "tmp", "vendor"}, "comma separated list of directories to ignore file changes in.")
	rootCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "will not start proxy if set")
	rootCmd.Flags().IntVarP(&proxyPort, "proxy-port", "p", 4000, "proxy port")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "timeout for proxied requests")
	rootCmd.Flags().BoolVar(&wait, "wait", false, "Wait for command to finish before proxying a request")
}

func start(args []string) {
	// Catch signals, need to do make sure to stop any executing commands.
	// Otherwise they become orphan proccesses.
	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	// If PORT is set, use that instead of AppPort. For things like foreman
	// where ports are automatically assigned.
	envPort, ok := os.LookupEnv("PORT")
	if ok {
		if envPort, err := strconv.Atoi(envPort); err == nil {
			appPort = envPort
		}
	}

	// Clean up ignored dirs.
	for i, dir := range ignored {
		ignored[i] = strings.TrimRight(strings.TrimSpace(dir), "/")
	}

	// Create a configuration
	c := &tychus.Configuration{
		Ignore:       ignored,
		ProxyEnabled: !noProxy,
		ProxyPort:    proxyPort,
		AppPort:      appPort,
		Timeout:      timeout,
		Logger:       tychus.NewLogger(debug),
		Wait:         wait,
	}

	// Run tychus
	t := tychus.New(args, c)
	go func() {
		err := t.Start()
		if err != nil {
			t.Stop()
			c.Logger.Fatal(err.Error())
		}
	}()

	<-stop
	t.Stop()
}
