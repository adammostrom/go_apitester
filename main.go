package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	csvwriter "main/csv_writer"
	"main/models"
	"main/parser"
	"main/utils"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {

	config, err := readYamlFile("req_config.yaml")
	utils.CheckError(err, "failed to read yaml file")

	// Before anything, send a request to the target api and check that its alive.
	/* 	alive, err := checkAlive(config.Path)
	   	if err != nil || !alive {
	   		log.Fatal(err)
	   		return
	   	} */

	flattenedBody, err := flattenYamlBody(config.Body, []string{})
	utils.CheckError(err, "Failed to parse and flatten yaml body")

	requests := generateRequestHandler(flattenedBody)
	for _, req := range requests {
		fmt.Printf("req: %v\n", req)
	}
	// Here, generate setup for requests, like creating a request for each value in the values if its a "value" operator. Maybe generate a bunch of requests to send, store them in a slice, inform user of amount of requests, and if user wnats to see, prints them before sending them?

	generateRequest(config, requests)
}

// Send a GET request to url endpoint and expect a status code 200 back.
func checkAlive(url string) (bool, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, nil
	}
	return false, fmt.Errorf("Server not alive at url: %s\nStatus code: %d", url, resp.StatusCode)

}

// Generate the request bodies from flattened/parsed yaml bodies.
func generateRequestHandler(bodyFields []models.Field) []map[string]any {

	if bodyFields == nil {
		fmt.Println("Empty fields list, unable to generate any requests.")
		return nil
	}

	prepared_requests := parser.GenerateValueBodies(bodyFields)

	amount_requests := len(prepared_requests)

	for _, field := range bodyFields {

		switch field.Mode {

		case string(models.ModeValues):
			// Make a new request and store in the list of requests
			//generateValueBodies(bodyFields)
			continue
		case string(models.ModeRandom):
			// Generate a request for a subset of randomized numbers within the range

			// Generate random numbers, including mininum and maximum, preferably one random per permutated value, can be increased.

			randoms, err := parser.GenerateRandomPoints(field, amount_requests)
			if err != nil {
				log.Fatal("Failed to generate Random Numbers from Handler")
			}

			for i, rand := range randoms {
				parser.SetPath(field.Path, rand, prepared_requests[i])
			}
		case string(models.ModeList):
			// For each request, make a subset of the list
			lists, err := parser.GenerateSubLists(field, amount_requests)
			if err != nil {
				log.Fatal("Failed to generate lists array from request handler.")
			}
			for i, list := range lists {
				parser.SetPath(field.Path, list, prepared_requests[i])
			}

		}

	}
	return prepared_requests
}

// Generate request should only need the body and the method.
func generateRequest(config models.YamlConfig, requests []map[string]any) {

	responses := []models.Response{}

	//prepared_requests := [][]byte{}
	for _, request := range requests {
		jsonBody, err := json.Marshal(request)
		utils.CheckError(err, "could not marshal body to JSON")

		fmt.Println("Request:", string(jsonBody))

		req, err := http.NewRequest(
			config.Method,
			config.Path,
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		// Measure time for a request --> response

		start := time.Now()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		duration := time.Since(start)

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Fatal(err)
		}

		responses = append(responses, models.Response{Body: string(body), Resp: resp, Time: time.Duration(duration.Milliseconds())})
	}
	parseResponse(responses, config)

}

func parseResponse(responses []models.Response, config models.YamlConfig) {

	csvwriter.WriteToCsv(config, responses)
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
