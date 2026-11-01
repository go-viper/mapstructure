package mapstructure

import (
	"strings"
	"testing"
)

// Decoding a number that does not fit the destination type used to wrap it
// without any error: 300 decoded into an int8 silently became 44 and 256 into
// a uint8 became 0. A negative value going into a uint was already rejected, so
// the too-large value was the only silent case left.
func TestDecode_SilentIntegerTruncation(t *testing.T) {
	t.Parallel()

	type narrow struct {
		I8 int8
		U8 uint8
	}

	tests := []struct {
		name string
		in   map[string]any
		want string
	}{
		{"int into int8", map[string]any{"I8": 300}, "overflows int8"},
		{"int into uint8", map[string]any{"U8": 300}, "overflows uint8"},
		{"uint into uint8", map[string]any{"U8": uint(256)}, "overflows uint8"},
		{"int into int8 below the range", map[string]any{"I8": -200}, "overflows int8"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out narrow
			err := Decode(tc.in, &out)
			if err == nil {
				t.Fatalf("expected an overflow error, got %+v", out)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err.Error(), tc.want)
			}
		})
	}
}

func TestDecode_ValuesInsideTheRangeStillDecode(t *testing.T) {
	t.Parallel()

	type narrow struct {
		I8 int8
		U8 uint8
	}
	var out narrow
	if err := Decode(map[string]any{"I8": -128, "U8": 255}, &out); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if out.I8 != -128 || out.U8 != 255 {
		t.Fatalf("unexpected result: %+v", out)
	}
}

// WeaklyTypedInput keeps its documented leniency: only the strict path reports
// the overflow.
func TestDecode_OverflowAllowedWhenWeaklyTyped(t *testing.T) {
	t.Parallel()

	type narrow struct {
		U8 uint8
	}
	var out narrow
	decoder, err := NewDecoder(&DecoderConfig{WeaklyTypedInput: true, Result: &out})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := decoder.Decode(map[string]any{"U8": -1}); err != nil {
		t.Fatalf("weakly typed input should keep wrapping: %s", err)
	}
}
