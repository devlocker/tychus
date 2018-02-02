package tychus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type watcher struct {
	events  chan event
	lastRun time.Time
	scan    chan bool
}

func newWatcher() *watcher {
	return &watcher{
		events:  make(chan event),
		lastRun: time.Now(),
		scan:    make(chan bool),
	}
}

func (w *watcher) start(c *Configuration) {
	for {
		<-w.scan

		c.Logger.Debug("Scan: Start")
		start := time.Now()

		modified := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if info.IsDir() && w.shouldSkipDir(path, c) {
				return filepath.SkipDir
			}

			if info.ModTime().After(w.lastRun) {
				return errors.New(path)
			}

			return nil
		})

		c.Logger.Debugf("Scan: took: %v", time.Since(start))

		w.lastRun = time.Now()

		if modified != nil {
			w.events <- event{op: changed, info: fmt.Sprintf("FS Change: %v", modified)}
		} else {
			w.events <- event{op: unchanged, info: "FS Unchanged"}
		}
	}
}

// Checks to see if this directory should be watched. Don't want to watch
// hidden directories (like .git) or ignored directories.
func (w *watcher) shouldSkipDir(path string, c *Configuration) bool {
	if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	for _, dir := range c.Ignore {
		if dir == path {
			return true
		}
	}

	return false
}
