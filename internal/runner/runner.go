// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package runner

import (
	"context"
	"solo/internal/requester"
	"sort"
	"sync"
	"time"
)

type Runner struct {
	service *requester.Service
}

func NewRunner(service *requester.Service) *Runner {
	return &Runner{service: service}
}

func (r *Runner) Run(ctx context.Context, opts RunnerOptions, onResult func(RunnerResult)) RunnerStats {
	var wg sync.WaitGroup
	resultChan := make(chan RunnerResult, opts.Iterations)

	// Semaphore to control concurrency
	sem := make(chan struct{}, opts.Concurrency)

	startTime := time.Now()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Dispatcher
	go func() {
	loop:
		for i := 0; i < opts.Iterations; i++ {
			select {
			case <-ctx.Done():
				break loop
			case sem <- struct{}{}:
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					defer func() { <-sem }()

					res, err := r.service.ExecuteRequest(ctx, opts.Request)

					var runnerRes RunnerResult
					runnerRes.Index = idx
					if err != nil {
						runnerRes.Error = err.Error()
						if opts.StopOnError {
							cancel()
						}
					} else {
						runnerRes.Response = res
					}

					resultChan <- runnerRes
					if onResult != nil {
						onResult(runnerRes)
					}
				}(i)
			}
		}
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and calculate stats
	stats := RunnerStats{
		StatusCounts: make(map[int]int),
	}

	var durations []int64

	for res := range resultChan {
		stats.TotalRequests++
		if res.Error != "" {
			stats.ErrorCount++
		} else if res.Response != nil {
			stats.SuccessCount++
			stats.StatusCounts[res.Response.StatusCode]++
			durations = append(durations, res.Response.Duration)
		}
	}

	totalDuration := time.Since(startTime)
	stats.TotalDuration = totalDuration.Milliseconds()

	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

		var sum int64
		stats.MinLatency = durations[0]
		stats.MaxLatency = durations[len(durations)-1]
		for _, d := range durations {
			sum += d
		}
		stats.AvgLatency = sum / int64(len(durations))

		p95Idx := int(float64(len(durations)) * 0.95)
		if p95Idx >= len(durations) {
			p95Idx = len(durations) - 1
		}
		stats.P95Latency = durations[p95Idx]
	}

	if totalDuration.Seconds() > 0 {
		stats.RequestsPerSec = float64(stats.TotalRequests) / totalDuration.Seconds()
	}

	return stats
}
