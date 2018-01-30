// Package tychus is a command line application that will watch your files and
// on change, trigger a rerun of a command. It's designed to work best with web
// applications, but certainly not lmited to.
//
// Unlike other application reloaders written in Go, Tychus is language
// agnostic. It can be used with Go, Rust, Python, Ruby, scripts, etc.
//
// If enabled, Tychus will serve an application through a proxy. This can help
// mitigate annoyances like reloading your web page before the app server
// finishes booting. Or attempting to make a request after the server starts,
// but before it is ready to accept requests.
package tychus

import (
	"github.com/devlocker/devproxy/devproxy"
)

type Orchestrator struct {
	config  *Configuration
	watcher *watcher
	runner  *runner
	proxy   *devproxy.Proxy
}

func New(args []string, c *Configuration) *Orchestrator {
	return &Orchestrator{
		config:  c,
		watcher: newWatcher(),
		runner:  newRunner(args),
		proxy: devproxy.New(&devproxy.Configuration{
			AppPort:   c.AppPort,
			ProxyPort: c.ProxyPort,
			Timeout:   c.Timeout,
			Logger:    c.Logger,
		}),
	}
}

// Starts Tychus. Any filesystem changes will cause the command passed in to be
// rerun. To avoid orphaning processes, make sure to call Stop before exiting.
func (o *Orchestrator) Start() error {
	stop := make(chan error, 1)

	go o.watcher.start(o.config)
	go o.runner.start(o.config)

	if o.config.ProxyEnabled {
		go func() {
			err := o.proxy.Start()
			if err != nil {
				stop <- err
			}
		}()
	}

	o.runner.restart <- true

	for {
		select {
		case event := <-o.watcher.events:
			o.config.Logger.Debug(event)
			switch event.op {
			case changed:
				o.proxy.Command <- devproxy.Command{
					Cmd: devproxy.Pause,
				}
				o.runner.restart <- true
			}

		case event := <-o.runner.events:
			o.config.Logger.Debug(event)
			switch event.op {
			case restarted:
				o.proxy.Command <- devproxy.Command{
					Cmd: devproxy.Serve,
				}
			case errored:
				o.proxy.Command <- devproxy.Command{
					Cmd:  devproxy.Error,
					Data: event.info,
				}
			}

		case err := <-stop:
			o.Stop()
			return err
		}
	}
}

// Stops Tychus and forces any processes started by it that may be running.
func (o *Orchestrator) Stop() error {
	return o.runner.kill()
}
