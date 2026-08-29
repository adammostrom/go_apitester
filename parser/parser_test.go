package parser

import (
	"reflect"
	"testing"
)

func TestSetPath(t *testing.T) {

	path := []string{
		"data",
		"parsed_data",
		"product_name",
	}

	got := setPath(path, "salmon")

	want := map[string]any{
		"data": map[string]any{
			"parsed_data": map[string]any{
				"product_name": "salmon",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"\nsetPath() = %#v \nwant %#v",
			got,
			want,
		)
	}

}
