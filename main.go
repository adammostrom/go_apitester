package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlConfig struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Expect struct {
		Expect int `yaml:"status"`
	} `yaml:"expect"`
	Body map[string]any `yaml:"body"`
}

type TestCase struct {
	Name   string
	Method string
	Path   string
	Body   []byte
	Expect Expectation
}

type Expectation struct {
	StatusCode int
}

const TEST_URL = "http://localhost:8080/"

func main() {

	/*     body := []byte{
	       "product_name": "Salmon",
	       "amount": 200,
	       "unit": "GRAM",
	       "fields": ["CALORIES", "FAT", "SATURATED_FAT", "TRANS_FAT", "CHOLESTEROL", "CARBOHYDRATES", "SUGARS",
	       ,"ADDED_SUGARS", "SUCROSE", "GLUCOSE", "FRUCTOSE", "LACTOSE", "STARCH", "FIBER", "PROTEINS", "SALT"
	       ,"ADDED_SALT","SODIUM", "VITAMIN_C", "VITAMIN_B1","VITAMIN_B2","VITAMIN_PP","VITAMIN_B6","VITAMIN_B9"
	       ,"VITAMIN_B12", "POTASSIUM", "CALCIUM", "IRON", "MAGNESIUM", "ZINC"]
	   } */

	config, err := readYamlFile("req_config.yaml")
	checkError(err, "failed to read yaml file")

	jsonBody, err := json.Marshal(config.Body)
	checkError(err, "could not marshal body to JSON\n")

	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(config.Method, config.Path, bodyReader)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	req_data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Request: ", string(req_data))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))
}

func sendRequest(w http.ResponseWriter, r *http.Request, yaml YamlConfig) {

	// TODO: Create Response Model and make sure that the Expect is equals that of the request
	// var resp models.IssueResponse

	// TODO: Parse the yaml.body and turn it into json (I guess)
	//req, err := http.NewRequest(http.MethodPost, yaml.Path, bytes.NewBuffer(yaml.Body))

	// TODO: CHeck method, if POST, ELSE IF GET etc
	//resp, err := http.Post(yaml.Path, "application/json", yaml.Body)

	/* 	w.WriteHeader(http.StatusOK)
	   	json.NewEncoder(w).Encode(resp) */

}

func readYamlFile(path string) (YamlConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return YamlConfig{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var config YamlConfig

	if err := yaml.Unmarshal(contents, &config); err != nil {
		return YamlConfig{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return config, nil
}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s - %v \n", msg, err)
		os.Exit(1)
	}
}
