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

var version = "master"

var appPort int
var debug bool
var ignored []string
var noProxy bool
var proxyPort int
var timeout int
var watch []string

var rootCmd = &cobra.Command{
	Use:   "tychus",
	Short: "Starts and reloads your application as you make changes to source files.",
	Long: `tychus is a command line utility to live-reload your application. tychus will
watch your filesystem for changes and automatically recompile and restart code
on change.

Example:
  tychus go run main.go -w .go
  tychus ruby myapp.rb --app-port=4567 --proxy-port=4000 --watch .rb,.erb --ignore node_modules

Example: No Proxy
  tychus ls --no-proxy

Example: Flags - use quotes
  tychus "ruby myapp.rb -p 5000 -e development" -a 5000 -p 4000 -w .rb,.erb

Example: Multiple Commands - use quotes
  tychus "go build -o my-bin && echo 'Done Building' && ./my-bin"
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
	rootCmd.Flags().StringSliceVarP(&watch, "watch", "w", []string{}, "comma separated list of extensions that will trigger a reload. If not set, will reload on any file change.")
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

	// Clean up watched file extensions
	for i, ext := range watch {
		ext = strings.TrimSpace(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}

		watch[i] = ext
	}

	// If PORT is set, use that instead of AppPort. For things like foreman
	// where ports are automatically assigned.
	envPort, ok := os.LookupEnv("PORT")
	if ok {
		if envPort, err := strconv.Atoi(envPort); err == nil {
			appPort = envPort
		}
	}

	// Create a configuration
	c := &tychus.Configuration{
		Extensions:   watch,
		Ignore:       ignored,
		ProxyEnabled: !noProxy,
		ProxyPort:    proxyPort,
		AppPort:      appPort,
		Timeout:      timeout,
		Logger:       tychus.NewLogger(debug),
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
