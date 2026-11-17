package mapstructure

import (
	"reflect"
	"testing"
)

func TestDecodeStructToTypedMapIgnoredFields(t *testing.T) {
	tests := []struct {
		name                 string
		input                any
		tagName              string
		ignoreUntaggedFields bool
	}{
		{
			name: "ignored field",
			input: struct {
				Name  string `mapstructure:"name"`
				Count int    `mapstructure:"-"`
			}{"kept", 42},
		},
		{
			name: "ignored field with options",
			input: struct {
				Name  string `mapstructure:"name"`
				Count int    `mapstructure:"-,omitempty"`
			}{"kept", 42},
		},
		{
			name: "untagged field",
			input: struct {
				Name  string `mapstructure:"name"`
				Count int
			}{"kept", 42},
			ignoreUntaggedFields: true,
		},
		{
			name: "fallback tag",
			input: struct {
				Name  string `json:"name"`
				Count int    `json:"-"`
			}{"kept", 42},
			tagName: "mapstructure,json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]string
			decoder, err := NewDecoder(&DecoderConfig{
				Result:               &result,
				TagName:              tt.tagName,
				IgnoreUntaggedFields: tt.ignoreUntaggedFields,
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := decoder.Decode(tt.input); err != nil {
				t.Fatal(err)
			}
			if want := map[string]string{"name": "kept"}; !reflect.DeepEqual(result, want) {
				t.Fatalf("got %#v, want %#v", result, want)
			}
		})
	}
}

func TestDecodeStructToTypedMapIncludedFields(t *testing.T) {
	input := struct {
		Name string `mapstructure:"-name"`
	}{"kept"}
	var result map[string]string
	if err := Decode(input, &result); err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"-name": "kept"}; !reflect.DeepEqual(result, want) {
		t.Fatalf("got %#v, want %#v", result, want)
	}
	if err := Decode(struct{ Count int }{42}, &result); err == nil {
		t.Fatal("expected an error for an included field with an incompatible type")
	}
}
