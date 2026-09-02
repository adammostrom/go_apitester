package runner

import (
	"log"
	csvwriter "main/csv_writer"
	"main/models"
	"main/printout"
	"main/reader"
	"main/requestHandler"
	"strings"
	"time"
)

func RunTests(runConfig models.RunConfig, file string) (string, error) {

	// Check file actually a yaml file
	if runConfig.Verbose {
		printout.Verbose = true
	}

	config, err := reader.ReadYamlFile(file)
	log.Fatal(err)

	flattenedBody, err := reader.FlattenYamlBody(config.Body, []string{})
	log.Fatal(err)

	requests := requestHandler.ParseAndGenerateRequests(flattenedBody)

	for _, req := range requests {
		printout.VPrintf("generated requests: %v\n", req)
	}
	// Here, generate setup for requests, like creating a request for each value in the values if its a "value" operator. Maybe generate a bunch of requests to send, store them in a slice, inform user of amount of requests, and if user wnats to see, prints them before sending them?

	responses, err := requestHandler.SendRequests(config, requests)
	if err != nil {
		log.Fatal(err)
	}

	outputFile := determineOutputFile(config, runConfig)

	csvwriter.WriteToCsv(config, responses, outputFile)
	return "success", nil

}

func determineOutputFile(config models.YamlConfig, runConfig models.RunConfig) string {
	if runConfig.Output != "" {
		return runConfig.Output
	}

	return cleanString(config.Name) + getTimeStamp() + ".csv"
}

func cleanString(input string) string {
	return strings.ToLower(strings.ReplaceAll(input, " ", "_"))
}

func getTimeStamp() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02_15-04-05")
}
