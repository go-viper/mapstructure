package mapstructure

import (
	"strings"
	"testing"
)

// A field tagged with "-" is ignored by the decoder, so it must not be
// reported as unset: ErrorUnset used to fail with "has unset fields: -".
func TestDecode_IgnoredFieldIsNotReportedAsUnset(t *testing.T) {
	t.Parallel()

	type target struct {
		A string \`mapstructure:"a"\`
		B string \`mapstructure:"-"\`
	}

	var out target
	metadata := Metadata{}
	decoder, err := NewDecoder(&DecoderConfig{
		ErrorUnset: true,
		Metadata:   &metadata,
		Result:     &out,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := decoder.Decode(map[string]any{"a": "x"}); err != nil {
		t.Fatalf("an ignored field must not be reported as unset: %s", err)
	}
	if out.A != "x" {
		t.Fatalf("the tagged field was not decoded: %+v", out)
	}
	for _, name := range metadata.Unset {
		if strings.HasSuffix(name, "-") {
			t.Fatalf("metadata still lists the ignored field: %v", metadata.Unset)
		}
	}
}

// The control case has to keep failing: a real field that the input did not
// set is still an error.
func TestDecode_RealUnsetFieldStillFails(t *testing.T) {
	t.Parallel()

	type target struct {
		A string \`mapstructure:"a"\`
		C string \`mapstructure:"c"\`
	}

	var out target
	decoder, err := NewDecoder(&DecoderConfig{ErrorUnset: true, Result: &out})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = decoder.Decode(map[string]any{"a": "x"})
	if err == nil {
		t.Fatal("expected an unset field error")
	}
	if !strings.Contains(err.Error(), "c") {
		t.Fatalf("error should name the unset field: %s", err)
	}
}
