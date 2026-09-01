package reader

import (
	"fmt"
	"main/models"
	"main/parser"
	"main/utils"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadYamlFile(path string) (models.YamlConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return models.YamlConfig{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var config models.YamlConfig

	if err := yaml.Unmarshal(contents, &config); err != nil {
		return models.YamlConfig{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return config, nil
}

/*
map[amount:map[random:map[max:200 min:0]] fields:map[values:CALORIES FAT SATURATED_FAT TRANS_FAT] product_name:map[values:salmon egg meatballs bread] unit:map[values:GRAM]]
*/

func FlattenYamlBody(body map[string]any, path []string) ([]models.Field, error) {
	// While key not equal to values: or random:, store key with value of next key

	var fields []models.Field

	for key, value := range body {

		switch key {

		case "values":
			// We expect that at "values" to find an array
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("values at %v must be an array", path)
			}
			fields = append(fields, models.Field{
				Path:   path,
				Mode:   "values",
				Values: values,
			})
		case "list":
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("List at %v must be an array", path)
			}
			fields = append(fields, models.Field{
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

			min_float, err := utils.ToFloat64(min)
			if err != nil {
				return nil, err
			}
			max_float, err := utils.ToFloat64(max)
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

			fields = append(fields, models.Field{
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
