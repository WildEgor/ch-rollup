package scheduler

// Opts ...
type Opts struct {
	dumpKind string
	retrySec int
}

// Opt ...
type Opt func(*Opts)

// WithDump ...
func WithDump(kind string) Opt {
	return func(opts *Opts) {
		opts.dumpKind = kind
	}
}

// WithRetrySec ...
func WithRetrySec(sec int) Opt {
	return func(opts *Opts) {
		opts.retrySec = sec
	}
}
