package parser

import (
	"reflect"
	"testing"
)

func TestGenerateRequests(t *testing.T) {

	bodyFields := []Field{
		Field{
			Path:   []string{"data"},
			Mode:   "values",
			Values: []any{"salmon", "egg"},
		},
	}

	// amount of requests we want
	var amount = 3

	requests, err := ParseAndGenerateRequests(bodyFields)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	requests2, err := ParseAndGenerateRequests(bodyFields)

	requests = append(requests, requests2...)

	want := requests[:len(requests)-1]

	got, err := GenerateRequestBodies(bodyFields, amount)

	if len(got) != len(want) {
		t.Errorf("Lengths not equal:  got: %v, want: %v\n", got, want)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got: %v, and Want: %v not equal!\n", got, want)
	}
}

func TestParseAndGenerateRequests(t *testing.T) {

}

// SetPath Test //
func TestSetPath(t *testing.T) {

	path := []string{
		"data",
		"parsed_data",
		"product_name",
	}
	got := map[string]any{}

	setPath(path, "salmon", got)

	want := map[string]any{
		"data": map[string]any{
			"parsed_data": map[string]any{
				"product_name": "salmon",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"\nsetPath() = %#v \nwant = %#v",
			got,
			want,
		)
	}

}

// GenerateValueBodies Test //

func TestGenerateValueBodies(t *testing.T) {

	fields := []Field{
		{
			Path:   []string{"product_name"},
			Mode:   string(ModeValues),
			Values: []any{"salmon", "egg", "beans", "soymilk"},
		},
		{
			Path:   []string{"unit"},
			Mode:   string(ModeValues),
			Values: []any{"GRAMS", "CUPS"},
		},
	}

	// Might fail due to map iteration not being deterministic and hence map ordering of items can become random

	// Invariants:

	// Size of cartesian product = m x n
	generated, err := GenerateValueBodies(fields)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	lenwanted := 1
	for _, item := range fields {
		lenwanted *= len(item.Values)
	}

	if len(generated) != lenwanted {
		t.Fatalf("got %d bodies, want %d", len(generated), lenwanted)
	}

	// Contains data:

	for _, body := range generated {
		product := body["product_name"]
		unit := body["unit"]

		values := []any{
			"salmon", "egg", "beans", "soymilk",
		}
		valuesGrams := []any{
			"GRAMS", "CUPS",
		}

		// Hardcoded index, but fine for testing
		if !contains(values, product) {
			t.Errorf("unexpected product: %v", product)
		}

		if !contains(valuesGrams, unit) {
			t.Errorf("unexpected unit: %v", unit)
		}
	}
}

func TestGenerateRandomPoints(t *testing.T) {

	field := Field{
		Path: []string{"amount"},
		Mode: string(ModeRandom),
		Values: []any{
			RandomOp{Operator: MAX_OP, Val: 200},
			RandomOp{Operator: MIN_OP, Val: 0},
		},
	}

	numbers, err := GenerateRandomPoints(field)
	if err != nil {
		t.Errorf("Error from function")
	}

	if len(numbers) != 8 {
		t.Errorf("Unexpected length: %v", len(numbers))
	}

}

// Util
func contains(values []any, wanted any) bool {
	for _, value := range values {
		if reflect.DeepEqual(value, wanted) {
			return true
		}
	}
	return false
}
