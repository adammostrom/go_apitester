package config

type RunConfig struct {
	Repeat   int
	Verbose  bool
	Output   string
	Requests int
}

type YamlConfig struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Expect struct {
		Expect int `yaml:"status"`
	} `yaml:"expect"`
	Body map[string]any `yaml:"body"`
}
