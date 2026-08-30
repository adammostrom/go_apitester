package parser

import (
	"main/models"
	"reflect"
	"testing"
)

// SetPath Test //
func TestSetPath(t *testing.T) {

	path := []string{
		"data",
		"parsed_data",
		"product_name",
	}
	got := map[string]any{}

	SetPath(path, "salmon", got)

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

	fields := []models.Field{
		{
			Path:   []string{"product_name"},
			Mode:   string(models.ModeValues),
			Values: []any{"salmon", "egg", "beans", "soymilk"},
		},
		{
			Path:   []string{"unit"},
			Mode:   string(models.ModeValues),
			Values: []any{"GRAMS", "CUPS"},
		},
	}

	// Might fail due to map iteration not being deterministic and hence map ordering of items can become random

	// Invariants:

	// Size of cartesian product = m x n
	generated := GenerateValueBodies(fields)

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

	field := models.Field{
		Path: []string{"amount"},
		Mode: string(models.ModeRandom),
		Values: []any{
			RandomOp{Operator: MAX_OP, Val: 200},
			RandomOp{Operator: MIN_OP, Val: 0},
		},
	}

	numbers, err := GenerateRandomPoints(field, 8)
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
