package fd

import "github.com/ssst0n3/fd-listener/pkg/listener/fd/event"

type Watcher interface {
	Watch(stop <-chan struct{}, event chan<- event.Events)
	Enable() (enabled bool)
}
