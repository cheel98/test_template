package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// APIRequest defines the structure for a single API request in api.json.
// Note that the params are kept as RawMessage to allow for flexible variable substitution.
type APIRequest struct {
	Request struct {
		Method string            `json:"method"`
		Params json.RawMessage `json:"params"`
	} `json:"request"`
	Response json.RawMessage `json:"response"`
}

// APIExecutor is the function type for executing an API call.
// It takes a map of variables and returns the response body.
type APIExecutor func(variables map[string]interface{}) json.RawMessage

var apiMap map[string]APIExecutor

// LoadApis loads all APIs from the given api.json file.
func LoadApis(apiFilePath string) error {
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
		apiMap[currentAPI.Request.Method] = func(variables map[string]interface{}) json.RawMessage {
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

			// In a real scenario, you would send this payload over HTTP.
			// For this test, we return the request payload so the worker can send it.
			payloadBytes, _ := json.Marshal(requestPayload)
			return payloadBytes
		}
	}

	return nil
}

// GetAPIExecutor returns the executor function for a given API method name.
func GetAPIExecutor(methodName string) (APIExecutor, bool) {
	executor, found := apiMap[methodName]
	return executor, found
}