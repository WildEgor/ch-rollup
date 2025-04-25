// Copyright 2025 LLC "Ozon Technologies".
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"context"
	"fmt"
	"github.com/ozontech/ch-rollup/internal/utils/retry"
	"golang.org/x/exp/maps"
	"strconv"
	"time"

	"github.com/ozontech/ch-rollup/pkg/rollup"
	"github.com/ozontech/ch-rollup/pkg/types"
)

const (
	tempTablePrefix      = "_temp"
	rollUpBackoffEachSec = 30
	_retryTimes          = 3
	_retryTimeout        = 5 * time.Second
)

func (s *Scheduler) rollUp(ctx context.Context) error {
	for _, task := range s.tasks {
		for id, rollUpSetting := range task.RollUpSettings {
			err := s.dbRollUp.Run(ctx, rollup.RunOptions{
				Database:     task.Database,
				Table:        task.Table,
				TempTable:    task.Table + tempTablePrefix,
				PartitionKey: task.PartitionKey,
				Columns:      prepareRollUpColumns(task.ColumnSettings, rollUpSetting.ColumnSettings),
				Interval:     rollUpSetting.Interval,
				After:        rollUpSetting.After,
				CopyInterval: task.CopyInterval,
			})
			if err != nil && s.dumper != nil {
				return s.dumper.Dump(task.ID, strconv.Itoa(id))
			}

			return err
		}
	}

	return nil
}

func (s *Scheduler) tryRollUp(ctx context.Context) {
	if s.dumper == nil {
		return
	}

	s.dumper.Listen(ctx, func(_ string, content string) error {
		var (
			rollUpSettingRef *types.RollUpSetting
			taskRef          *types.Task
		)

		for _, t := range s.tasks {
			for id, ro := range taskRef.RollUpSettings {
				roID, err := strconv.ParseInt(content, 0, 64)
				if err != nil {
					return err
				}

				if int(roID) == id {
					taskRef = &t
					rollUpSettingRef = &ro
				}
			}
		}

		if taskRef == nil || rollUpSettingRef == nil {
			return nil
		}

		return retry.Retry(_retryTimes, _retryTimeout, func(attempt int64) error {
			err := s.dbRollUp.Run(ctx, rollup.RunOptions{
				Database:     taskRef.Database,
				Table:        taskRef.Table,
				TempTable:    taskRef.Table + tempTablePrefix,
				PartitionKey: taskRef.PartitionKey,
				Columns:      prepareRollUpColumns(taskRef.ColumnSettings, rollUpSettingRef.ColumnSettings),
				Interval:     rollUpSettingRef.Interval,
				After:        rollUpSettingRef.After,
				CopyInterval: taskRef.CopyInterval,
			})

			if attempt == _retryTimes && err != nil {
				fmt.Printf("retry error %v", err)
			}

			return err
		})
	}, s.opts.retrySec)
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
