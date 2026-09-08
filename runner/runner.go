package runner

import (
	"fmt"
	"main/config"
	httpClient "main/httpClient"
	"main/models"
	"main/parser"
	"main/reader"
	reporter "main/reporter"
	"os"
	"strings"
)

const FLAG_REPEAT_MAX = 100

func validateFlags(config config.RunConfig) error {

	if config.Repeat < 0 {
		return fmt.Errorf("Repeat cannot be negative. Provided: %d\n", config.Repeat)
	}
	if config.Repeat > FLAG_REPEAT_MAX {
		return fmt.Errorf("Repeat amount exceedes the maxium. Provided: %d, max: %d", config.Repeat, FLAG_REPEAT_MAX)
	}

	return nil
}

func RunTests(runConfig config.RunConfig, file string) error {

	if err := validateFlags(runConfig); err != nil {
		return err
	}

	if runConfig.Verbose {
		reporter.Verbose = true
	}

	yamlConfig, err := reader.ReadYamlFile(file)
	if err != nil {
		return fmt.Errorf("Failed to read Yaml file: %s. error: %w", file, err)
	}

	// TODO: Combine with readYamlFile and just return the valid yaml configuration
	if err = reader.ValidateYAML(yamlConfig); err != nil {
		return err
	}

	flattenedBody, err := reader.FlattenYamlBody(yamlConfig.Body, []string{})
	if err != nil {
		return fmt.Errorf("Failed to parse yaml body section: %s. error: %w ", yamlConfig.Body, err)
	}

	requestBodies, err := parser.GenerateRequestBodies(flattenedBody, runConfig.Requests)
	if err != nil {
		return err
	}

	for i, req := range requestBodies {
		reporter.VPrintf("generated request body: %d/%d, %v\n", i+1, len(requestBodies), req)
	}
	// Here, generate setup for requests, like creating a request for each value in the values if its a "value" operator. Maybe generate a bunch of requests to send, store them in a slice, inform user of amount of requests, and if user wnats to see, prints them before sending them?

	responses := []models.Response{}

	responses, err = sendRequests(yamlConfig, requestBodies, runConfig.Requests)
	if err != nil {
		return err
	}

	outputFile := determineOutputFile(yamlConfig.Name, runConfig)
	if runConfig.Output != "" {
		reporter.WriteToCsv(yamlConfig, responses, outputFile)
		return nil
	}

	reporter.PrintResults(yamlConfig, responses)
	return nil

}

func sendRequests(yamlConfig config.YamlConfig, requestBodies []map[string]any, amount int) ([]models.Response, error) {

	responses := []models.Response{}

	for i := 0; i < amount; i++ {
		request := map[string]any{}

		if len(requestBodies) != 0 {
			request = requestBodies[i]
		}

		response, err := httpClient.SendRequest(yamlConfig.Method, yamlConfig.Path, request)
		if err != nil {
			responses = append(responses, models.Response{Result: "FAIL", Error: err})
			continue
		}
		if response.Resp.StatusCode == yamlConfig.Expect.Expect {
			response.Result = "OK"
		} else {
			response.Result = "FAIL"
		}
		responses = append(responses, response)
	}
	return responses, nil

}

func determineOutputFile(testName string, runConfig config.RunConfig) string {

	if runConfig.Output != "" {
		return runConfig.Output
	}

	return cleanString(testName) + reporter.GetTimeStamp() + ".csv"
}

func cleanString(input string) string {
	return strings.ToLower(strings.ReplaceAll(input, " ", "_"))
}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}
