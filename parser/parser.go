package parser

import (
	"fmt"
	"main/reporter"
	"math/rand/v2"
)

/*
TODO: Optimize
Given a request body and an amount, returns generated requests with the length of amount.
The generated requests are generated in a subcommand and returns the minimum length of the cartesian product of value/list values of the yaml file.
*/
func GenerateRequestBodies(bodyFields []Field, amount int) ([]map[string]any, error) {

	if amount < 0 {
		return nil, fmt.Errorf("Amount must be GEQ 0, provided was: %d", amount)
	}

	requests, err := ParseAndGenerateRequests(bodyFields)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []map[string]any{}, nil
	}
	// Default flag value is 0, so assumption is that 0 means: user never specified an amount
	if amount == 0 {
		amount = len(requests)
	}

	sizeRequests := len(requests)

	if sizeRequests > amount {
		return requests[0:amount], nil
	} else if sizeRequests == amount {
		return requests, nil
	} else if sizeRequests < amount {

		iterations := amount / sizeRequests

		for iterations > 0 {
			newRequests, err := ParseAndGenerateRequests(bodyFields)
			if err != nil {
				return nil, err
			}
			for _, newRequest := range newRequests {
				requests = append(requests, newRequest)
			}
			iterations--
		}
	}
	return requests[:amount], nil

}

// Generate the request bodies from flattened/parsed yaml bodies.
func ParseAndGenerateRequests(bodyFields []Field) ([]map[string]any, error) {

	if bodyFields == nil {
		reporter.VPrintf("Empty fields list, returning empty requests: %v\n", bodyFields)
		return []map[string]any{}, nil
	}

	bodies, err := GenerateValueBodies(bodyFields)
	if err != nil {
		return nil, err
	}

	for _, body := range bodies {
		for _, field := range bodyFields {

			switch field.Mode {

			case string(ModeValues):
				continue

			case string(ModeRandom):

				rand, err := GenerateRandomPoints(field)
				if err != nil {
					return nil, fmt.Errorf("Failed to generate Random Numbers from Handler")
				}
				setPath(field.Path, rand, body)

			case string(ModeSubset):

				setPath(field.Path, randomSubset(field.Values), body)

			case string(ModeList):
				setPath(field.Path, field.Values, body)

			}
		}
		reporter.VPrintf("prepared request body: %v\n", body)
	}
	return bodies, nil
}

// Take all Value operators in the fields, and make a cartesian product of them
func GenerateValueBodies(fields []Field) ([]map[string]any, error) {

	bodies := []map[string]any{
		{},
	}

	if fields == nil {
		return nil, fmt.Errorf("Fields empty, returning empty : %v", bodies)
	}

	for _, field := range fields {

		if field.Mode != string(ModeValues) {
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
	return bodies, nil
}

type RandomOp struct {
	Operator int
	Val      float64
}

const (
	MAX_OP = iota
	MIN_OP
)

// Refactor: Instead of making it return a list of random values, just have it generate one value each time?
func GenerateRandomPoints(randomField Field) (float64, error) {

	var min float64
	var max float64

	// Will never fail, pointless check, but anyways...
	if randomField.Mode != string(ModeRandom) {
		return 0, fmt.Errorf("Field mode not equal to list operator: %v", randomField.Mode)
	}

	if len(randomField.Values) < 1 {
		return 0, fmt.Errorf("Values empty, expected atleast a min and a max operator.")
	}

	// Should only be two items, min and max.
	for _, v := range randomField.Values {

		if value, ok := v.(RandomOp); !ok {
			return 0, fmt.Errorf("Expected randomOp from Values list of %v\n", v)
		} else {
			switch value.Operator {
			case MAX_OP:
				max = value.Val
			case MIN_OP:
				min = value.Val
			}
		}
	}
	if min > max {
		return 0, fmt.Errorf("Min cannot be larger than max: min:%d max%d\n", min, max)
	}

	return randomFloat(min, max), nil
}

func GenerateSubLists(field Field, multiplier int) ([][]any, error) {

	subLists := [][]any{}

	size := len(field.Values)

	// Will never fail, pointless check, but anyways...
	if field.Mode != string(ModeSubset) {
		return nil, fmt.Errorf("Field mode not equal to list operator: %v \n", field.Mode)
	}
	if size < 1 {
		// Not necessarily an error, sometimes endpoints requires tests for empty body fields.
		fmt.Println("Values list empty.")
		return nil, nil
	}

	for i := 0; i < multiplier; i++ {

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

func cloneMap(src map[string]any) map[string]any {

	dst := make(map[string]any)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Takes a slice of paths, and appends them into a recursive map
func setPath(path []string, val any, target map[string]any) {

	current := target

	for i, key := range path {

		if i == len(path)-1 {
			current[key] = val
			return
		}

		// Does this key already contain a map?
		next, ok := current[key].(map[string]any)

		if !ok {
			next = make(map[string]any)
			current[key] = next
		}

		current = next

	}
}

type Field struct {
	Path   []string
	Mode   string
	Values []any
}

type Mode string

// Add more eventually
const (
	ModeValues     Mode = "values"
	ModeList       Mode = "list"
	ModeRandom     Mode = "random"
	ModeSubset     Mode = "subset"
	ModeStochastic Mode = "stochastic"
	ModeStatic     Mode = "static"
)
