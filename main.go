package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/models"
	"main/parser"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

// Global for now
var testOption models.TestOption

func main() {

	config, err := readYamlFile("req_config.yaml")
	checkError(err, "failed to read yaml file")

	var path []string

	flattened, err := flattenYamlBody(config.Body, path)
	checkError(err, "Failed to parse and flatten yaml body")

	requests := generateRequestHandler(config, flattened)
	for _, req := range requests {
		fmt.Printf("req: %v\n", req)
	}

	// Here, generate setup for requests, like creating a request for each value in the values if its a "value" operator. Maybe generate a bunch of requests to send, store them in a slice, inform user of amount of requests, and if user wnats to see, prints them before sending them?

	//generateRequest(config, testOption)
}

func generateRequestHandler(config models.YamlConfig, bodyFields []models.Field) []map[string]any {

	// Take the url, method and body, and send it to generate request.
	// Derive here what should be in the body, which operator etc.

	// If Mode == values, then create one request per value of list, so json becomes "path:value"
	// We have to build requests basically

	cartesianValues := parser.GenerateValueBodies(bodyFields)

	if bodyFields == nil {
		fmt.Println("Empty fields list, unable to generate any requests.")
		return nil
	}

	for _, field := range bodyFields {

		switch field.Mode {

		case string(models.ModeValues):
			// Make a new request and store in the list of requests
			//generateValueBodies(bodyFields)
			continue
		case string(models.ModeRandom):
			// Generate a request for a subset of randomized numbers within the range

			// Generate random numbers, including mininum and maximum, preferably one random per permutated value, can be increased.

			randoms, err := parser.GenerateRandomPoints(field, len(cartesianValues))
			if err != nil {
				log.Fatal("Failed to generate Random Numbers from Handler")
			}

			for i, rand := range randoms {
				parser.SetPath(field.Path, rand, cartesianValues[i])
			}
		case string(models.ModeList):
			// For each request, make a subset of the list
			lists, err := parser.GenerateSubLists(field, len(cartesianValues))
			if err != nil {
				log.Fatal("Failed to generate lists array from request handler.")
			}
			for i, list := range lists {
				parser.SetPath(field.Path, list, cartesianValues[i])
			}

		}

	}
	return cartesianValues
}

type TaggedValue struct {
	Value map[string]any
	Tag   int
}

func makeCartesianProduct(items []TaggedValue) [][]map[string]any {

	product := [][]map[string]any{}

	for i := 0; i < len(items); i++ {
		for j := i; j < len(items); j++ {
			current := []map[string]any{}
			// If they have different tags
			if items[i].Tag != items[j].Tag {
				current = append(current, items[i].Value, items[j].Value)
				product = append(product, current)
			}
		}
	}
	return product
}

// Generate request should only need the body and the method.
func generateRequest(config models.YamlConfig, options models.TestOption) {

	// For now, add the amount here, but make this into a function that randomizes, and then break it out so that it generates different ones each request
	config.Body[options.Random.FieldName] = options.Random.Max

	jsonBody, err := json.Marshal(config.Body)
	checkError(err, "could not marshal body to JSON\n")

	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(config.Method, config.Path, bodyReader)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	req_data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Request: ", string(req_data))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))

}

func readYamlFile(path string) (models.YamlConfig, error) {
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

func flattenYamlBody(body map[string]any, path []string) ([]models.Field, error) {
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

			nestedFields, err := flattenYamlBody(nested, append(path, key))
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

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}
