package controller

// Valid server states.
const (
	StateUnknown string = "UNKNOWN"
	StateRunning string = "RUNNING"
	StateStopped string = "STOPPED"
)

var definedServerStates = []string{StateRunning, StateStopped}
