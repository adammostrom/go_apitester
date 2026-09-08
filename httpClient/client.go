package requestHandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/models"
	"main/reporter"
	"net/http"
	"time"
)

// TODO: Split up and refactor function

// Sends a single request
func SendRequest(configMethod string, configPath string, request map[string]any) (models.Response, error) {

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return models.Response{}, fmt.Errorf("Failed to marshal json body: %w\n", err)
	}

	req, err := http.NewRequest(
		configMethod,
		configPath,
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return models.Response{}, fmt.Errorf("Failed to create new request: %w\n", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Measure time for a request --> response

	start := time.Now()
	reporter.VPrintf("sending request: %v. timestamp: %v \n", request, start.Format("15:04:03"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.Response{}, fmt.Errorf("request failed: %w\n", err)
	}

	duration := time.Since(start)
	reporter.VPrintf(" - success. ms: %v\n", duration.Milliseconds())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Response{}, fmt.Errorf("Failed to read response body, %w\n", err)
	}
	defer resp.Body.Close()
	reporter.VPrintf("response: %s\n ", body)

	response := models.Response{
		Body: string(body),
		Resp: resp,
		Time: time.Duration(duration.Milliseconds()),
	}

	return response, nil
}
