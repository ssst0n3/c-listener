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
	Stat *stat.Stat
}

type Events []Event

func (e Events) Map(m map[int]*stat.Stat) {
	for _, event := range e {
		switch event.Type {
		case Open, Change:
			m[event.Fd] = event.Stat
		case Close:
			delete(m, event.Fd)
		default:
			panic("unhandled default case")
		}
	}
}
