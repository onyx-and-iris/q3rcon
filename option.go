package q3rcon

import "time"

// Option is a functional option type that allows us to configure the VbanTxt.
type Option func(*Rcon)

// WithLoginTimeout is a functional option to set the login timeout
func WithLoginTimeout(timeout time.Duration) Option {
	return func(rcon *Rcon) {
		rcon.loginTimeout = timeout
	}
}

// WithDefaultTimeout is a functional option to set the default response timeout
func WithDefaultTimeout(timeout time.Duration) Option {
	return func(rcon *Rcon) {
		rcon.defaultTimeout = timeout
	}
}

// WithTimeouts is a functional option to set the timeouts for responses per command
func WithTimeouts(timeouts map[string]time.Duration) Option {
	return func(rcon *Rcon) {
		rcon.timeouts = timeouts
	}
}
