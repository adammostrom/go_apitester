package models

type Request struct {
	Name   string
	Path   string
	Method string
	Expect Expectation
	Body   []byte
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

type RandomMinMax struct {
	FieldName string
	Max       any
	Min       any
}

type Expectation struct {
	StatusCode int
}

type Mode string

// Add more eventually
const (
	ModeValues     Mode = "values"
	ModeList       Mode = "list"
	ModeRandom     Mode = "random"
	ModeStochastic Mode = "stochastic"
	ModeStatic     Mode = "static"
)

type Field struct {
	Path   []string
	Mode   string
	Values []any
}
