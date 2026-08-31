package parser

import (
	"fmt"
	"main/models"
	"math/rand/v2"
)

// Takes a slice of paths, and appends them into a recursive map
func SetPath(path []string, val any, target map[string]any) {

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

// Take all Value operators in the fields, and make a cartesian product of them
func GenerateValueBodies(fields []models.Field) []map[string]any {

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

				SetPath(field.Path, value, newBody)

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
func GenerateRandomPoints(randomField models.Field, multiplier int) ([]float64, error) {

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

func GenerateSubLists(field models.Field, multiplier int) ([][]any, error) {

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
