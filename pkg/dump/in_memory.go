package dump

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// InMemoryDumper ...
type InMemoryDumper struct {
	storage map[string]string
	mu      sync.RWMutex
}

// NewInMemoryDumper ...
func NewInMemoryDumper() Dumper {
	return &InMemoryDumper{
		storage: make(map[string]string),
	}
}

// Dump ...
func (imd *InMemoryDumper) Dump(id string, content string) error {
	imd.mu.Lock()
	defer imd.mu.Unlock()
	imd.storage[id] = content
	return nil
}

// ProcessNext ...
func (imd *InMemoryDumper) ProcessNext(sender Sender) error {
	imd.mu.RLock()
	defer imd.mu.RUnlock()

	for id, content := range imd.storage {
		err := sender(id, content)
		if err != nil {
			return err
		}

		delete(imd.storage, id)
	}

	return nil
}

// Listen ...
func (imd *InMemoryDumper) Listen(ctx context.Context, sender Sender, interval int) {
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(interval))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for {
					err := imd.ProcessNext(sender)
					if err != nil {
						if !errors.Is(err, ErrNoDumps) {
							fmt.Printf("ERROR: %+v\n", err)
						}
						break
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
