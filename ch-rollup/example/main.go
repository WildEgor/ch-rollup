// Copyright 2025 LLC "Ozon Technologies".
// SPDX-License-Identifier: Apache-2.0

// Package main contains example of ch-rollup usage.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	staticCluster "github.com/ozontech/ch-rollup/pkg/database/static"
	"github.com/ozontech/ch-rollup/pkg/rollup"
	"github.com/ozontech/ch-rollup/pkg/scheduler"
	"github.com/ozontech/ch-rollup/pkg/types"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tasks := []types.Task{
		{
			ID:           "test",
			Database:     "default",
			Table:        "test_table_agg",
			PartitionKey: time.Hour * 24,
			RollUpSettings: []types.RollUpSetting{
				{
					After:    time.Hour * 24,
					Interval: time.Second,
					ColumnSettings: []types.ColumnSetting{
						{
							Name:       "rollup_interval",
							Expression: "3600",
						},
					},
				},
			},
			ColumnSettings: []types.ColumnSetting{
				{
					Name: "col1",
				},
				{
					Name:       "counter",
					Expression: "countStateMerge(counter)",
				},
				{
					Name:         "event_time",
					IsRollUpTime: true,
				},
			},
		},
	}

	cluster, err := staticCluster.New(ctx, staticCluster.NewOptions{
		Address:  "127.0.0.1:9000",
		Username: "default",
		Password: "ch-rollup",
	})
	if err != nil {
		panic(err)
	}

	s, err := scheduler.New(
		tasks,
		rollup.New(cluster),
		scheduler.WithDump("in_memory"),
		scheduler.WithRetrySec(10),
	)
	if err != nil {
		panic(err)
	}

	eventsChan, err := s.Run(ctx)
	if err != nil {
		panic(err)
	}

	for event := range eventsChan {
		fmt.Println(event.String()) // we can log/alert error of rollup here.
	}
}
