package scheduler

type Opts struct {
	dumpKind string
}

type Opt func(*Opts)

func WithDump(kind string) Opt {
	return func(o *Opts) {
		o.dumpKind = kind
	}
}
