package report

import (
	"fmt"
	"time"

	"perf-tester/pkg/worker"
)

// GenerateReport generates and prints a report of the performance test results
// @param results The results of the performance test
// @param totalDuration The total duration of the performance test
// @param totalCasesCompleted The total number of test cases completed
func GenerateReport(results []worker.Result, totalDuration time.Duration, totalCasesCompleted int) {
	totalRequests := len(results)
	successfulRequests := 0
	var totalRequestTime time.Duration

	for _, r := range results {
		if r.Error == nil && r.StatusCode == 200 {
			successfulRequests++
			totalRequestTime += r.Duration
		}
	}

	failed := totalRequests - successfulRequests
	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(successfulRequests) / float64(totalRequests) * 100
	}

	avgRequestTime := time.Duration(0)
	if successfulRequests > 0 {
		avgRequestTime = totalRequestTime / time.Duration(successfulRequests)
	}

	reqsPerSecond := 0.0
	if totalDuration.Seconds() > 0 {
		reqsPerSecond = float64(totalRequests) / totalDuration.Seconds()
	}

	fmt.Println("\n--- Performance Test Report ---")
	fmt.Printf("Total Duration: %s\n", totalDuration)
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful Requests: %d\n", successfulRequests)
	fmt.Printf("Failed Requests: %d\n", failed)
	fmt.Printf("Success Rate: %.2f%%\n", successRate)
	fmt.Printf("Average Request Time: %v\n", avgRequestTime)
	fmt.Printf("Requests Per Second: %.2f\n", reqsPerSecond)

	tps := 0.0
	if totalDuration.Seconds() > 0 {
		tps = float64(totalCasesCompleted) / totalDuration.Seconds()
	}
	fmt.Printf("Transactions Per Second (TPS): %.2f\n", tps)

	if failed > 0 {
		fmt.Println("\n--- Errors ---")
		for _, r := range results {
			if r.Error != nil {
				fmt.Printf("- %v\n", r.Error)
			}
		}
	}

	fmt.Println("-----------------------------")
}
