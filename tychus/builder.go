package tychus

import (
	"bytes"
	"io"
	"os/exec"
	"path"
	"strings"
)

type builder struct {
	rebuild  chan bool
	events   chan event
	buildCmd []string
	cmd      *exec.Cmd
}

func newBuilder(c *Configuration) *builder {
	var buildCmd []string

	if c.Build.Enabled {
		buildCmd = append(
			strings.Fields(c.Build.BuildCommand),
			"-o",
			path.Join(c.Build.TargetPath, c.Build.BinName),
		)
	}

	return &builder{
		rebuild:  make(chan bool),
		events:   make(chan event),
		buildCmd: buildCmd,
	}
}

// Wait for rebuild messages. Once received, rebuilds the binary. Should
// another file system event occur while a build is running, the build will be
// aborted. This is to prevent the case where the user saves, triggering a
// rebuild then saves in while the build is occuring, causing another process
// to start or a binary is reloaded with now stale code.
func (b *builder) start(c *Configuration) {
	for {
		<-b.rebuild

		go func() {
			if b.cmd != nil && b.cmd.Process != nil {
				b.cmd.Process.Kill()
			}

			c.Logger.Debugf("Building: %v", strings.Join(b.buildCmd, " "))

			b.cmd = exec.Command(b.buildCmd[0], b.buildCmd[1:]...)

			var stderr bytes.Buffer
			b.cmd.Stdout = nil
			b.cmd.Stderr = io.MultiWriter(&stderr)

			err := b.cmd.Start()
			if err != nil {
				b.events <- event{info: stderr.String(), op: errored}
				return
			}

			if err = b.cmd.Wait(); err != nil {
				// If ProcessState exists, means process was not aborted by
				// another go routine - just failed to compile or some other
				// error that the user should be presented with.
				if b.cmd.ProcessState != nil {
					b.events <- event{info: stderr.String(), op: errored}
				}

				return
			}

			b.cmd = nil

			b.events <- event{info: "Rebuilt", op: rebuilt}
		}()
	}
}
