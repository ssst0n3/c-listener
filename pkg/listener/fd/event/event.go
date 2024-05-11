package event

import "github.com/ssst0n3/fd-listener/pkg/listener/fd/stat"

const (
	Unknown int = iota
	Close
	Open
	Change
)

type Event struct {
	Type int
	Pid  int
	Fd   int
	Stat stat.Stat
}

type Events []Event
