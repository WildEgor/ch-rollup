package dump

import (
	"errors"
	"log"
	"sync"
	"time"
)

type InMemoryDumper struct {
	storage map[string]string // keep json-string
	mu      sync.RWMutex
}

func NewInMemoryDumper() Dumper {
	return &InMemoryDumper{
		storage: make(map[string]string),
	}
}

func (imd *InMemoryDumper) Dump(id string, content string) error {
	imd.mu.Lock()
	defer imd.mu.Unlock()
	key := imd.makeKey(id)
	imd.storage[key] = content
	return nil
}

func (imd *InMemoryDumper) ProcessNext(sender Sender) error {
	imd.mu.RLock()
	defer imd.mu.Unlock()

	for key, content := range imd.storage {
		err := sender(key, content)
		if err != nil {
			return err
		}

		delete(imd.storage, key)
	}

	return nil
}

func (imd *InMemoryDumper) Listen(sender Sender, interval int) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	go func() {
		for range ticker.C {
			for {
				err := imd.ProcessNext(sender)
				if err != nil {
					if !errors.Is(err, ErrNoDumps) {
						log.Printf("ERROR: %+v\n", err)
					}
					break
				}
			}
		}
	}()
}

func (imd *InMemoryDumper) makeKey(id string) string {
	return id
}
