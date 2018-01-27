package tychus

import "fmt"

type op uint32

// Possible operations
const (
	restarted op = 1 << iota
	rebuilt
	changed
	errored
)

type event struct {
	op   op
	info string
}

func (e event) String() string {
	return fmt.Sprintf("Event: %v", e.info)
}
