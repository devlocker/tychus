// Package tychus is a command line application that will watch your files and
// on change, trigger a rerun of a command. It's designed to work best with web
// applications, but certainly not lmited to.
//
// Unlike other application reloaders written in Go, Tychus is language
// agnostic. It can be used with Go, Rust, Python, Ruby, scripts, etc.
//
// Tychus has 3 parts to it's configuration.
//
// 1. Watch: configures what extensions to watch in what directories. A change
// to a watched file will trigger an application reload.
//
// 2. Build: if enabled, adds a build step (compile) after a file system change
// is detected.
//
// 3. Proxy: if enabled, will serve an application through a proxy. This can
// help mitigate annoyances like reloading your web page before the app server
// finishes booting (and getting an error page).
package tychus

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/devlocker/devproxy/devproxy"
)

func Start(args []string, c *Configuration) error {
	args, err := formatArgs(args, c)
	if err != nil {
		return err
	}

	w := newWatcher()
	b := newBuilder(c)
	r := newRunner(args)
	p := devproxy.New(&devproxy.Configuration{
		AppPort:   c.Proxy.AppPort,
		ProxyPort: c.Proxy.ProxyPort,
		Timeout:   c.Proxy.Timeout,
		Logger:    c.Logger,
	})

	go w.start(c)
	go b.start(c)
	go r.start(c)

	if c.Proxy.Enabled {
		go p.Start()
	}

	if c.Build.Enabled {
		b.rebuild <- true
	} else {
		r.restart <- true
	}

	for {
		select {
		// Watcher events
		case event := <-w.events:
			c.Logger.Debug(event)
			switch event.op {
			case changed:
				if c.Build.Enabled {
					p.Command <- devproxy.Command{Cmd: devproxy.Pause}
					b.rebuild <- true
				} else {
					r.restart <- true
				}
			}

		// Builder events
		case event := <-b.events:
			c.Logger.Debug(event)
			switch event.op {
			case rebuilt:
				c.Logger.Success("Build: Successful")
				r.restart <- true
			case errored:
				c.Logger.Error("Build: Failed\n" + event.info)
				p.Command <- devproxy.Command{Cmd: devproxy.Error, Data: event.info}
			}

		// Runner events
		case event := <-r.events:
			c.Logger.Debug(event)
			switch event.op {
			case restarted:
				p.Command <- devproxy.Command{Cmd: devproxy.Serve}
			case errored:
				p.Command <- devproxy.Command{Cmd: devproxy.Error, Data: event.info}
			}
		}
	}
}

// Format arguments to take into account any build targets, bin names. And make
// sure to expand any quotes strings.
func formatArgs(args []string, c *Configuration) ([]string, error) {
	if c.Build.Enabled {
		args = append([]string{filepath.Join(
			c.Build.TargetPath,
			c.Build.BinName,
		)}, args...)
	}

	// Can occur when running with build disabled. Since no binary to run, need
	// some command to run, e.g. "ruby myapp.rb"
	if len(args) < 1 {
		return nil, errors.New("Not enough arguments")
	}

	// Expand quoted strings - split by whitespace:
	// []string{"ls -al"} => []string{"ls", "-al"}.
	for i, a := range args {
		args = append(args[:i], append(strings.Fields(a), args[i+1:]...)...)
	}

	return args, nil
}
