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
	"time"
)

type Orchestrator struct {
	config  *Configuration
	watcher *watcher
	runner  *runner
	proxy   *proxy
}

func New(args []string, c *Configuration) *Orchestrator {
	return &Orchestrator{
		config:  c,
		watcher: newWatcher(),
		runner:  newRunner(args),
		proxy:   newProxy(c),
	}
}

func (o *Orchestrator) Start() error {
	stop := make(chan error, 1)

	go func() {
		err := o.proxy.start()
		if err != nil {
			stop <- err
		}
	}()

	go o.watcher.start(o.config)
	go o.runner.start(o.config)

	o.runner.restart <- true

	for {
		select {

		// Proxy events: If a request comes in, pause the websever and start
		// scanning for changes. Unless the last time a command ran it errored.
		// In which case just rerun the command regardless of whether or not
		// the FS has been changed.
		case event := <-o.proxy.events:
			o.config.Logger.Debug(event)

			switch event.op {
			case requested:
				if o.proxy.mode == mode_errored {
					o.runner.restart <- true
				} else {
					o.watcher.scan <- true
				}

				o.proxy.pause()
			}

		// Watcher events: If FS has changed since the last time the watcher
		// checked, go ahead and trigger a restart. Otherwise, unpause the
		// proxy.
		case event := <-o.watcher.events:
			o.config.Logger.Debug(event)

			switch event.op {
			case changed:
				o.runner.restart <- true
			case unchanged:
				o.proxy.serve()
			}

		// Runner events. If restart successful, go ahead an unpause the proxy.
		// If the command exited with an error code, have the proxy display the
		// error message.
		case event := <-o.runner.events:
			o.config.Logger.Debug(event)

			switch event.op {
			case restarted:
				o.watcher.lastRun = time.Now()
				o.proxy.serve()
			case errored:
				o.proxy.error(event.info)
			}

		// Stop Tychus
		case err := <-stop:
			o.Stop()
			return err
		}
	}
}

// Stops Tychus and forces any processes started by it that may be running.
func (o *Orchestrator) Stop() {
	o.runner.kill()
}
