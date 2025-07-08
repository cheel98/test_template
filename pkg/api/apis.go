package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"perf-tester/pkg/config"
	"strings"
)

// APIRequest defines the structure for a single API request in api.json.
// Note that the params are kept as RawMessage to allow for flexible variable substitution.
type APIRequest struct {
	Request struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	} `json:"request"`
	Response json.RawMessage `json:"response"`
}

// APIExecutor is the function type for executing an API call.
// It takes a map of variables and returns the response body.
type APIExecutor func(variables map[string]interface{}) map[string]interface{}

var apiMap map[string]APIExecutor

// LoadApis loads all APIs from the given api.json file.
func LoadApis(apiFilePath string, cfg *config.Config) error {
	apiMap = make(map[string]APIExecutor)

	file, err := ioutil.ReadFile(apiFilePath)
	if err != nil {
		return fmt.Errorf("failed to read API file: %w", err)
	}

	var apis []APIRequest
	if err := json.Unmarshal(file, &apis); err != nil {
		return fmt.Errorf("failed to unmarshal APIs: %w", err)
	}

	for _, api := range apis {
		// Capture the api for the closure
		currentAPI := api
		apiMap[currentAPI.Request.Method] = func(variables map[string]interface{}) (response map[string]interface{}) {
			// Create the request body
			paramsStr := string(currentAPI.Request.Params)
			for key, value := range variables {
				placeholder := fmt.Sprintf("{{%s}}", key)
				// For string values, we need to make sure they are properly quoted in the JSON
				if strVal, ok := value.(string); ok {
					paramsStr = strings.ReplaceAll(paramsStr, placeholder, strVal)
				} else {
					// For non-string values, convert them to string
					paramsStr = strings.ReplaceAll(paramsStr, placeholder, fmt.Sprintf("%v", value))
				}
			}

			// Construct the full request payload
			requestPayload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  currentAPI.Request.Method,
				"params":  json.RawMessage(paramsStr),
			}
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
			}

			requestBody, err := json.Marshal(requestPayload)
			if err != nil {
				response["error"] = err.Error()
				return
			} else {
				resp, err := http.Post(cfg.URL, "application/json", bytes.NewBuffer(requestBody))
				if err != nil {
					response["error"] = err.Error()
					return
				}
				defer resp.Body.Close()
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					response["error"] = err.Error()
					return
				}

				var respMap map[string]interface{}
				if err := json.Unmarshal(respBody, &respMap); err != nil {
					response["error"] = err.Error()
					return
				}
				response = respMap
			}
			return
		}
	}
	return nil
}

// GetAPIExecutor returns the executor function for a given API method name.
func GetAPIExecutor(methodName string) (APIExecutor, bool) {
	executor, found := apiMap[methodName]
	return executor, found
}
