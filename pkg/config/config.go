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

// TestStep represents a single step in a test case.
type TestStep struct {
	ID       string                 `yaml:"id,omitempty"`
	Method   string                 // The API method name (extracted from the step key)
	Request  map[string]interface{} `yaml:"request,omitempty"`
	Response map[string]interface{} `yaml:"response,omitempty"`
	// 循环次数
	Loop int `yaml:"loop,omitempty"` // 为 -1 时 表示一直请求直到成功, 默认为1
	// 每次循环的间隔时间
	Interval int `yaml:"interval,omitempty"` // 单位为毫秒, 默认为1000
	MaxRetry int `yaml:"maxRetry,omitempty"` // 最大重试次数, 为 -1 时 表示不重试, 默认为10
}

// UnmarshalYAML implements custom YAML unmarshalling
func (t *TestStep) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Create a temporary struct to hold the unmarshalled data
	type TempStep struct {
		ID       string `yaml:"id,omitempty"`
		Method   string
		Request  map[string]interface{} `yaml:"request,omitempty"`
		Response map[string]interface{} `yaml:"response,omitempty"`
		Loop     int                    `yaml:"loop,omitempty"`
		Interval int                    `yaml:"interval,omitempty"`
		MaxRetry int                    `yaml:"maxRetry,omitempty"`
	}

	// Unmarshal into temporary struct
	var temp TempStep
	if err := unmarshal(&temp); err != nil {
		return err
	}

	// Copy values to the actual TestStep
	t.ID = temp.ID
	t.Method = temp.Method
	t.Request = temp.Request
	t.Response = temp.Response

	// Set default values
	t.Loop = temp.Loop
	if t.Loop == 0 {
		t.Loop = 1
	}

	t.Interval = temp.Interval
	if t.Interval == 0 {
		t.Interval = 1000 // Default interval is 1000ms
	}

	t.MaxRetry = temp.MaxRetry
	if t.MaxRetry == 0 {
		t.MaxRetry = 10 // Default max retry is 10
	}

	return nil
}

// TestCase represents a single test case in the test flow.
type TestCase struct {
	Name      string                 `yaml:"name"`
	Steps     map[string]TestStep    `yaml:"steps"`
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
