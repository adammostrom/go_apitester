package csvwriter

import (
	"encoding/csv"
	"main/utils"
	"os"
	"time"
)

const OUTPUT_PATH = "/data"

func getTimeStamp() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02-15:04:05")
}

func createCSV(config YamlConfig) (*os.File, error) {

	file, err := os.Create(testName + getTimeStamp() + ".csv")
	utils.CheckError(err, "Failed to create csv file")

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	colHeaders := make([]string, len(headers))

}
