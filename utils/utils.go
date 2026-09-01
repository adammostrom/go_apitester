package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func CheckError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}

func ToFloat64(value any) (float64, error) {
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

func CleanString(input string) string {
	return strings.ToLower(strings.ReplaceAll(input, " ", "_"))
}

func GetTimeStamp() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02_15-04-05")
}
