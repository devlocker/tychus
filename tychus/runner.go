package tychus

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
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
			c.Logger.Debugf("Running: %v", strings.Join(r.args, " "))
			if err := r.rerun(); err != nil {
				r.events <- event{op: errored, info: err.Error()}
			}
		}()
	}
}

// Rerun the command.
func (r *runner) rerun() error {
	if r.cmd != nil && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		return nil
	}

	var stderr bytes.Buffer
	mw := io.MultiWriter(&stderr, os.Stderr)

	r.cmd = exec.Command("/bin/sh", "-c", strings.Join(r.args, " "))
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = mw

	// Setup a process group so when this process gets stopped, so do any child
	// process that it may spawn.
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := r.cmd.Start()
	if err != nil {
		return errors.New(stderr.String())
	}

	r.events <- event{info: "Restarted", op: restarted}

	if err := r.cmd.Wait(); err != nil {
		// Program errored. Only log it if it exit status is postive, as status
		// code -1 is returned when the process was killed by kill().
		if exiterr, ok := err.(*exec.ExitError); ok {
			ws := exiterr.Sys().(syscall.WaitStatus)
			if ws.ExitStatus() > 0 {
				r.events <- event{info: stderr.String(), op: errored}
			}
		}
	}

	return nil
}

// Kill the existing process & process group
func (r *runner) kill() error {
	if r.cmd != nil && r.cmd.Process != nil {
		if pgid, err := syscall.Getpgid(r.cmd.Process.Pid); err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		}

		syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)

		r.cmd = nil
	}

	return nil
}
