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

type Orchestrator struct {
	config  *Configuration
	watcher *watcher
	runner  *runner
	proxy   *proxy
}

func New(args []string, c *Configuration) *Orchestrator {
	return &Orchestrator{
		config:  c,
		watcher: newWatcher(c),
		runner:  newRunner(c, args),
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

	err := o.runner.run()
	if err != nil {
		o.proxy.error(err)
	}

	for {
		select {
		case <-o.proxy.requests:
			o.config.Logger.Debug("Request started")

			modified := o.watcher.scan()
			if modified {
				err := o.runner.run()
				if err != nil {
					o.proxy.error(err)
					continue
				}
			}

			o.proxy.serve()

		case err := <-o.runner.errors:
			o.config.Logger.Debug("Runner errored")
			o.proxy.error(err)

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
