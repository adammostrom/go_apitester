package reader

import (
	"fmt"
	"main/config"
	"main/parser"
	"os"
	"unicode"

	"gopkg.in/yaml.v3"
)

var methods = [...]string{
	"GET",
	"POST",
	"DELETE",
	"UPDATE",
	"PATCH",
}

func ReadYamlFile(path string) (config.YamlConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return config.YamlConfig{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var config config.YamlConfig

	if err := yaml.Unmarshal(contents, &config); err != nil {
		return config, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return config, nil
}

/*
Enforce things that matter to correctness, e.g.:

name must exist
name must be non-empty
method must be a valid HTTP method
repeat >= 0
random.min <= random.max
required operators have the correct structure/type

Rule of thumb:

If violating it could break the program or make the configuration ambiguous, validate it. Otherwise, be permissive.
*/
func ValidateYAML(config config.YamlConfig) error {

	if config.Name == "" {
		return fmt.Errorf("Name section in yaml file not provided\n")
	}
	if config.Method == "" {
		return fmt.Errorf("Method section in yaml file not provided\n")
	}
	if config.Path == "" {
		return fmt.Errorf("Path URL not provided\n")
	}
	if !validateMethod(config.Method) {
		return fmt.Errorf("method provided is not valid: %v. Valid methods: %v\n", config.Method, methods)
	}
	if !validName(config.Name) {
		return fmt.Errorf("Illegal symbol in name section of yaml file: %s", config.Name)
	}
	if !validateExpected(config.Expect.Expect) {
		return fmt.Errorf("Expected status code of yaml file not accepted: %d", config.Expect.Expect)
	}

	return nil
}

func validName(name string) bool {
	for _, letter := range name {
		if unicode.IsSymbol(letter) {
			return false
		}
	}
	return true
}

func validateMethod(method string) bool {

	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func validateExpected(expected int) bool {

	var statusCodes = []int{200, 404, 506}

	for _, s := range statusCodes {
		if expected == s {
			return true
		}
	}
	return false
}

func validateSchemaStruct() {

	/*

		Legit Schema:
			name: nutrition-test
			method: POST
			path: /nutrition
			body:
			  product_name:
			    values: [salmon, egg]
	*/
}

/*
map[amount:map[random:map[max:200 min:0]] fields:map[values:CALORIES FAT SATURATED_FAT TRANS_FAT] product_name:map[values:salmon egg meatballs bread] unit:map[values:GRAM]]
*/

func FlattenYamlBody(yamlBody map[string]any, path []string) ([]parser.Field, error) {
	// While key not equal to values: or random:, store key with value of next key

	var fields []parser.Field

	for key, value := range yamlBody {

		switch key {

		case "values":
			// We expect that at "values" to find an array
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("values at %v must be an array", path)
			}
			fields = append(fields, parser.Field{
				Path:   path,
				Mode:   "values",
				Values: values,
			})
		case "subsets":
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("List at %v must be an array", path)
			}
			fields = append(fields, parser.Field{
				Path:   path,
				Mode:   "list",
				Values: values,
			})
		case "random":
			random, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("random at %v must have the structure: random: max:INT min: INT", path)
			}
			//TODO: Check for these (that they are filled out)
			min := random["min"]
			max := random["max"]

			min_float, err := toFloat64(min)
			if err != nil {
				return nil, err
			}
			max_float, err := toFloat64(max)
			if err != nil {
				return nil, err
			}
			//min_float, minOk := min.(float64)
			//max_float, maxOk := max.(float64)
			/* if !minOk || maxOk {
				return nil, fmt.Errorf("Random values provided at %v failed to be read as integers: min: %d, max: %d\n", key, min_float, max_float)
			} */

			// TODO: Maybe redesign this, for now this is fine in order to pass ops into Values
			ops := []any{}

			ops = append(ops,
				parser.RandomOp{Operator: parser.MAX_OP, Val: max_float},
				parser.RandomOp{Operator: parser.MIN_OP, Val: min_float},
			)

			fields = append(fields, parser.Field{
				Path:   path,
				Mode:   "random",
				Values: ops,
			})

			// Not an operator, hence another level in the yaml
		default:
			nested, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected object at %v", append(path, key))
			}

			nestedFields, err := FlattenYamlBody(nested, append(path, key))
			if err != nil {
				return nil, err
			}
			fields = append(fields, nestedFields...)
		}
	}

	return fields, nil
}

func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("expected number, got %T", value)
	}
}
