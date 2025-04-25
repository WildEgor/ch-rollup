// Copyright 2025 LLC "Ozon Technologies".
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"context"
	"encoding/json"

	"golang.org/x/exp/maps"

	"github.com/ozontech/ch-rollup/pkg/rollup"
	"github.com/ozontech/ch-rollup/pkg/types"
)

const (
	tempTablePrefix  = "_temp"
	tryRollUpEachSec = 30
)

func (s *Scheduler) rollUp(ctx context.Context) error {
	for _, task := range s.tasks {
		for _, rollUpSetting := range task.RollUpSettings {
			ro := rollup.RunOptions{
				Database:     task.Database,
				Table:        task.Table,
				TempTable:    task.Table + tempTablePrefix,
				PartitionKey: task.PartitionKey,
				Columns:      prepareRollUpColumns(task.ColumnSettings, rollUpSetting.ColumnSettings),
				Interval:     rollUpSetting.Interval,
				After:        rollUpSetting.After,
				CopyInterval: task.CopyInterval,
			}
			err := s.dbRollUp.Run(ctx, ro)
			if err != nil {
				if s.dumper != nil {
					if b, mErr := json.Marshal(&ro); mErr != nil {
						return s.dumper.Dump(task.ID, string(b))
					}
				}

				return err
			}
		}
	}

	return nil
}

func (s *Scheduler) tryRollUp(ctx context.Context) {
	if s.dumper != nil {
		s.dumper.Listen(func(id string, content string) error {
			var ro rollup.RunOptions
			if err := json.Unmarshal([]byte(content), &ro); err != nil {
				// TODO
				return nil
			}
			err := s.dbRollUp.Run(ctx, ro)
			if err != nil {
				// TODO
				return nil
			}

			return nil
		}, tryRollUpEachSec)
	}
}

func prepareRollUpColumns(globalColumnSettings, currentColumnSettings []types.ColumnSetting) []types.ColumnSetting {
	result := make(map[string]types.ColumnSetting, len(globalColumnSettings)+len(currentColumnSettings))

	for _, columnSettings := range globalColumnSettings {
		result[columnSettings.Name] = columnSettings
	}

	for _, columnSettings := range currentColumnSettings {
		result[columnSettings.Name] = columnSettings
	}

	return maps.Values(result)
}
