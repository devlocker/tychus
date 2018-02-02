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
	args   []string
	cmd    *exec.Cmd
	errors chan error
	stderr *bytes.Buffer
	config *Configuration
}

func newRunner(c *Configuration, args []string) *runner {
	return &runner{
		args:   args,
		config: c,
		errors: make(chan error),
	}
}

func (r *runner) run() error {
	r.kill()

	if err := r.execute(); err != nil {
		return err
	}

	if r.config.Wait {
		r.wait()
	} else {
		go r.wait()
	}

	return nil
}

func (r *runner) execute() error {
	if r.cmd != nil && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		return nil
	}

	r.stderr = &bytes.Buffer{}
	mw := io.MultiWriter(r.stderr, os.Stderr)

	r.cmd = exec.Command("/bin/sh", "-c", strings.Join(r.args, " "))
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = mw

	// Setup a process group so when this process gets stopped, so do any child
	// process that it may spawn.
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := r.cmd.Start(); err != nil {
		return errors.New(r.stderr.String())
	}

	return nil
}

// Wait for the command to finish. If the process exits with an error, only log
// it if it exit status is postive, as status code -1 is returned when the
// process was killed by runner#kill.
func (r *runner) wait() {
	err := r.cmd.Wait()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			ws := exiterr.Sys().(syscall.WaitStatus)
			if ws.ExitStatus() > 0 {
				r.errors <- errors.New(r.stderr.String())
			}
		}
	}
}

// Kill the existing process & process group
func (r *runner) kill() {
	if r.cmd != nil && r.cmd.Process != nil {
		if pgid, err := syscall.Getpgid(r.cmd.Process.Pid); err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		}

		syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)

		r.cmd = nil
	}
}
