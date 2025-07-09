package worker

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	Method     string
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

	// Store step responses for cross-step variable references
	stepResponses := make(map[string]map[string]interface{})

	// Merge test case variables with step responses for variable resolution
	allVariables := make(map[string]interface{})
	for k, v := range testCase.Variables {
		allVariables[k] = v
	}
	stepIdx := -1
	for stepKey, step := range testCase.Steps {
		stepIdx++
		// Extract method name from step key (e.g., "wallet_sendTBlockC")
		methodName := stepKey
		step.Method = methodName
		executor, found := api.GetAPIExecutor(methodName)
		if !found {
			continue
		}

		// Prepare variables for this step
		stepVariables := make(map[string]interface{})
		for k, v := range allVariables {
			stepVariables[k] = v
		}

		// Add request variables if defined
		if step.Request != nil {
			for k, v := range step.Request {
				// Resolve cross-step variable references
				resolvedValue := w.resolveVariableReference(v, stepResponses)
				stepVariables[k] = resolvedValue
			}
		}
		var resp map[string]interface{}
		for step.Loop > 0 || (step.Loop <= -1 && step.Loop+step.MaxRetry >= 0) {
			step.Loop--
			// Execute the API call
			resp = executor(stepVariables)

			if code, ok := resp["code"]; ok && code != 200 {
				return
			}
			// Check for errors
			if errorVal, ok := resp["error"]; ok {
				// For infinite loop (Loop <= -1), continue to next iteration
				// Otherwise return to stop processing
				if step.Loop > -1 || step.Loop+step.MaxRetry < 0 {
					w.ResultsChan <- Result{
						Method:   step.Method,
						Duration: time.Since(startTime),
						Error:    errors.New(fmt.Sprintf("%v", errorVal)),
					}
					return
				}
				time.Sleep(time.Duration(step.Interval) * time.Millisecond)
				continue
			}
			// No error case
			// Break infinite loop if no error
			if step.Loop <= -1 {
				break
			}
			time.Sleep(time.Duration(step.Interval) * time.Millisecond)
		}

		// Store step response for future reference
		if step.ID == "" {
			step.ID = strconv.Itoa(stepIdx)
		}

		// Merge additional response fields if defined
		if step.Response != nil {
			for k, v := range step.Response {
				resp[k] = v
			}
		}
		stepResponses[step.ID] = map[string]interface{}{
			"response": resp,
		}
	}

	duration := time.Since(startTime)
	w.ResultsChan <- Result{
		Duration:   duration,
		StatusCode: 200,
	}
}

// resolveVariableReference resolves cross-step variable references like "sendTBlockC.response.receipt"
func (w *Worker) resolveVariableReference(value interface{}, stepResponses map[string]map[string]interface{}) interface{} {
	if strVal, ok := value.(string); ok {
		// Check if it's a cross-step reference (contains dots)
		if strings.Contains(strVal, ".") {
			parts := strings.Split(strVal, ".")
			if len(parts) >= 3 && parts[1] == "response" {
				stepID := parts[0]
				if stepResp, exists := stepResponses[stepID]; exists {
					return w.getNestedValue(stepResp, parts[1:])
				}
			}
		}
	}
	return value
}

// getNestedValue retrieves a nested value from a map using dot notation
func (w *Worker) getNestedValue(data map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return data
	}

	current := data
	for i, key := range path {
		if i == len(path)-1 {
			return current[key]
		}

		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			// Try to convert to map if it's a different type
			if nextVal := current[key]; nextVal != nil {
				v := reflect.ValueOf(nextVal)
				if v.Kind() == reflect.Map {
					nextMap := make(map[string]interface{})
					for _, mapKey := range v.MapKeys() {
						nextMap[fmt.Sprintf("%v", mapKey.Interface())] = v.MapIndex(mapKey).Interface()
					}
					current = nextMap
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}
	return current
}
