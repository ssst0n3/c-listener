package ebpf

import "github.com/ssst0n3/fd-listener/pkg/listener/fd/event"

type Header struct{}

type Message struct {
	Header Header
	Data   []byte
}

func (m Message) parse() (events event.Events, err error) {
	return
}
