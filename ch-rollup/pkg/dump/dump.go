package dump

import (
	"context"
	"errors"
)

// ErrNoDumps - signal that dumps not found
var ErrNoDumps = errors.New("no dumps")

// Sender ...
type Sender func(id string, content string) error

//go:generate go run go.uber.org/mock/mockgen -source dump.go -package=mock -destination=mock/dump.go

// Dumper ...
type Dumper interface {
	// Dump content to path
	Dump(id string, content string) error
	// ProcessNext ...
	ProcessNext(sender Sender) error
	// Listen ...
	Listen(ctx context.Context, sender Sender, interval int)
}
