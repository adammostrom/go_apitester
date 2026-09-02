package requestHandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/models"
	"main/parser"
	"main/printout"
	"net/http"
	"os"
	"time"
)

// Generate the request bodies from flattened/parsed yaml bodies.
func ParseAndGenerateRequests(bodyFields []models.Field) []map[string]any {

	if bodyFields == nil {
		fmt.Println("Empty fields list, unable to generate any requests.")
		return nil
	}

	prepared_requests := parser.GenerateValueBodies(bodyFields)

	// TODO: if the prepared_requests (permutated yaml values) are less than amount of requests, copy them randomly to fill out the list of requests. If the permutation is larger than the amount of requests, cut off the list, and save it presentable to the user like "requests NOT sent, to send a fully covered permutation, increase amount of requests to: "
	// Should be determined by the user input (amount of requests)
	amount_requests := len(prepared_requests)

	for _, field := range bodyFields {

		switch field.Mode {

		case string(models.ModeValues):
			continue

		// Generate random numbers, including mininum and maximum, preferably one random per permutated value, can be increased.
		case string(models.ModeRandom):

			randoms, err := parser.GenerateRandomPoints(field, amount_requests)
			if err != nil {
				// TODO: Make sure empty (not filled out min, max) returns in this crashing.
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
// TODO: Split up and refactor function
func SendRequests(config models.YamlConfig, requests []map[string]any) ([]models.Response, error) {

	responses := []models.Response{}

	//prepared_requests := [][]byte{}
	for _, request := range requests {
		jsonBody, err := json.Marshal(request)
		checkError(err, "could not marshal body to JSON")

		req, err := http.NewRequest(
			config.Method,
			config.Path,
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		// Measure time for a request --> response

		start := time.Now()
		printout.VPrintf("sending request: %v. timestamp: %v ", request, start.Format("15:04:03"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		duration := time.Since(start)
		printout.VPrintf(" - success. ms: %v\n", duration.Milliseconds())

		body, err := io.ReadAll(resp.Body)

		printout.VPrintf("response: %v\n ", resp)
		resp.Body.Close()

		if err != nil {
			return nil, err
		}
		result := "failed"
		if resStatusCode := config.Expect.Expect; resStatusCode == resp.StatusCode {
			result = "passed"
		}

		responses = append(responses, models.Response{
			Body:   string(body),
			Resp:   resp,
			Time:   time.Duration(duration.Milliseconds()),
			Result: result})
	}
	return responses, nil

}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}
