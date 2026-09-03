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

func RunTests(runConfig config.RunConfig, file string) error {

	// Check file actually a yaml file
	if runConfig.Verbose {
		reporter.Verbose = true
	}

	config, err := reader.ReadYamlFile(file)
	if err != nil {
		return fmt.Errorf("Failed to read Yaml file: %s. error: %w", file, err)
	}

	flattenedBody, err := reader.FlattenYamlBody(config.Body, []string{})
	if err != nil {
		return fmt.Errorf("Failed to parse yaml body section: %s. error: %w ", config.Body, err)
	}

	requests := parser.GenerateRequests(flattenedBody, runConfig)

	for _, req := range requests {
		reporter.VPrintf("generated requests: %v\n", req)
	}
	// Here, generate setup for requests, like creating a request for each value in the values if its a "value" operator. Maybe generate a bunch of requests to send, store them in a slice, inform user of amount of requests, and if user wnats to see, prints them before sending them?

	responses, err := sendRequests(config, requests)
	if err != nil {
		return err
	}

	outputFile := determineOutputFile(config.Name, runConfig)
	if runConfig.Output != "" {
		reporter.WriteToCsv(config.Name, config.Method, config.Expect.Expect, responses, outputFile)
		return nil
	}

	/*
			testName,
		time.Now().Format(time.RFC3339),
		testMethod,
		response.Resp.Request.URL.Path,
		response.Result,
		response.Resp.Status,
		strconv.Itoa(response.Resp.StatusCode),
		strconv.Itoa(testExpectedMethod),
		response.Time.String(),
		response.Body,

	*/

	reporter.PrintResults(config, responses)
	return nil

}

func sendRequests(config config.YamlConfig, requests []map[string]any) ([]models.Response, error) {

	responses := []models.Response{}

	//prepared_requests := [][]byte{}
	for _, request := range requests {

		response, err := httpClient.SendRequest(config.Method, config.Path, request)
		if err != nil {
			//fmt.Printf("Request #%d failed -> %v\nwith error: %v", i, request, err)

			responses = append(responses, models.Response{Result: "FAIL", Error: err})
			continue
		}
		if response.Resp.StatusCode == config.Expect.Expect {
			response.Result = "PASS"
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
