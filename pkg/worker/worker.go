package worker

import (
	"errors"
	"fmt"
	"time"

	"perf-tester/pkg/api"
	"perf-tester/pkg/config"
)

// Result holds the result of a single request
// @description Result holds the result of a single request
// @field Duration The duration of the request
// @field StatusCode The HTTP status code of the response
// @field Error An error if the request failed
type Result struct {
	Duration   time.Duration
	StatusCode int
	Error      error
}

// Worker is a single worker that sends requests to the JSON-RPC endpoint
// @description Worker is a single worker that sends requests to the JSON-RPC endpoint
// @field ID The ID of the worker
// @field Config The configuration for the performance test
// @field ResultsChan The channel to send results to
type Worker struct {
	ID          int
	Config      *config.Config
	ResultsChan chan<- Result
}

// NewWorker creates a new worker
// @param id The ID of the worker
// @param config The configuration for the performance test
// @param resultsChan The channel to send results to
// @return *Worker The new worker
func NewWorker(id int, config *config.Config, resultsChan chan<- Result) *Worker {
	return &Worker{
		ID:          id,
		Config:      config,
		ResultsChan: resultsChan,
	}
}

// Run starts the worker
func (w *Worker) Run(requests <-chan config.TestCase) {
	for req := range requests {
		w.sendRequest(req)
	}
}

// sendRequest sends a single JSON-RPC request
func (w *Worker) sendRequest(testCase config.TestCase) {
	if w.Config == nil {
		w.ResultsChan <- Result{Error: fmt.Errorf("worker config is nil")}
		return
	}
	startTime := time.Now()
	for _, step := range testCase.Steps {
		executor, found := api.GetAPIExecutor(step)
		if !found {
			continue
		}
		resp := executor(testCase.Variables)
		if _, ok := resp["error"]; !ok {
			w.ResultsChan <- Result{
				Duration: time.Since(startTime),
				Error:    errors.New(resp["error"].(string)),
			}
		}
	}
	duration := time.Since(startTime)
	w.ResultsChan <- Result{
		Duration:   duration,
		StatusCode: 200,
	}
}
