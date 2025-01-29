package main

import (
	"context"
	"time"
)

type Scheduler struct {
	interval time.Duration
}

func NewScheduler(interval time.Duration) *Scheduler {
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}

	return &Scheduler{
		interval: interval,
	}
}

func (s *Scheduler) Run(ctx context.Context, doFn func(ctx context.Context) error) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	firstRun := true

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := doFn(ctx); err != nil {
				return err
			}
			if firstRun {
				firstRun = false
				ticker.Reset(s.interval)
			}
		}
	}
}
