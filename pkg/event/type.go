package event

type Type int

const (
	Unknown Type = iota
	ProcessNew
	ProcessExit
)
