package tychus

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"
)

type runner struct {
	args    []string
	cmd     *exec.Cmd
	events  chan event
	restart chan bool
}

func newRunner(args []string) *runner {
	return &runner{
		args:    args,
		events:  make(chan event),
		restart: make(chan bool),
	}
}

func (r *runner) start(c *Configuration) {
	for {
		<-r.restart

		if err := r.kill(); err != nil {
			r.events <- event{op: errored, info: err.Error()}
		}

		go func() {
			if err := r.rerun(); err != nil {
				r.events <- event{op: errored, info: err.Error()}
			}
		}()
	}
}

// Kill running process. Give it a chance to exit cleanly, otherwise kill it
// after a certain timeout.
func (r *runner) kill() error {
	if r.cmd != nil && r.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			r.cmd.Wait()
			close(done)
		}()

		r.cmd.Process.Signal(os.Interrupt)

		select {
		case <-time.After(3 * time.Second):
			if err := r.cmd.Process.Kill(); err != nil {
				return err
			}
		case <-done:
		}

		r.cmd = nil
	}

	return nil
}

// Rerun the command.
func (r *runner) rerun() error {
	if r.cmd != nil && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		return nil
	}

	var stderr bytes.Buffer
	mw := io.MultiWriter(&stderr, os.Stderr)

	r.cmd = exec.Command(r.args[0], r.args[1:]...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = mw

	err := r.cmd.Start()
	if err != nil {
		return errors.New(stderr.String())
	}

	r.events <- event{info: "Restarted", op: restarted}

	if err := r.cmd.Wait(); err != nil {
		r.events <- event{info: stderr.String(), op: errored}
	}

	return nil
}
