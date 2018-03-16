package tychus

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MichaelTJones/walk"
)

type watcher struct {
	config  *Configuration
	lastRun time.Time
}

func newWatcher(c *Configuration) *watcher {
	return &watcher{
		config:  c,
		lastRun: time.Now(),
	}
}

func (w *watcher) scan() bool {
	w.config.Logger.Debug("Watcher: Start")
	start := time.Now()

	modified := walk.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && w.shouldSkipDir(path) {
			return walk.SkipDir
		}

		if info.ModTime().After(w.lastRun) {
			w.config.Logger.Debugf("Watcher: Found modified file: %v", path)
			return errors.New(path)
		}

		return nil
	})

	w.config.Logger.Debugf("Watcher: Scan finished: %v", time.Since(start))

	return modified != nil
}

// Checks to see if this directory should be watched. Don't want to watch
// hidden directories (like .git) or ignored directories.
func (w *watcher) shouldSkipDir(path string) bool {
	if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	for _, dir := range w.config.Ignore {
		if dir == path {
			return true
		}
	}

	return false
}
