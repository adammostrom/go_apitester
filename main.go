package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlConfig struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Expect struct {
		Expect int `yaml:"status"`
	} `yaml:"expect"`
	Body map[string]any `yaml:"body"`
}

// Hardcoded test cases in the yaml file (can add more later)
type TestOption struct {
	Random RandomMinMax
}

type TestCase struct {
	Options TestOption
	Config  YamlConfig
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

// Global for now
var testOption TestOption

func main() {

	/*     body := []byte{
	       "product_name": "Salmon",
	       "amount": 200,
	       "unit": "GRAM",
	       "fields": ["CALORIES", "FAT", "SATURATED_FAT", "TRANS_FAT", "CHOLESTEROL", "CARBOHYDRATES", "SUGARS",
	       ,"ADDED_SUGARS", "SUCROSE", "GLUCOSE", "FRUCTOSE", "LACTOSE", "STARCH", "FIBER", "PROTEINS", "SALT"
	       ,"ADDED_SALT","SODIUM", "VITAMIN_C", "VITAMIN_B1","VITAMIN_B2","VITAMIN_PP","VITAMIN_B6","VITAMIN_B9"
	       ,"VITAMIN_B12", "POTASSIUM", "CALCIUM", "IRON", "MAGNESIUM", "ZINC"]
	   } */

	config, err := readYamlFile("req_config.yaml")
	checkError(err, "failed to read yaml file")

	var path []string

	flattened, err := flattenYamlBody(config.Body, path)
	checkError(err, "Failed to parse and flatten yaml body")

	// Here, generate setup for requests, like creating a request for each value in the values if its a "value" operator. Maybe generate a bunch of requests to send, store them in a slice, inform user of amount of requests, and if user wnats to see, prints them before sending them?
	fmt.Printf("flattened: %v\n", flattened)
	fmt.Printf("path: %v\n", path)

	fmt.Printf("config.Body: %v\n", flattened)

	generateRequest(config, testOption)
}

type Request struct {
	Name   string
	Path   string
	Method string
	Expect Expectation
	Body   []byte
}

func generateRequestHandler(config YamlConfig, bodyFields []Field) {

	// Take the url, method and body, and send it to generate request.
	// Derive here what should be in the body, which operator etc.

	// If Mode == values, then create one request per value of list, so json becomes "path:value"
	// We have to build requests basically

	requests := []Request{}

	if bodyFields == nil {
		fmt.Println("Empty fields list, unable to generate any requests.")
		return
	}

	for _, field := range bodyFields {

		switch field.Mode {

		case string(ModeValues):
			// Make a new request and store in the list of requests
			generateValueBodies(bodyFields, requests)

		case string(ModeRandom):
			// Generate a request for a subset of randomized numbers within the range

		case string(ModeList):
			// For each request, make a subset of the list

		}

	}
}

func cloneMap(src map[string]any) map[string]any {

	dst := make(map[string]any)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Take all Value operators in the fields, and make a cartesian product of them
func generateValueBodies(fields []Field, requests []Request) map[string]any {

	// Initially : [{}]
	body := []map[string]any{
		{},
	}

	values := map[string]any{}

	for _, field := range fields {

		if field.Mode == string(ModeValues) {
			for i := 0; i < len(field.Path)-1; i++ {
				values[field.Path[i]] = field.Path[i+1]
			}
		}

	}
}

// Generate request should only need the body and the method.
func generateRequest(config YamlConfig, options TestOption) {

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

func readYamlFile(path string) (YamlConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return YamlConfig{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var config YamlConfig

	if err := yaml.Unmarshal(contents, &config); err != nil {
		return YamlConfig{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return config, nil
}

/*

map[amount:map[random:map[max:200 min:0]] fields:map[values:CALORIES FAT SATURATED_FAT TRANS_FAT] product_name:map[values:salmon egg meatballs bread] unit:map[values:GRAM]]


*/

type Field struct {
	Path   []string
	Mode   string
	Values []any
}

type RandomOp struct {
	Operator string
	Val      any
}

func flattenYamlBody(body map[string]any, path []string) ([]Field, error) {
	// While key not equal to values: or random:, store key with value of next key

	var fields []Field

	for key, value := range body {

		switch key {

		case "values":
			// We expect that at "values" to find an array
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("values at %v must be an array", path)
			}
			fields = append(fields, Field{
				Path:   path,
				Mode:   "values",
				Values: values,
			})
		case "list":
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("List at %v must be an array", path)
			}
			fields = append(fields, Field{
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

			// TODO: Maybe redesign this, for now this is fine in order to pass ops into Values
			ops := []any{}

			ops = append(ops,
				RandomOp{Operator: "max", Val: max},
				RandomOp{Operator: "min", Val: min},
			)

			fields = append(fields, Field{
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

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}
