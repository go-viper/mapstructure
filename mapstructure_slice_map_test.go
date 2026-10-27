package mapstructure

import (
	"reflect"
	"strings"
	"testing"
)

func TestDecoder_WeakSliceOfMapsZeroFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		input                any
		want                 map[string]int
		nilWithoutZeroFields bool
	}{
		{
			name:  "disjoint keys",
			input: []map[string]any{{"a": 1}, {"b": 2}},
			want:  map[string]int{"a": 1, "b": 2},
		},
		{
			name:  "later duplicate keys win",
			input: []map[string]any{{"a": 1, "b": 2}, {"a": 3}},
			want:  map[string]int{"a": 3, "b": 2},
		},
		{
			name:  "interface slice",
			input: []any{map[string]any{"a": "1"}, map[string]any{"b": "2"}},
			want:  map[string]int{"a": 1, "b": 2},
		},
		{
			name:  "array",
			input: [2]map[string]int{{"a": 1}, {"b": 2}},
			want:  map[string]int{"a": 1, "b": 2},
		},
		{
			name:  "empty final map",
			input: []map[string]int{{"a": 1}, {}},
			want:  map[string]int{"a": 1},
		},
		{
			name:                 "nil final map",
			input:                []map[string]int{{"a": 1}, nil},
			want:                 map[string]int{"a": 1},
			nilWithoutZeroFields: true,
		},
		{
			name:  "nested slice",
			input: []any{[]map[string]int{{"a": 1}, {"b": 2}}, map[string]int{"c": 3}},
			want:  map[string]int{"a": 1, "b": 2, "c": 3},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, zeroFields := range []bool{false, true} {
				name := "merge existing"
				if zeroFields {
					name = "clear existing"
				}
				t.Run(name, func(t *testing.T) {
					got := map[string]int{"stale": 9}
					want := make(map[string]int, len(test.want)+1)
					for key, value := range test.want {
						want[key] = value
					}
					if !zeroFields {
						want["stale"] = 9
						if test.nilWithoutZeroFields {
							// A nil map still clears the result when merging existing values.
							want = nil
						}
					}
					config := &DecoderConfig{Result: &got, WeaklyTypedInput: true, ZeroFields: zeroFields}
					decoder, err := NewDecoder(config)
					if err != nil {
						t.Fatal(err)
					}
					if err := decoder.Decode(test.input); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("expected %#v, got %#v", want, got)
					}
					if config.ZeroFields != zeroFields {
						t.Fatal("decoder changed ZeroFields configuration")
					}
				})
			}
		})
	}
}

func TestDecoder_WeakSliceOfMapsZeroFieldsNestedValues(t *testing.T) {
	t.Parallel()

	got := map[string]map[string]int{"stale": {"value": 9}}
	config := &DecoderConfig{Result: &got, WeaklyTypedInput: true, ZeroFields: true}
	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatal(err)
	}
	input := []map[string]any{
		{"first": map[string]int{"old": 1}, "second": map[string]int{"keep": 2}},
		{"first": map[string]int{"new": 3}},
	}
	if err := decoder.Decode(input); err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]int{"first": {"new": 3}, "second": {"keep": 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}

	// Reusing the same decoder must still clear the previous result.
	if err := decoder.Decode([]map[string]any{{"third": map[string]int{"last": 4}}}); err != nil {
		t.Fatal(err)
	}
	want = map[string]map[string]int{"third": {"last": 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
	if !config.ZeroFields {
		t.Fatal("decoder changed ZeroFields configuration")
	}
}

func TestDecoder_WeakSliceOfMapsZeroFieldsError(t *testing.T) {
	t.Parallel()

	var got map[string]int
	config := &DecoderConfig{Result: &got, WeaklyTypedInput: true, ZeroFields: true}
	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatal(err)
	}
	err = decoder.Decode([]map[string]any{{"first": 1}, {"bad": "not an integer", "second": 2}})
	if err == nil || !strings.Contains(err.Error(), "[1][bad]") {
		t.Fatalf("expected an error for the second map's bad value, got %v", err)
	}
	want := map[string]int{"first": 1, "second": 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
	if !config.ZeroFields {
		t.Fatal("decoder changed ZeroFields configuration")
	}
}

func TestDecoder_WeakSliceOfMapsZeroFieldsHookTargets(t *testing.T) {
	t.Parallel()

	got := map[string]int{"stale": 9}
	var targets []map[string]int
	decoder, err := NewDecoder(&DecoderConfig{
		Result: &got, WeaklyTypedInput: true, ZeroFields: true,
		DecodeHook: func(from, to reflect.Value) (any, error) {
			if from.Kind() == reflect.Map {
				target := map[string]int{}
				for key, value := range to.Interface().(map[string]int) {
					target[key] = value
				}
				targets = append(targets, target)
			}
			return from.Interface(), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode([]map[string]int{{"first": 1}, {"second": 2}}); err != nil {
		t.Fatal(err)
	}
	want := []map[string]int{{"stale": 9}, {"first": 1}}
	if !reflect.DeepEqual(targets, want) {
		t.Fatalf("expected hook targets %#v, got %#v", want, targets)
	}
}
