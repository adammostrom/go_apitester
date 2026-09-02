package reporter

import (
	"fmt"
	"main/config"
	"main/models"
	"time"
)

var Verbose bool

// Printer for verbose messages if verbose flag is true
func VPrintf(format string, args ...any) {
	if Verbose {
		fmt.Printf(format, args...)
	}
}

func PrintResults(config config.YamlConfig, responses []models.Response) {
	for _, response := range responses {

		if response.Error != nil {
			fmt.Printf(
				"[FAIL] %s > %s - %s - ERROR: %v\n",
				GetTimeStamp(),
				config.Name,
				response.Result,
				response.Error,
			)

			VPrintf(
				"%v",
				response.Time.Milliseconds(),
			)

			continue
		}

		fmt.Printf(
			"[INFO] %s > %s - %s - %d - %d\n",
			GetTimeStamp(),
			config.Name,
			response.Result,
			response.Resp.StatusCode,
			config.Expect.Expect,
		)

		VPrintf(
			"%v, %v",
			response.Resp.Body,
			response.Time.Milliseconds(),
		)
	}
}

func GetTimeStamp() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02_15-04-05")
}
