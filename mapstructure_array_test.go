package mapstructure

import (
	"reflect"
	"strings"
	"testing"
)

func TestDecodeArrayInputValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  any
		result any
		want   any
		weak   bool
		zero   bool
		err    string
	}{
		{
			name: "populated array with too many elements", input: []int{1, 2, 3},
			result: &[2]int{7, 8}, want: [2]int{7, 8},
			err: "expected source data to have length less or equal to 2, got 3",
		},
		{
			name: "array of slices with too many elements", input: [][]int{{1}, {2}},
			result: &[1][]int{}, want: [1][]int{},
			err: "expected source data to have length less or equal to 1, got 2",
		},
		{
			name: "empty array of slices with an element", input: [][]int{{1}},
			result: &[0][]int{}, want: [0][]int{},
			err: "expected source data to have length less or equal to 0, got 1",
		},
		{
			name: "populated array with scalar input", input: 3,
			result: &[2]int{7, 8}, want: [2]int{7, 8},
			err: "source data must be an array or slice, got int",
		},
		{
			name: "array of slices with scalar input", input: 3,
			result: &[1][]int{}, want: [1][]int{},
			err: "source data must be an array or slice, got int",
		},
		{
			name: "populated array with string input", input: "ab",
			result: &[2]int{7, 8}, want: [2]int{7, 8},
			err: "source data must be an array or slice, got string",
		},
		{
			name: "populated array with weak scalar input", input: 3,
			result: &[2]int{7, 8}, want: [2]int{3, 8}, weak: true,
		},
		{
			name: "array of slices with weak scalar input", input: 3,
			result: &[1][]int{}, want: [1][]int{{3}}, weak: true,
		},
		{
			name: "populated array with weak empty map input", input: map[string]int{},
			result: &[2]int{7, 8}, want: [2]int{}, weak: true,
		},
		{
			name: "populated array with weak nonempty map input", input: map[string]int{"a": 1},
			result: &[2]int{7, 8}, want: [2]int{7, 8}, weak: true,
			err: "source data must be an array or slice, got map",
		},
		{
			name: "zero array with too many elements", input: []int{1, 2, 3},
			result: &[2]int{}, want: [2]int{},
			err: "expected source data to have length less or equal to 2, got 3",
		},
		{
			name: "zero array with scalar input", input: 3,
			result: &[2]int{}, want: [2]int{},
			err: "source data must be an array or slice, got int",
		},
		{
			name: "short input preserves remaining elements", input: []int{1},
			result: &[2]int{7, 8}, want: [2]int{1, 8},
		},
		{
			name: "short input clears remaining elements with ZeroFields", input: []int{1},
			result: &[2]int{7, 8}, want: [2]int{1, 0}, zero: true,
		},
		{
			name: "short input preserves remaining slices", input: [][]int{{1}},
			result: &[2][]int{{7}, {8}}, want: [2][]int{{1}, {8}},
		},
		{
			name: "full array input", input: [2]int{1, 2},
			result: &[2]int{7, 8}, want: [2]int{1, 2},
		},
		{
			name: "nil input preserves destination", input: nil,
			result: &[2]int{7, 8}, want: [2]int{7, 8},
		},
		{
			name: "empty slice preserves destination", input: []int{},
			result: &[2]int{7, 8}, want: [2]int{7, 8},
		},
		{
			name: "empty slice clears destination with ZeroFields", input: []int{},
			result: &[2]int{7, 8}, want: [2]int{}, zero: true,
		},
		{
			name: "empty array", input: []int{},
			result: &[0]int{}, want: [0]int{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if err := recover(); err != nil {
					t.Errorf("Decode panicked: %v", err)
				}
			}()

			decoder, err := NewDecoder(&DecoderConfig{
				Result: tt.result, WeaklyTypedInput: tt.weak, ZeroFields: tt.zero,
			})
			if err != nil {
				t.Fatal(err)
			}
			err = decoder.Decode(tt.input)
			if tt.err == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Fatalf("expected error containing %q, got %v", tt.err, err)
			}
			if got := reflect.ValueOf(tt.result).Elem().Interface(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestDecodeArrayInputValidationNested(t *testing.T) {
	t.Parallel()

	type config struct {
		Values [2]int
	}
	result := config{Values: [2]int{7, 8}}
	decoder, err := NewDecoder(&DecoderConfig{Result: &result, RootName: "config"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("Decode panicked: %v", err)
		}
	}()
	err = decoder.Decode(map[string]any{"Values": []int{1, 2, 3}})
	want := "'config.Values' expected source data to have length less or equal to 2, got 3"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
	if result.Values != [2]int{7, 8} {
		t.Errorf("invalid input changed destination: %#v", result.Values)
	}
}
