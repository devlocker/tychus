package tychus

import (
	"log"
	"os"
)

type Logger interface {
	Debug(interface{})
	Debugf(string, ...interface{})
	Fatal(...interface{})
	Printf(string, ...interface{})
	Print(...interface{})
}

func NewLogger(debug bool) Logger {
	l := &logger{
		Logger: log.New(os.Stdout, "[tychus] ", 0),
		debug:  debug,
	}

	return l
}

type logger struct {
	*log.Logger
	debug bool
}

func (l *logger) Debug(msg interface{}) {
	if l.debug {
		l.Printf("DEBUG: %v", msg)
	}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if l.debug {
		l.Printf("DEBUG: "+format, args...)
	}
}
