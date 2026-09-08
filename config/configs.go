package config

type RunConfig struct {
	Repeat   int
	Verbose  bool
	Output   string
	Requests int
	Rate     int
}

type YamlConfig struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Expect struct {
		// We expect atleast the yaml to contain status with a given key of "expect"
		Expect int `yaml:"status"`
	} `yaml:"expect"`
	Body map[string]any `yaml:"body"`
}
