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

	flattened := flattenYamlBody(config.Body)
	config.Body = flattened

	fmt.Printf("config.Body: %v\n", flattened)

	generateRequest(config, testOption)
}

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

func flattenYamlBody(body map[string]any) map[string]any {
	// While key not equal to values: or random:, store key with value of next key

	result := map[string]any{}

	for k, v := range body {

		nested, ok := v.(map[string]any)

		if !ok {
			result[k] = v
			continue
		}

		if values, ok := nested["values"]; ok {
			result[k] = values
			continue
		}

		if random, ok := nested["random"].(map[string]any); ok {
			testOption.Random = RandomMinMax{
				FieldName: k,
				Min:       random["min"],
				Max:       random["max"],
			}
			fmt.Printf("testCase.Random: %v\n", testOption.Random)
			continue
		}

		result[k] = flattenYamlBody(nested)
	}
	return result

}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}
