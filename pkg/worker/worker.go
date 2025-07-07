package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
func (w *Worker) Run(requests <-chan config.Request) {
	for req := range requests {
		w.sendRequest(req)
	}
}

// sendRequest sends a single JSON-RPC request
func (w *Worker) sendRequest(req config.Request) {
	if w.Config == nil {
		w.ResultsChan <- Result{Error: fmt.Errorf("worker config is nil")}
		return
	}
	startTime := time.Now()

	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  req.Method,
		"params":  req.Params,
		"id":      w.ID,
	})
	if err != nil {
		w.ResultsChan <- Result{Error: err}
		return
	}

	resp, err := http.Post(w.Config.URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		w.ResultsChan <- Result{Error: err}
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	w.ResultsChan <- Result{
		Duration:   duration,
		StatusCode: resp.StatusCode,
	}
}
