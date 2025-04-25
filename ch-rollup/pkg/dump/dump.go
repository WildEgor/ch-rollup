package dump

type Sender func(id string, content string) error

type Dumper interface {
	// Dump content to path
	Dump(id string, content string) error
	// ProcessNext ...
	ProcessNext(sender Sender) error
	// Listen ...
	Listen(sender Sender, interval int)
}
