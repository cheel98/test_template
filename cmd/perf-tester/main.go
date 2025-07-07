package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"perf-tester/pkg/api"
	"perf-tester/pkg/config"
	"perf-tester/pkg/report"
	"perf-tester/pkg/worker"
)

func main() {
	apiFile := flag.String("api", "api.json", "Path to the API definition file")
	testFlowFile := flag.String("flow", "test_flow.yml", "Path to the test flow file")
	configFile := flag.String("config", "config.yml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfigFromYAML(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := api.LoadApis(*apiFile); err != nil {
		log.Fatalf("Failed to load APIs: %v", err)
	}

	testFlow, err := config.LoadTestFlow(*testFlowFile)
	if err != nil {
		log.Fatalf("Failed to load test flow: %v", err)
	}

	resultsChan := make(chan worker.Result)
	var allResults []worker.Result
	var resultsMutex sync.Mutex
	startTime := time.Now()

	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for result := range resultsChan {
			resultsMutex.Lock()
			allResults = append(allResults, result)
			resultsMutex.Unlock()
		}
	}()

	totalCasesCompleted := 0
	for _, testCase := range testFlow.Cases {
		fmt.Printf("Running test case: %s\n", testCase.Name)
		requestsChan := make(chan config.Request, testCase.Loop*len(testCase.Steps))
		var wg sync.WaitGroup

		for i := 0; i < testCase.Thread; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				w := worker.NewWorker(workerID, cfg, resultsChan)
				w.Run(requestsChan)
			}(i)
		}

		for i := 0; i < testCase.Loop*testCase.Thread; i++ {
			for _, step := range testCase.Steps {
				executor, found := api.GetAPIExecutor(step)
				if !found {
					log.Printf("API method not found: %s", step)
					continue
				}
				rawResponse := executor(testCase.Variables)
				// todo rawResponse 不应该是requestData
				var requestData map[string]interface{}
				if err := json.Unmarshal(rawResponse, &requestData); err != nil {
					log.Printf("Error unmarshalling request data: %v", err)
					continue
				}

				requestsChan <- config.Request{
					Method: requestData["method"].(string),
					Params: requestData["params"].([]interface{}),
				}
			}
		}
		close(requestsChan)

		wg.Wait()
		totalCasesCompleted += testCase.Loop * testCase.Thread
	}

	close(resultsChan)
	collectorWg.Wait()

	totalDuration := time.Since(startTime)

	report.GenerateReport(allResults, totalDuration, totalCasesCompleted)

	fmt.Println("\nPerformance test finished.")
}
