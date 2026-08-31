package csvwriter

import (
	"encoding/csv"
	"log"
	"main/models"
	"os"
	"strconv"
	"strings"
	"time"
)

const OUTPUT_PATH = "/data"

func getTimeStamp() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02_15-04-05")
}

// Alternative for now, bundle params into struct or send whole config
func createCSV(testName string) (*os.File, *csv.Writer, error) {
	file, err := os.Create(
		cleanString(testName) + getTimeStamp() + ".csv",
	)
	if err != nil {
		return nil, nil, err
	}

	writer := csv.NewWriter(file)

	// TODO: Do we really need url in the csv? Maybe add error and request number
	err = writer.Write([]string{
		"test_name",
		"timestamp",
		"method",
		"url",
		"result",
		"status",
		"statuscode_return",
		"statuscode_expect",
		"duration_ms",
		"request",
		"response",
	})
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, writer, nil
}

func WriteToCsv(config models.YamlConfig, responses []models.Response) {

	file, writer, err := createCSV(config.Name)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	defer writer.Flush()

	for _, response := range responses {

		result := "FAILED"

		if config.Expect.Expect == response.Resp.StatusCode {
			result = "PASS"
		}

		records := []string{
			config.Name,
			time.Now().Format(time.RFC3339),
			config.Method,
			response.Resp.Request.URL.Path,
			result,
			response.Resp.Status,
			strconv.Itoa(response.Resp.StatusCode),
			strconv.Itoa(config.Expect.Expect),
			response.Time.String(),
			response.Body,
			//response.Resp.Request.GetBody,
		}

		err := writer.Write(records)
		if err != nil {
			log.Fatal(err)
		}

	}

	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}

func cleanString(input string) string {
	return strings.ToLower(strings.ReplaceAll(input, " ", "_"))
}
