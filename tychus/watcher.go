package tychus

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher
	events chan event
}

func newWatcher() *watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Could not start watcher")
	}

	return &watcher{
		Watcher: w,
		events:  make(chan event),
	}
}

func (w *watcher) start(c *Configuration) {
	go w.watchForChanges(c)

	// With the way most editors work, they will generate multiple events on
	// save. To avoid excessive restarts, batch all messages together within
	// some interval and notify the orchestrator once per interval.
	tick := time.Tick(200 * time.Millisecond)
	events := make([]fsnotify.Event, 0)

	for {
		select {
		case event := <-w.Events:
			if event.Op == fsnotify.Chmod {
				continue
			}

			if w.isWatchedFile(event.Name, c) {
				events = append(events, event)
			}

		case <-tick:
			if len(events) == 0 {
				continue
			}

			c.Logger.Debugf("FS Changes: %v", events)
			w.events <- event{info: "File System Change", op: changed}

			events = make([]fsnotify.Event, 0)
		}
	}
}

// Setup file watchers for all valid extensions and directories.
func (w *watcher) watchForChanges(c *Configuration) error {
	for {
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return errors.New("nil directory")
			}

			if info.IsDir() {
				if w.shouldSkipDir(path, c) {
					return filepath.SkipDir
				}

				c.Logger.Debugf("Watching: %v", path)
				w.Add(path)
			}

			return nil
		})

		if err != nil {
			break
		}

		// Walk once a second.
		time.Sleep(1 * time.Second)
	}

	return errors.New("Watcher died")
}

// Checks to see if this directory should be watched. Don't want to watch
// hidden directories (like .git) or ignored directories.
func (w *watcher) shouldSkipDir(path string, c *Configuration) bool {
	if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	for _, f := range c.Ignore {
		f = strings.TrimSpace(f)
		f = strings.TrimRight(f, "/")

		if f == path {
			return true
		}
	}

	return false
}

// Checks to see if path matches a configured extension.
func (w *watcher) isWatchedFile(path string, c *Configuration) bool {
	if len(c.Extensions) == 0 {
		return true
	}

	ext := filepath.Ext(path)
	for _, e := range c.Extensions {
		if e == ext {
			return true
		}
	}

	return false
}
