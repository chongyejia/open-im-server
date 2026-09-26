package options

// Opts opts.
type Opts struct {
	SendID        string
	GroupID       string
	SessionType   int32
	Signal        *Signal
	IOSPushSound  string
	IOSBadgeCount bool
	Ex            string
}

// Signal message id.
type Signal struct {
	ClientMsgID string
}
