package reporter

import (
	"encoding/csv"
	"log"
	"main/config"
	"main/models"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const OUTPUT_PATH = "result_outputs"

/*
TODO : Make the user also able to input their own file and/or filename to be printed at
apitest run tests.yaml --verbose
apitest run tests.yaml --parallel 4
apitest run tests.yaml --output results.csv

*/

// Alternative for now, bundle params into struct or send whole config
func createCSV(outputFileName string) (*os.File, *csv.Writer, error) {

	err := os.MkdirAll(OUTPUT_PATH, 0755)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Create(filepath.Join(OUTPUT_PATH, outputFileName+".csv"))
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

func WriteToCsv(config config.YamlConfig, responses []models.Response, fileName string) {

	file, writer, err := createCSV(fileName)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	defer writer.Flush()

	for _, response := range responses {
		records := []string{}

		if response.Error != nil {
			records = []string{
				config.Name,
				time.Now().Format(time.RFC3339),
				config.Method,
				config.Path,
				response.Result,
				"ERROR",
				response.Error.Error(),
			}
		} else {
			records = []string{
				config.Name,
				time.Now().Format(time.RFC3339),
				config.Method,
				config.Path,
				response.Result,
				response.Resp.Status,
				strconv.Itoa(response.Resp.StatusCode), // <--- BUG HERE WHEN RESPONSE IS NIL, SEGFAULT
				strconv.Itoa(config.Expect.Expect),
				response.Time.String(),
				response.Body,
				//response.Resp.Request.GetBody,
			}

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
