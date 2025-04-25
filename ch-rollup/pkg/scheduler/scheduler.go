// Copyright 2025 LLC "Ozon Technologies".
// SPDX-License-Identifier: Apache-2.0

// Package scheduler implements ch-rollup scheduler.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/ozontech/ch-rollup/pkg/dump"
	"time"

	"github.com/ozontech/ch-rollup/pkg/rollup"
	"github.com/ozontech/ch-rollup/pkg/types"
)

//go:generate go run go.uber.org/mock/mockgen -source scheduler.go -package=mock -destination=mock/scheduler.go

// RollUp ...
type RollUp interface {
	Run(ctx context.Context, opts rollup.RunOptions) error
}

const (
	defaultSchedulerInterval = time.Hour
	defaultDumpCheckSec      = 30
)

// Scheduler of ch-rollup.
type Scheduler struct {
	opts     *Opts
	dumper   dump.Dumper
	tasks    []types.Task
	dbRollUp RollUp
}

var (
	errNewNilRollup = errors.New("rollUp must be not nil")
)

// New returns new Scheduler.
func New(tasks types.Tasks, rollUp RollUp, options ...Opt) (*Scheduler, error) {
	if err := tasks.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate tasks: %w", err)
	}

	if rollUp == nil {
		return nil, errNewNilRollup
	}

	s := &Scheduler{
		tasks:    tasks,
		dbRollUp: rollUp,
	}

	for _, opt := range options {
		opt(s.opts)
	}

	switch s.opts.dumpKind {
	case "in_memory":
		s.dumper = dump.NewInMemoryDumper()
	default:
		// TODO
	}

	return s, nil
}

var (
	errSchedulerNotInitialized = errors.New("scheduler not initialized")
)

// Run Scheduler.
func (s *Scheduler) Run(ctx context.Context) (<-chan Event, error) {
	if s == nil || s.dbRollUp == nil {
		return nil, errSchedulerNotInitialized
	}

	eventChan := make(chan Event)

	s.tryRollUp(ctx)

	go func() {
		defer close(eventChan)

		// Let's do first rollup immediately.
		eventChan <- Event{
			Type:  EventTypeRollUp,
			Error: s.rollUp(ctx),
		}

		ticker := time.NewTicker(defaultSchedulerInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				eventChan <- Event{
					Type:  EventTypeRollUp,
					Error: s.rollUp(ctx),
				}

				ticker.Reset(defaultSchedulerInterval)
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, nil
}
