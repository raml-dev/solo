package runner

import (
	"yapla/internal/requester"
)

// RunnerOptions defines the parameters for a parallel execution run.
type RunnerOptions struct {
	Concurrency int                        `json:"concurrency"`
	Iterations  int                        `json:"iterations"`
	StopOnError bool                       `json:"stopOnError"`
	Request     requester.ExecutionOptions `json:"request"`
}

// RunnerResult represents the outcome of a single request within a run.
type RunnerResult struct {
	Index    int                     `json:"index"`
	Response *requester.ResponseData `json:"response"`
	Error    string                  `json:"error,omitempty"`
}

// RunnerStats provides aggregated metrics for the entire run.
type RunnerStats struct {
	TotalRequests int            `json:"totalRequests"`
	SuccessCount  int            `json:"successCount"`
	ErrorCount    int            `json:"errorCount"`
	MinLatency    int64          `json:"minLatency"`
	MaxLatency    int64          `json:"maxLatency"`
	AvgLatency    int64          `json:"avgLatency"`
	P95Latency    int64          `json:"p95Latency"`
	TotalDuration int64          `json:"totalDuration"` // Total time for the whole run
	RequestsPerSec float64       `json:"requestsPerSec"`
	StatusCounts  map[int]int    `json:"statusCounts"`
}
