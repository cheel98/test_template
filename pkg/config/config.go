package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration for the performance test
// @description Config holds the configuration for the performance test
// @field URL The target JSON-RPC endpoint URL

type Config struct {
	URL string `yaml:"url"`
}

// TestFlow represents the structure of the test_flow.yml file.
type TestFlow struct {
	Cases []TestCase `yaml:"cases"`
}

// TestCase represents a single test case in the test flow.
type TestCase struct {
	Name      string                 `yaml:"name"`
	Steps     []string               `yaml:"steps"`
	Loop      int                    `yaml:"loop"`
	Thread    int                    `yaml:"thread"`
	Variables map[string]interface{} `yaml:"variables"`
}

// Request represents a single request to be sent by a worker.
type Request struct {
	Method string
	Params []interface{}
}

// LoadConfigFromYAML loads the configuration from a YAML file.
func LoadConfigFromYAML(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadTestFlow loads the test flow from a YAML file.
func LoadTestFlow(path string) (*TestFlow, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var testFlow TestFlow
	if err := yaml.Unmarshal(file, &testFlow); err != nil {
		return nil, err
	}

	return &testFlow, nil
}
