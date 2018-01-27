package tychus

import (
	"log"
	"os"

	"github.com/fatih/color"
)

type Logger interface {
	Debug(interface{})
	Debugf(string, ...interface{})
	Error(string)
	Fatal(...interface{})
	Printf(string, ...interface{})
	Success(string)
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

func (l *logger) Error(msg string) {
	l.Print(color.RedString(msg))
}

func (l *logger) Success(msg string) {
	l.Print(color.GreenString(msg))
}
