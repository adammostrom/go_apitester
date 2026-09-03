package parser

import (
	"fmt"
	"log"
	"main/config"
	"main/models"
	"math/rand/v2"
)

func GenerateRequests(bodyFields []Field, runConfig config.RunConfig) []map[string]any {

	amount := runConfig.Requests

	if amount < 0 {
		return nil
	}

	requests := ParseAndGenerateRequests(bodyFields)

	if amount == 0 {
		amount = len(requests)
	}

	sizeRequests := len(requests)

	if sizeRequests > amount {
		return requests[0:amount]
	} else if sizeRequests == amount {
		return requests
	} else if sizeRequests < amount {

		iterations := amount / sizeRequests

		for iterations > 0 {
			newRequests := ParseAndGenerateRequests(bodyFields)
			for _, newRequest := range newRequests {
				requests = append(requests, newRequest)
			}
			iterations--
		}
	}
	return requests[:amount]

}

// Takes a slice of paths, and appends them into a recursive map
func setPath(path []string, val any, target map[string]any) {

	body := target

	current := body

	for i, key := range path {
		if i == len(path)-1 {
			current[key] = val
			break
		}
		next := map[string]any{}
		current[key] = next

		current = next

	}
	//return body
}

type Field struct {
	Path   []string
	Mode   string
	Values []any
}

// Generate the request bodies from flattened/parsed yaml bodies.
func ParseAndGenerateRequests(bodyFields []Field) []map[string]any {

	if bodyFields == nil {
		fmt.Println("Empty fields list, unable to generate any requests.")
		return nil
	}

	prepared_requests := GenerateValueBodies(bodyFields)

	// TODO: if the prepared_requests (permutated yaml values) are less than amount of requests, copy them randomly to fill out the list of requests. If the permutation is larger than the amount of requests, cut off the list, and save it presentable to the user like "requests NOT sent, to send a fully covered permutation, increase amount of requests to: "
	// Should be determined by the user input (amount of requests)
	amount_requests := len(prepared_requests)

	for _, field := range bodyFields {

		switch field.Mode {

		case string(models.ModeValues):
			continue

		// Generate random numbers, including mininum and maximum, preferably one random per permutated value, can be increased.
		case string(models.ModeRandom):

			randoms, err := GenerateRandomPoints(field, amount_requests)
			if err != nil {
				// TODO: Make sure empty (not filled out min, max) returns in this crashing.
				log.Fatal("Failed to generate Random Numbers from Handler")
			}

			for i, rand := range randoms {
				setPath(field.Path, rand, prepared_requests[i])
			}
		case string(models.ModeList):
			// For each request, make a subset of the list
			lists, err := GenerateSubLists(field, amount_requests)
			if err != nil {
				log.Fatal("Failed to generate lists array from request handler.")
			}
			for i, list := range lists {
				setPath(field.Path, list, prepared_requests[i])
			}

		}

	}
	return prepared_requests
}

// Take all Value operators in the fields, and make a cartesian product of them
func GenerateValueBodies(fields []Field) []map[string]any {

	// Initially : [{}]
	bodies := []map[string]any{
		{},
	}

	// {[product_name] values [salmon egg meatballs bread]}

	// The values found in the yaml file

	for _, field := range fields {

		if field.Mode != string(models.ModeValues) {
			continue
		}

		newBodies := []map[string]any{}

		for _, body := range bodies {

			for _, value := range field.Values {

				newBody := cloneMap(body)

				setPath(field.Path, value, newBody)

				newBodies = append(newBodies, newBody)
			}
		}
		bodies = newBodies
	}
	return bodies
}

func cloneMap(src map[string]any) map[string]any {

	dst := make(map[string]any)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

type RandomOp struct {
	Operator int
	Val      float64
}

const (
	MAX_OP = iota
	MIN_OP
)

// TODO: Currently only supporting integers.
func GenerateRandomPoints(randomField Field, multiplier int) ([]float64, error) {

	points := []float64{}

	var min float64
	var max float64

	// Will never fail, pointless check, but anyways...
	if randomField.Mode != string(models.ModeRandom) {
		return nil, fmt.Errorf("Field mode not equal to list operator: %v", randomField.Mode)
	}

	if len(randomField.Values) < 1 {
		return nil, fmt.Errorf("Values empty, expected atleast a min and a max operator.")
	}
	// Should only be two items, min and max.
	for _, v := range randomField.Values {

		if value, ok := v.(RandomOp); !ok {
			return nil, fmt.Errorf("Expected randomOp from Values list of %v\n", v)
		} else {
			switch value.Operator {
			case MAX_OP:
				max = value.Val
			case MIN_OP:
				min = value.Val
			}
		}
	}

	// TODO: Should always include the min and max, for edge case testing
	points = append(points, min)
	points = append(points, max)

	for i := 0; i < multiplier-2; i++ {

		random := randomFloat(min, max)

		if !Contains_float(points, random) {
			points = append(points, random)
		} else {
			// Iterate an extra time
			i--
		}
	}
	return points, nil
}

func GenerateSubLists(field Field, multiplier int) ([][]any, error) {

	subLists := [][]any{}

	size := len(field.Values)

	// Will never fail, pointless check, but anyways...
	if field.Mode != string(models.ModeList) {
		return nil, fmt.Errorf("Field mode not equal to list operator: %v \n", field.Mode)
	}
	if size < 1 {
		// Not necessarily an error, sometimes endpoints requires tests for empty body fields.
		fmt.Println("Values list empty.")
		return nil, nil
	}

	for i := 0; i < multiplier; i++ {
		/* 		list := []any{}

		   		upper_bound := randomInt(1, size)

		   		list = append(list, field.Values[:upper_bound]...)
		   		subLists = append(subLists, list) */
		subLists = append(subLists, randomSubset(field.Values))
	}

	return subLists, nil
}

func randomSubset(values []any) []any {
	subset := []any{}

	for _, value := range values {
		if rand.IntN(2) == 1 {
			subset = append(subset, value)
		}
	}

	if len(subset) == 0 {
		subset = append(subset, values[rand.IntN(len(values))])
	}

	return subset
}

// ----- Utils ------ //
func randomInt(min, max int) int {
	return min + rand.IntN(max)
}

func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// Take a list of floats, check if the value is in the list. O(n) expensive.
func Contains_float(values []float64, wanted float64) bool {
	for _, value := range values {
		if wanted == value {
			return true
		}
	}
	return false
}
