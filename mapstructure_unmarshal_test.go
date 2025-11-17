// This file tests the Unmarshaler interface implementation, which allows types
// to control their own decoding process (similar to json.Unmarshaler).
//
// Test Categories:
//
//  1. BASIC FUNCTIONALITY
//     [✓] TestUnmarshalerPointerReceiver - Pointer receiver implementation with string input
//     [✓] TestUnmarshalerValueReceiver - Value receiver (shows error from value receiver)
//     [✓] TestUnmarshalerDisabled - Unmarshaler NOT called when DisableUnmarshaler is true
//     [✓] TestUnmarshalerDisabled - Backwards compatibility with struct field decoding
//     [✓] TestUnmarshalerAlias - Works with aliased types (type CustomTrimmedString string)
//
//  2. ERROR HANDLING
//     [✓] TestUnmarshalerError - Errors from UnmarshalMapstructure are properly propagated
//     [ ] Multiple errors in nested structures with Unmarshalers
//     [ ] Partial decode failures in slices/maps
//
//  3. INPUT TYPES TO UNMARSHALER
//     [✓] String input - Most tests use this (TestUnmarshalerPointerReceiver, etc.)
//     [✓] Nil input - TestUnmarshalerSlicePointers, TestUnmarshalerMapPointerValues
//
//  4. COLLECTIONS & NESTED STRUCTURES
//     [✓] TestUnmarshalerStruct - Struct with fields that implement Unmarshaler (ptr & value)
//     [✓] TestUnmarshalerStructToStruct - Decode from struct to struct with Unmarshaler fields
//     [✓] TestUnmarshalerStructToStructWithOmitzero - Omitzero affects what Unmarshaler sees
//     [✓] TestUnmarshalerStructToStructTypeMismatch - Struct-to-struct type mismatch errors
//     [✓] TestUnmarshalerSlice - Slice of value types ([]CustomTypePtr)
//     [✓] TestUnmarshalerSlicePointers - Slice of pointer types ([]*CustomTypePtr)
//     [✓] TestUnmarshalerMapValues - Map with value types (map[string]CustomTypePtr)
//     [✓] TestUnmarshalerMapPointerValues - Map with pointer types (map[string]*CustomTypePtr)
//     [✓] TestUnmarshalerArray - Fixed-size array of value types ([3]CustomTypePtr)
//     [✓] TestUnmarshalerArrayPointers - Fixed-size array of pointer types ([2]*CustomTypePtr)
//     [✓] TestUnmarshalerEmbedded - Plain embedded fields without squash (value & pointer)
//
//  5. INTERACTION WITH OTHER FEATURES
//     [✓] TestUnmarshalerWithDecodeHook - DecodeHook transforms input before Unmarshaler
//     [✓] TestUnmarshalerWithMetadata - Metadata tracking (Keys/Unused/Unset)
//     [✓] TestUnmarshalerWithErrorUnused - ErrorUnused detects unused top-level keys
//     [✓] TestUnmarshalerWithWeaklyTypedInput - Unmarshaler takes precedence over WeaklyTypedInput
//     [✓] TestUnmarshalerWithZeroFields - ZeroFields doesn't affect Unmarshaler types
//
//  6. POINTER & ADDRESSABILITY SCENARIOS
//     [✓] TestUnmarshalerStruct - Pointer field (*CustomTypePtr) and nil pointer handling
//     [✓] TestUnmarshalerDoublePointer - Double and triple pointer scenarios (fields)
//     [✓] TestUnmarshalerTargetPointer - Target is pointer (var result *CustomTypePtr)
//     [✓] TestUnmarshalerTargetDoublePointer - Target is double pointer (var result **CustomTypePtr)
//
//  7. STRUCT TAGS & OPTIONS
//     [✓] TestUnmarshalerWithTag - Works with mapstructure tag (value & pointer fields)
//     [✓] TestUnmarshalerWithSquash - Works with squash tag (value & pointer embeddings)
package mapstructure

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func decode(t *testing.T, conf *DecoderConfig, input any) error {
	decoder, err := NewDecoder(conf)
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	return decoder.Decode(input)
}

// decodeSuccess test that decoding succeeds
func decodeSuccess(t *testing.T, conf *DecoderConfig, input any) {
	if err := decode(t, conf, input); err != nil {
		t.Fatalf("Expected decoding to succeed: %v", err)
	}
}

// decodeError test that decoding fails with an expected error message
func decodeFail(t *testing.T, conf *DecoderConfig, input any, expected string) {
	if err := decode(t, conf, input); err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Expected decoding to fail with error containing '%s', got: %v", expected, err)
	}
}

// CustomTypePtr implements Unmarshaler interface as a pointer receiver. It is a simple example type
// which parses strings of the form `TAG-ID` into a struct.
type CustomTypePtr struct {
	Tag string
	ID  int
}

func (c *CustomTypePtr) UnmarshalMapstructure(input any) error {
	v, ok := input.(string)
	if !ok {
		return fmt.Errorf("expected string input for CustomTypePtr, got %T", input)
	}
	parts := strings.SplitN(v, "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for CustomTypePtr: expected 'TAG-ID', got '%s'", v)
	}
	c.Tag = parts[0]
	var err error
	c.ID, err = strconv.Atoi(parts[1])
	return err
}

func (c *CustomTypePtr) AssertDecoded(t *testing.T, tag string, id int) {
	if c == nil {
		t.Fatal("Expected CustomTypePtr to be non-nil")
	}
	if c.Tag != tag {
		t.Errorf("Expected Tag '%s', got '%s'", tag, c.Tag)
	}
	if c.ID != id {
		t.Errorf("Expected ID %d, got %d", id, c.ID)
	}
}

// CustomTypeValue implements an Unmarshaler as a value. This is an anti-pattern.
type CustomTypeValue struct{}

func (c CustomTypeValue) UnmarshalMapstructure(input any) error {
	// This won't actually modify the original value since it's a value receiver
	// but we can test that it's called
	return fmt.Errorf("value receiver should not be used")
}

// MapSumType accepts a map and stores the sum of its integer values
type MapSumType struct {
	Sum   int
	Count int
}

func (m *MapSumType) UnmarshalMapstructure(input any) error {
	inputMap, ok := input.(map[string]any)
	if !ok {
		return fmt.Errorf("MapSumType expects map[string]any, got %T", input)
	}

	m.Sum = 0
	m.Count = 0
	for key, value := range inputMap {
		// Try to convert value to int
		var intVal int
		switch v := value.(type) {
		case int:
			intVal = v
		case int64:
			intVal = int(v)
		case float64:
			intVal = int(v)
		default:
			return fmt.Errorf("MapSumType: cannot convert value for key '%s' (type %T) to int", key, value)
		}
		m.Sum += intVal
		m.Count++
	}
	return nil
}

// Test unmarshaler implemented as pointer receiver
func TestUnmarshalerPointerReceiver(t *testing.T) {
	var result CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		"TEST-123",
	)
	result.AssertDecoded(t, "TEST", 123)
}

// Test unmarshaler implemented as value receiver. We expect to decode fine, but should be
// considered anti-pattern.
func TestUnmarshalerValueReceiver(t *testing.T) {
	var result CustomTypeValue
	decodeFail(t,
		&DecoderConfig{Result: &result},
		"TEST-123",
		"value receiver should not be used",
	)
}

// Test that unmarshaling works when the target itself is a pointer to a type with Unmarshaler
// This tests: var result *CustomTypePtr (not var result CustomTypePtr)
func TestUnmarshalerTargetPointer(t *testing.T) {
	var result *CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		"TARGET-789",
	)
	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}
	result.AssertDecoded(t, "TARGET", 789)
}

// Test that unmarshaling works when the target is a pointer to pointer
func TestUnmarshalerTargetDoublePointer(t *testing.T) {
	var result **CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		"DOUBLE-999",
	)
	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}
	if *result == nil {
		t.Fatal("Expected *result to be non-nil")
	}
	(**result).AssertDecoded(t, "DOUBLE", 999)
}

// Test that unmarshaling fails when unmarshaler returns an error
func TestUnmarshalerError(t *testing.T) {
	var result CustomTypePtr
	decodeFail(t,
		&DecoderConfig{Result: &result},
		"TEST123",
		"invalid format for CustomTypePtr",
	)
}

// Test unmarshaling behavior with DisableUnmarshaler true
func TestUnmarshalerDisabled(t *testing.T) {
	// Test 1: String input fails when DisableUnmarshaler is true
	var result1 CustomTypePtr
	decodeFail(t,
		&DecoderConfig{
			Result:             &result1,
			DisableUnmarshaler: true,
		},
		"TEST-123",
		"expected a map or struct",
	)

	// Test 2: Backwards compatibility - can still decode by providing actual struct fields
	// When DisableUnmarshaler is true, decoder falls back to normal struct decoding
	var result2 CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{
			Result:             &result2,
			DisableUnmarshaler: true,
		},
		map[string]any{
			"Tag": "COMPAT",
			"ID":  999,
		},
	)
	result2.AssertDecoded(t, "COMPAT", 999)

	// Test 3: Same backwards compatibility works in nested structs
	type Container struct {
		Value CustomTypePtr
		Name  string
	}
	var result3 Container
	decodeSuccess(t,
		&DecoderConfig{
			Result:             &result3,
			DisableUnmarshaler: true,
		},
		map[string]any{
			"Value": map[string]any{
				"Tag": "NESTED",
				"ID":  777,
			},
			"Name": "container",
		},
	)
	result3.Value.AssertDecoded(t, "NESTED", 777)
	if result3.Name != "container" {
		t.Errorf("Expected Name 'container', got '%s'", result3.Name)
	}
}

// Tests that unmarshaling works with aliased native types
type CustomTrimmedString string

func (c *CustomTrimmedString) UnmarshalMapstructure(input any) error {
	v, ok := input.(string)
	if !ok {
		return fmt.Errorf("expected string input for CustomTrimmedString, got %T", input)
	}
	*c = CustomTrimmedString(strings.TrimSpace(v))
	return nil
}

// Test that unmarshaling works with aliased types
func TestUnmarshalerAlias(t *testing.T) {
	var result CustomTrimmedString
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		"  abcd  ",
	)
	if result != "abcd" {
		t.Errorf("Expected 'abcd', got '%s'", result)
	}
}

// Test that unmarshaling works within structs containing both pointer and value fields
// Also tests nil pointer handling
func TestUnmarshalerStruct(t *testing.T) {
	type Container struct {
		PtrField   *CustomTypePtr
		ValueField CustomTypePtr
		NilPtr     *CustomTypePtr
	}

	var result Container
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]any{
			"PtrField":   "ALPHA-100",
			"ValueField": "BETA-200",
			"NilPtr":     nil, // Explicit nil value should remain nil
		},
	)
	result.PtrField.AssertDecoded(t, "ALPHA", 100)
	result.ValueField.AssertDecoded(t, "BETA", 200)
	if result.NilPtr != nil {
		t.Errorf("Expected NilPtr to remain nil, got %+v", result.NilPtr)
	}
}

// Test decoding from struct to struct where target has Unmarshaler fields
// This tests the struct-to-map-to-struct flow
func TestUnmarshalerStructToStruct(t *testing.T) {
	// Source struct with plain fields
	type Source struct {
		Code string
		Item string
		Name string
	}

	// Target struct where fields implement Unmarshaler
	type Target struct {
		Code CustomTypePtr
		Item *CustomTrimmedString
		Name string
	}

	source := Source{
		Code: "GAMMA-300",
		Item: "  trimmed  ",
		Name: "test",
	}

	var result Target
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		source,
	)

	// Code should be unmarshaled via CustomTypePtr.UnmarshalMapstructure
	result.Code.AssertDecoded(t, "GAMMA", 300)

	// Item should be unmarshaled via CustomTrimmedString.UnmarshalMapstructure
	if result.Item == nil {
		t.Fatalf("Expected Item to be set, got nil")
	}
	if *result.Item != "trimmed" {
		t.Errorf("Expected Item 'trimmed', got '%s'", *result.Item)
	}

	// Name is a plain string, should be copied directly
	if result.Name != "test" {
		t.Errorf("Expected Name 'test', got '%s'", result.Name)
	}
}

// Test struct-to-struct decoding with omitzero tag affecting what Unmarshaler receives
// The omitzero tag on source struct fields causes zero-valued fields to be omitted
// from the intermediary map, so Unmarshaler never sees them
func TestUnmarshalerStructToStructWithOmitzero(t *testing.T) {
	// Source struct with omitzero tags
	type Source struct {
		Code   string `mapstructure:"code,omitzero"`
		Item   string `mapstructure:"item,omitzero"`
		Status string `mapstructure:"status,omitzero"`
		Name   string `mapstructure:"name"`
	}

	// Target struct where fields implement Unmarshaler
	type Target struct {
		Code   *CustomTypePtr
		Item   *CustomTrimmedString
		Status *CustomTrimmedString
		Name   string
	}

	// Test 1: All fields have values - all should be decoded
	source1 := Source{
		Code:   "ALPHA-100",
		Item:   "  item1  ",
		Status: "  active  ",
		Name:   "test1",
	}

	var result1 Target
	decodeSuccess(t,
		&DecoderConfig{Result: &result1},
		source1,
	)

	if result1.Code == nil {
		t.Fatalf("Expected Code to be set")
	}
	result1.Code.AssertDecoded(t, "ALPHA", 100)

	if result1.Item == nil {
		t.Fatalf("Expected Item to be set")
	}
	if *result1.Item != "item1" {
		t.Errorf("Expected Item 'item1', got '%s'", *result1.Item)
	}

	if result1.Status == nil {
		t.Fatalf("Expected Status to be set")
	}
	if *result1.Status != "active" {
		t.Errorf("Expected Status 'active', got '%s'", *result1.Status)
	}

	if result1.Name != "test1" {
		t.Errorf("Expected Name 'test1', got '%s'", result1.Name)
	}

	// Test 2: Some fields have zero values with omitzero tag
	// Zero-valued fields with omitzero are omitted from intermediary map
	// So Unmarshaler never receives them, fields remain nil
	source2 := Source{
		Code:   "BETA-200",
		Item:   "", // Zero value with omitzero - will be omitted
		Status: "", // Zero value with omitzero - will be omitted
		Name:   "test2",
	}

	var result2 Target
	decodeSuccess(t,
		&DecoderConfig{Result: &result2},
		source2,
	)

	if result2.Code == nil {
		t.Fatalf("Expected Code to be set")
	}
	result2.Code.AssertDecoded(t, "BETA", 200)

	// Item and Status were omitted due to omitzero, so pointers should be nil
	if result2.Item != nil {
		t.Errorf("Expected Item to be nil (omitted by omitzero), got '%s'", *result2.Item)
	}
	if result2.Status != nil {
		t.Errorf("Expected Status to be nil (omitted by omitzero), got '%s'", *result2.Status)
	}

	if result2.Name != "test2" {
		t.Errorf("Expected Name 'test2', got '%s'", result2.Name)
	}

	// Test 3: Pre-existing values in target with omitzero omissions
	// Verify that omitted fields don't cause Unmarshaler to be called
	result3 := Target{
		Code:   &CustomTypePtr{Tag: "EXISTING", ID: 999},
		Item:   new(CustomTrimmedString),
		Status: new(CustomTrimmedString),
		Name:   "existing",
	}
	*result3.Item = "existing-item"
	*result3.Status = "existing-status"

	source3 := Source{
		Code:   "GAMMA-300",
		Item:   "", // Zero value with omitzero - omitted, won't overwrite
		Status: "", // Zero value with omitzero - omitted, won't overwrite
		Name:   "test3",
	}

	decodeSuccess(t,
		&DecoderConfig{Result: &result3},
		source3,
	)

	// Code was in map, so it gets decoded
	result3.Code.AssertDecoded(t, "GAMMA", 300)

	// Item and Status were omitted, so pre-existing values remain
	if result3.Item == nil || *result3.Item != "existing-item" {
		t.Errorf("Expected Item to keep existing value 'existing-item', got %v", result3.Item)
	}
	if result3.Status == nil || *result3.Status != "existing-status" {
		t.Errorf("Expected Status to keep existing value 'existing-status', got %v", result3.Status)
	}

	if result3.Name != "test3" {
		t.Errorf("Expected Name 'test3', got '%s'", result3.Name)
	}

	// Test 4: All omitzero fields are zero including Code
	// Verify that all Unmarshaler fields preserve pre-existing values when omitted
	result4 := Target{
		Code:   &CustomTypePtr{Tag: "EXISTING-CODE", ID: 888},
		Item:   new(CustomTrimmedString),
		Status: new(CustomTrimmedString),
		Name:   "existing4",
	}
	*result4.Item = "existing-item4"
	*result4.Status = "existing-status4"

	source4 := Source{
		Code:   "", // Zero value with omitzero - omitted
		Item:   "", // Zero value with omitzero - omitted
		Status: "", // Zero value with omitzero - omitted
		Name:   "test4",
	}

	decodeSuccess(t,
		&DecoderConfig{Result: &result4},
		source4,
	)

	// All omitzero fields were zero, so all should preserve pre-existing values
	if result4.Code == nil {
		t.Fatalf("Expected Code to preserve pre-existing value, got nil")
	}
	result4.Code.AssertDecoded(t, "EXISTING-CODE", 888)

	if result4.Item == nil || *result4.Item != "existing-item4" {
		t.Errorf("Expected Item to keep existing value 'existing-item4', got %v", result4.Item)
	}
	if result4.Status == nil || *result4.Status != "existing-status4" {
		t.Errorf("Expected Status to keep existing value 'existing-status4', got %v", result4.Status)
	}

	// Name doesn't have omitzero, so it's always in the map and gets decoded
	if result4.Name != "test4" {
		t.Errorf("Expected Name 'test4', got '%s'", result4.Name)
	}
}

// Test that struct-to-struct with Unmarshaler expecting wrong type fails appropriately
// When source has nested struct and target Unmarshaler expects a different type
func TestUnmarshalerStructToStructTypeMismatch(t *testing.T) {
	type SourceNested struct {
		Tag string
		ID  int
	}
	type Source struct {
		Product SourceNested
		Name    string
	}

	type Target struct {
		Product CustomTypePtr // Expects string, will get map
		Name    string
	}

	source := Source{
		Product: SourceNested{
			Tag: "DELTA",
			ID:  400,
		},
		Name: "product-name",
	}

	var result Target
	// This should fail because CustomTypePtr expects string but gets map
	decodeFail(t,
		&DecoderConfig{Result: &result},
		source,
		"expected string input for CustomTypePtr",
	)
}

// Test omitzero with map field and Unmarshaler that accepts maps
// Tests whether zero-value map (nil) with omitzero prevents Unmarshaler from being called
func TestUnmarshalerMapWithOmitzero(t *testing.T) {
	type Source struct {
		Data   map[string]any `mapstructure:"data,omitzero"`
		Scores map[string]any `mapstructure:"scores,omitzero"`
		Name   string         `mapstructure:"name"`
	}

	type Target struct {
		Data   *MapSumType
		Scores *MapSumType
		Name   string
	}

	// Test 1: Non-zero maps - Unmarshaler should be called and sum values
	source1 := Source{
		Data: map[string]any{
			"a": 10,
			"b": 20,
			"c": 30,
		},
		Scores: map[string]any{
			"x": 5,
			"y": 15,
		},
		Name: "test1",
	}

	var result1 Target
	decodeSuccess(t,
		&DecoderConfig{Result: &result1},
		source1,
	)

	if result1.Data == nil {
		t.Fatalf("Expected Data to be set")
	}
	if result1.Data.Sum != 60 {
		t.Errorf("Expected Data.Sum 60 (10+20+30), got %d", result1.Data.Sum)
	}
	if result1.Data.Count != 3 {
		t.Errorf("Expected Data.Count 3, got %d", result1.Data.Count)
	}

	if result1.Scores == nil {
		t.Fatalf("Expected Scores to be set")
	}
	if result1.Scores.Sum != 20 {
		t.Errorf("Expected Scores.Sum 20 (5+15), got %d", result1.Scores.Sum)
	}
	if result1.Scores.Count != 2 {
		t.Errorf("Expected Scores.Count 2, got %d", result1.Scores.Count)
	}

	// Test 2: Zero-value (nil) maps with omitzero - Unmarshaler should NOT be called
	// Fields should remain nil because they're omitted from the intermediary map
	source2 := Source{
		Data:   nil, // Zero value (nil map) with omitzero - will be omitted
		Scores: nil, // Zero value (nil map) with omitzero - will be omitted
		Name:   "test2",
	}

	var result2 Target
	decodeSuccess(t,
		&DecoderConfig{Result: &result2},
		source2,
	)

	// Data and Scores were omitted due to omitzero, so they should remain nil
	// The Unmarshaler was NEVER called
	if result2.Data != nil {
		t.Errorf("Expected Data to be nil (omitted by omitzero), got Sum=%d Count=%d", result2.Data.Sum, result2.Data.Count)
	}
	if result2.Scores != nil {
		t.Errorf("Expected Scores to be nil (omitted by omitzero), got Sum=%d Count=%d", result2.Scores.Sum, result2.Scores.Count)
	}

	if result2.Name != "test2" {
		t.Errorf("Expected Name 'test2', got '%s'", result2.Name)
	}

	// Test 3: Empty (non-nil) maps with omitzero
	// Important: omitzero uses reflect.IsZero(), which returns false for empty (non-nil) maps
	// So empty maps ARE included in the intermediary map, and Unmarshaler IS called
	source3 := Source{
		Data:   map[string]any{}, // Empty map (not nil) - NOT zero, so not omitted
		Scores: map[string]any{}, // Empty map (not nil) - NOT zero, so not omitted
		Name:   "test3",
	}

	var result3 Target
	decodeSuccess(t,
		&DecoderConfig{Result: &result3},
		source3,
	)

	// Empty maps are NOT zero values (reflect.IsZero returns false for them)
	// So they're included in the intermediary map and Unmarshaler IS called
	if result3.Data == nil {
		t.Fatalf("Expected Data to be set (empty map is not zero)")
	}
	if result3.Data.Sum != 0 {
		t.Errorf("Expected Data.Sum 0 (empty map), got %d", result3.Data.Sum)
	}
	if result3.Data.Count != 0 {
		t.Errorf("Expected Data.Count 0 (empty map), got %d", result3.Data.Count)
	}

	if result3.Scores == nil {
		t.Fatalf("Expected Scores to be set (empty map is not zero)")
	}
	if result3.Scores.Sum != 0 {
		t.Errorf("Expected Scores.Sum 0 (empty map), got %d", result3.Scores.Sum)
	}
	if result3.Scores.Count != 0 {
		t.Errorf("Expected Scores.Count 0 (empty map), got %d", result3.Scores.Count)
	}

	// Test 4: Pre-existing values with zero-value maps and omitzero
	// Verify that omitted maps preserve pre-existing Unmarshaler values
	result4 := Target{
		Data:   &MapSumType{Sum: 999, Count: 10},
		Scores: &MapSumType{Sum: 888, Count: 20},
		Name:   "existing",
	}

	source4 := Source{
		Data:   nil, // Zero value with omitzero - omitted
		Scores: nil, // Zero value with omitzero - omitted
		Name:   "test4",
	}

	decodeSuccess(t,
		&DecoderConfig{Result: &result4},
		source4,
	)

	// Pre-existing values should be preserved since fields were omitted
	if result4.Data == nil {
		t.Fatalf("Expected Data to preserve pre-existing value, got nil")
	}
	if result4.Data.Sum != 999 || result4.Data.Count != 10 {
		t.Errorf("Expected Data to preserve Sum=999 Count=10, got Sum=%d Count=%d", result4.Data.Sum, result4.Data.Count)
	}

	if result4.Scores == nil {
		t.Fatalf("Expected Scores to preserve pre-existing value, got nil")
	}
	if result4.Scores.Sum != 888 || result4.Scores.Count != 20 {
		t.Errorf("Expected Scores to preserve Sum=888 Count=20, got Sum=%d Count=%d", result4.Scores.Sum, result4.Scores.Count)
	}

	if result4.Name != "test4" {
		t.Errorf("Expected Name 'test4', got '%s'", result4.Name)
	}
}

// Test that omitzero on a struct field affects whether Unmarshaler is called.
// When a source struct has omitzero on a nested struct field, and that struct is zero,
// the field is omitted and the Unmarshaler is NOT called.
func TestUnmarshalerNestedStructWithOmitzero(t *testing.T) {
	type Numbers struct {
		A int
		B int
		C int
	}

	type Source struct {
		Data Numbers `mapstructure:"data,omitzero"`
		Name string  `mapstructure:"name"`
	}

	type Target struct {
		Data *MapSumType
		Name string
	}

	// Test 1: All fields zero with omitzero - struct is zero, field omitted, Unmarshaler NOT called
	source1 := Source{
		Data: Numbers{A: 0, B: 0, C: 0}, // Zero struct
		Name: "test1",
	}
	var result1 Target
	decodeSuccess(t, &DecoderConfig{Result: &result1}, source1)

	// Unmarshaler should NOT be called, Data should remain nil
	if result1.Data != nil {
		t.Errorf("Expected Data to be nil (omitted due to zero struct), got %+v", result1.Data)
	}
	if result1.Name != "test1" {
		t.Errorf("Expected Name 'test1', got '%s'", result1.Name)
	}

	// Test 2: Some fields non-zero - struct is not zero, Unmarshaler IS called
	source2 := Source{
		Data: Numbers{A: 5, B: 0, C: 10},
		Name: "test2",
	}
	var result2 Target
	decodeSuccess(t, &DecoderConfig{Result: &result2}, source2)

	// Unmarshaler should be called and receive the struct converted to map
	if result2.Data == nil {
		t.Fatal("Expected Data to be initialized, got nil")
	}
	expectedSum := 5 + 0 + 10
	if result2.Data.Sum != expectedSum {
		t.Errorf("Expected Sum=%d (A + B + C), got Sum=%d", expectedSum, result2.Data.Sum)
	}
	if result2.Data.Count != 3 {
		t.Errorf("Expected Count=3 (all fields included), got Count=%d", result2.Data.Count)
	}

	// Test 3: Zero struct with omitzero, but target has pre-existing value
	source3 := Source{
		Data: Numbers{A: 0, B: 0, C: 0},
		Name: "test3",
	}
	result3 := Target{
		Data: &MapSumType{Sum: 999, Count: 10},
		Name: "old",
	}
	decodeSuccess(t, &DecoderConfig{Result: &result3}, source3)

	// Field omitted, pre-existing value preserved
	if result3.Data == nil {
		t.Fatal("Expected Data to preserve pre-existing value, got nil")
	}
	if result3.Data.Sum != 999 {
		t.Errorf("Expected Sum=999 (preserved), got Sum=%d", result3.Data.Sum)
	}
	if result3.Data.Count != 10 {
		t.Errorf("Expected Count=10 (preserved), got Count=%d", result3.Data.Count)
	}
	if result3.Name != "test3" {
		t.Errorf("Expected Name 'test3', got '%s'", result3.Name)
	}

	// Test 4: Without omitzero, zero struct IS passed to Unmarshaler
	type SourceNoOmit struct {
		Data Numbers `mapstructure:"data"` // No omitzero
		Name string  `mapstructure:"name"`
	}
	source4 := SourceNoOmit{
		Data: Numbers{A: 0, B: 0, C: 0},
		Name: "test4",
	}
	var result4 Target
	decodeSuccess(t, &DecoderConfig{Result: &result4}, source4)

	// Unmarshaler IS called even though struct is zero (no omitzero)
	if result4.Data == nil {
		t.Fatal("Expected Data to be initialized (no omitzero), got nil")
	}
	if result4.Data.Sum != 0 {
		t.Errorf("Expected Sum=0 (all zeros), got Sum=%d", result4.Data.Sum)
	}
	if result4.Data.Count != 3 {
		t.Errorf("Expected Count=3 (all fields included), got Count=%d", result4.Data.Count)
	}
}

// Test that unmarshaling works with slices of types that implement Unmarshaler
func TestUnmarshalerSlice(t *testing.T) {
	var result []CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		[]string{
			"FIRST-1",
			"SECOND-2",
			"THIRD-3",
		},
	)
	if len(result) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(result))
	}
	result[0].AssertDecoded(t, "FIRST", 1)
	result[1].AssertDecoded(t, "SECOND", 2)
	result[2].AssertDecoded(t, "THIRD", 3)
}

// Test that unmarshaling works with slices of pointer types that implement Unmarshaler
// including nil values
func TestUnmarshalerSlicePointers(t *testing.T) {
	var result []*CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		[]any{
			"ALPHA-10",
			nil,
			"BETA-20",
		},
	)
	if len(result) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(result))
	}
	result[0].AssertDecoded(t, "ALPHA", 10)
	if result[1] != nil {
		t.Errorf("Expected nil element at index 1, got %+v", result[1])
	}
	result[2].AssertDecoded(t, "BETA", 20)
}

// Test that unmarshaling works with maps where values implement Unmarshaler
func TestUnmarshalerMapValues(t *testing.T) {
	var result map[string]CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]string{
			"key1": "FOO-100",
			"key2": "BAR-200",
			"key3": "BAZ-300",
		},
	)
	if len(result) != 3 {
		t.Fatalf("Expected 3 map entries, got %d", len(result))
	}
	foo := result["key1"]
	foo.AssertDecoded(t, "FOO", 100)
	bar := result["key2"]
	bar.AssertDecoded(t, "BAR", 200)
	baz := result["key3"]
	baz.AssertDecoded(t, "BAZ", 300)
}

// Test that unmarshaling works with maps where pointer values implement Unmarshaler
// including nil values
func TestUnmarshalerMapPointerValues(t *testing.T) {
	var result map[string]*CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]any{
			"x":   "DELTA-111",
			"nil": nil,
			"y":   "GAMMA-222",
		},
	)
	if len(result) != 3 {
		t.Fatalf("Expected 3 map entries, got %d", len(result))
	}
	result["x"].AssertDecoded(t, "DELTA", 111)
	if result["nil"] != nil {
		t.Errorf("Expected nil value for key 'nil', got %+v", result["nil"])
	}
	result["y"].AssertDecoded(t, "GAMMA", 222)
}

// Test that unmarshaling works with fixed-size arrays of types that implement Unmarshaler
func TestUnmarshalerArray(t *testing.T) {
	var result [3]CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		[]string{
			"ONE-1",
			"TWO-2",
			"THREE-3",
		},
	)
	result[0].AssertDecoded(t, "ONE", 1)
	result[1].AssertDecoded(t, "TWO", 2)
	result[2].AssertDecoded(t, "THREE", 3)
}

// Test that unmarshaling works with fixed-size arrays of pointer types that implement Unmarshaler
func TestUnmarshalerArrayPointers(t *testing.T) {
	var result [2]*CustomTypePtr
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		[]string{
			"ZETA-99",
			"OMEGA-88",
		},
	)
	result[0].AssertDecoded(t, "ZETA", 99)
	result[1].AssertDecoded(t, "OMEGA", 88)
}

// Test that unmarshaling works with mapstructure struct tags for field renaming
func TestUnmarshalerWithTag(t *testing.T) {
	type Container struct {
		Code     CustomTypePtr  `mapstructure:"product_code"`
		Item     *CustomTypePtr `mapstructure:"custom_item"`
		Metadata string         `mapstructure:"meta"`
	}

	var result Container
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]any{
			"product_code": "PROD-123",
			"custom_item":  "ITEM-999",
			"meta":         "some metadata",
		},
	)
	result.Code.AssertDecoded(t, "PROD", 123)
	result.Item.AssertDecoded(t, "ITEM", 999)
	if result.Metadata != "some metadata" {
		t.Errorf("Expected Metadata 'some metadata', got '%s'", result.Metadata)
	}
}

// Test that unmarshaling works with squashed embedded structs containing Unmarshaler fields
// This tests fields with Unmarshaler within squashed structs (both value and pointer embeddings)
func TestUnmarshalerWithSquash(t *testing.T) {
	type EmbeddedValue struct {
		Code CustomTypePtr
		ID   int
	}
	type EmbeddedPtr struct {
		Item *CustomTypePtr
		Val  string
	}
	type Container struct {
		EmbeddedValue `mapstructure:",squash"`
		*EmbeddedPtr  `mapstructure:",squash"`
		Name          string
	}

	var result Container
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]any{
			"Code": "SQUASH-100",
			"ID":   42,
			"Item": "PTR-999",
			"Val":  "test",
			"Name": "container",
		},
	)
	result.EmbeddedValue.Code.AssertDecoded(t, "SQUASH", 100)
	if result.EmbeddedValue.ID != 42 {
		t.Errorf("Expected ID 42, got %d", result.EmbeddedValue.ID)
	}
	if result.EmbeddedPtr == nil {
		t.Fatal("Expected EmbeddedPtr to be non-nil")
	}
	result.EmbeddedPtr.Item.AssertDecoded(t, "PTR", 999)
	if result.EmbeddedPtr.Val != "test" {
		t.Errorf("Expected Val 'test', got '%s'", result.EmbeddedPtr.Val)
	}
	if result.Name != "container" {
		t.Errorf("Expected Name 'container', got '%s'", result.Name)
	}
}

// Test that unmarshaling works with plain embedded fields (without squash) that implement Unmarshaler
// Without squash, embedded fields are accessed by their type name as a key
func TestUnmarshalerEmbedded(t *testing.T) {
	type Container struct {
		CustomTypePtr        // Value embedding
		*CustomTrimmedString // Pointer embedding
		Name                 string
	}

	var result Container
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]any{
			"CustomTypePtr":       "EMBED-500",
			"CustomTrimmedString": "  spaces  ",
			"Name":                "container",
		},
	)
	result.AssertDecoded(t, "EMBED", 500)
	if result.CustomTrimmedString == nil {
		t.Fatal("Expected CustomTrimmedString to be non-nil")
	}
	if *result.CustomTrimmedString != "spaces" {
		t.Errorf("Expected CustomTrimmedString 'spaces', got '%s'", *result.CustomTrimmedString)
	}
	if result.Name != "container" {
		t.Errorf("Expected Name 'container', got '%s'", result.Name)
	}
}

// Test that unmarshaling works with double pointers (pointer to pointer) that implement Unmarshaler
// Like json.Unmarshal, we should allocate both pointer levels and call UnmarshalMapstructure
func TestUnmarshalerDoublePointer(t *testing.T) {
	type Container struct {
		DoublePtr **CustomTypePtr
		TriplePtr ***CustomTrimmedString
		Name      string
	}

	var result Container
	decodeSuccess(t,
		&DecoderConfig{Result: &result},
		map[string]any{
			"DoublePtr": "DOUBLE-123",
			"TriplePtr": "  triple  ",
			"Name":      "test",
		},
	)

	// Check double pointer
	if result.DoublePtr == nil {
		t.Fatal("Expected DoublePtr to be non-nil")
	}
	if *result.DoublePtr == nil {
		t.Fatal("Expected *DoublePtr to be non-nil")
	}
	(**result.DoublePtr).AssertDecoded(t, "DOUBLE", 123)

	// Check triple pointer
	if result.TriplePtr == nil {
		t.Fatal("Expected TriplePtr to be non-nil")
	}
	if *result.TriplePtr == nil {
		t.Fatal("Expected *TriplePtr to be non-nil")
	}
	if **result.TriplePtr == nil {
		t.Fatal("Expected **TriplePtr to be non-nil")
	}
	if ***result.TriplePtr != "triple" {
		t.Errorf("Expected 'triple', got '%s'", ***result.TriplePtr)
	}

	if result.Name != "test" {
		t.Errorf("Expected Name 'test', got '%s'", result.Name)
	}
}

// Test that DecodeHook can transform input before Unmarshaler sees it
// This verifies the order: DecodeHook → Unmarshaler → default
func TestUnmarshalerWithDecodeHook(t *testing.T) {
	// DecodeHook that transforms int to string format expected by CustomTypePtr
	intToCustomString := func(from, to reflect.Type, data any) (any, error) {
		// Check if we're converting from int to CustomTypePtr (struct)
		if from.Kind() == reflect.Int && to == reflect.TypeOf(CustomTypePtr{}) {
			if fromInt, ok := data.(int); ok {
				// Transform int to string format that CustomTypePtr expects
				return fmt.Sprintf("HOOK-%d", fromInt), nil
			}
		}
		return data, nil
	}

	type Container struct {
		Value CustomTypePtr
		Count int
	}

	var result Container
	decodeSuccess(t,
		&DecoderConfig{
			Result:     &result,
			DecodeHook: intToCustomString,
		},
		map[string]any{
			"Value": 123, // int input, should be transformed by hook before Unmarshaler sees it
			"Count": 456,
		},
	)

	// Verify the Unmarshaler received the transformed string "HOOK-123"
	// which should be parsed as Tag="HOOK", ID=123
	result.Value.AssertDecoded(t, "HOOK", 123)
	if result.Count != 456 {
		t.Errorf("Expected Count 456, got %d", result.Count)
	}
}

// Test that Metadata tracking works correctly with Unmarshaler
func TestUnmarshalerWithMetadata(t *testing.T) {
	type Container struct {
		Value      CustomTypePtr
		Name       string
		UnsetField int
	}

	var result Container
	var md Metadata

	decodeSuccess(t,
		&DecoderConfig{
			Result:   &result,
			Metadata: &md,
		},
		map[string]any{
			"Value":  "META-555",
			"Name":   "test",
			"Unused": "ignored",
		},
	)
	result.Value.AssertDecoded(t, "META", 555)
	if result.Name != "test" {
		t.Errorf("Expected Name 'test', got '%s'", result.Name)
	}

	// Check Keys (successfully decoded fields)
	expectedKeys := []string{"Name", "Value"}
	if len(md.Keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d: %v", len(expectedKeys), len(md.Keys), md.Keys)
	}
	for _, key := range expectedKeys {
		found := false
		for _, k := range md.Keys {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key '%s' in metadata Keys, got: %v", key, md.Keys)
		}
	}

	// Check Unused (keys in input but not matched)
	if len(md.Unused) != 1 || md.Unused[0] != "Unused" {
		t.Errorf("Expected Unused to be ['Unused'], got: %v", md.Unused)
	}

	// Check Unset (fields in struct but not in input)
	if len(md.Unset) != 1 || md.Unset[0] != "UnsetField" {
		t.Errorf("Expected Unset to be ['UnsetField'], got: %v", md.Unset)
	}
}

// PartialUnmarshaler is a type that only uses some keys from its input
// This tests that unused keys within the Unmarshaler's input don't cause issues
type PartialUnmarshaler struct {
	UsedValue string
	// Note: We don't store UnusedValue
}

func (p *PartialUnmarshaler) UnmarshalMapstructure(input any) error {
	m, ok := input.(map[string]any)
	if !ok {
		return fmt.Errorf("PartialUnmarshaler expects map input, got %T", input)
	}
	// Only use the "used" key, ignore "unused" key
	if v, ok := m["used"].(string); ok {
		p.UsedValue = v
	}
	// Deliberately don't touch m["unused"] - it's consumed by this Unmarshaler
	// but not actually stored anywhere
	return nil
}

// Test that ErrorUnused works correctly with Unmarshaler
// Unused keys at the top level should cause an error, but keys consumed by
// Unmarshaler (even if not all are used internally) should not
func TestUnmarshalerWithErrorUnused(t *testing.T) {
	type Container struct {
		Partial PartialUnmarshaler
		Name    string
	}

	// Test 1: Unused keys at the struct level should error
	var result1 Container
	decodeFail(t,
		&DecoderConfig{
			Result:      &result1,
			ErrorUnused: true,
		},
		map[string]any{
			"Partial": map[string]any{
				"used":   "value1",
				"unused": "ignored-by-unmarshaler",
			},
			"Name":          "test",
			"ExtraTopLevel": "should-cause-error", // This should cause ErrorUnused
		},
		"ExtraTopLevel",
	)

	// Test 2: No unused keys at struct level - should succeed
	// Even though Unmarshaler doesn't use all keys in its input
	var result2 Container
	decodeSuccess(t,
		&DecoderConfig{
			Result:      &result2,
			ErrorUnused: true,
		},
		map[string]any{
			"Partial": map[string]any{
				"used":   "value2",
				"unused": "ignored-by-unmarshaler", // Not an error - consumed by Unmarshaler
			},
			"Name": "test2",
		},
	)
	if result2.Partial.UsedValue != "value2" {
		t.Errorf("Expected UsedValue 'value2', got '%s'", result2.Partial.UsedValue)
	}
	if result2.Name != "test2" {
		t.Errorf("Expected Name 'test2', got '%s'", result2.Name)
	}
}

// CustomPlainString is a string alias WITHOUT Unmarshaler for testing WeaklyTypedInput
type CustomPlainString string

// Test that Unmarshaler takes precedence over WeaklyTypedInput
// WeaklyTypedInput should only apply when there's NO Unmarshaler
func TestUnmarshalerWithWeaklyTypedInput(t *testing.T) {
	type Container struct {
		WithUnmarshaler    CustomTrimmedString // Has UnmarshalMapstructure
		WithoutUnmarshaler CustomPlainString   // Plain alias, no Unmarshaler
	}

	// Without WeaklyTypedInput: int input fails for both
	var result1 Container
	decodeFail(t,
		&DecoderConfig{
			Result:           &result1,
			WeaklyTypedInput: false,
		},
		map[string]any{
			"WithUnmarshaler":    123,
			"WithoutUnmarshaler": 456,
		},
		"", // Both fail with type mismatch
	)

	// With WeaklyTypedInput: WeaklyTypedInput applies only to WithoutUnmarshaler
	// WithUnmarshaler still receives raw int (Unmarshaler takes precedence)
	var result2 Container
	decodeFail(t,
		&DecoderConfig{
			Result:           &result2,
			WeaklyTypedInput: true,
		},
		map[string]any{
			"WithUnmarshaler":    123, // Still fails - Unmarshaler gets raw int
			"WithoutUnmarshaler": 456, // Would succeed - WeaklyTypedInput converts int->string
		},
		"expected string input for CustomTrimmedString, got int",
	)

	// Test that WithoutUnmarshaler works when WithUnmarshaler is provided correctly
	var result3 Container
	decodeSuccess(t,
		&DecoderConfig{
			Result:           &result3,
			WeaklyTypedInput: true,
		},
		map[string]any{
			"WithUnmarshaler":    "  hello  ", // String input for Unmarshaler
			"WithoutUnmarshaler": 456,         // Int converted by WeaklyTypedInput
		},
	)
	if result3.WithUnmarshaler != "hello" {
		t.Errorf("Expected WithUnmarshaler 'hello', got '%s'", result3.WithUnmarshaler)
	}
	if result3.WithoutUnmarshaler != "456" {
		t.Errorf("Expected WithoutUnmarshaler '456', got '%s'", result3.WithoutUnmarshaler)
	}
}

// Test that ZeroFields option works correctly with Unmarshaler
// IMPORTANT: ZeroFields does NOT affect types with Unmarshaler!
// Types with Unmarshaler maintain their pre-existing values, and the Unmarshaler
// is responsible for setting/clearing fields as needed.
func TestUnmarshalerWithZeroFields(t *testing.T) {
	type Container struct {
		Partial PartialUnmarshaler
		Name    string
		Count   int
	}

	// Test 1: Without ZeroFields - Unmarshaler doesn't set value, pre-existing value persists
	result1 := Container{
		Partial: PartialUnmarshaler{UsedValue: "pre-existing"},
		Name:    "initial-name",
		Count:   100,
	}
	decodeSuccess(t,
		&DecoderConfig{
			Result:     &result1,
			ZeroFields: false,
		},
		map[string]any{
			"Partial": map[string]any{
				// Not providing "used" key - Unmarshaler won't set UsedValue
			},
			"Name":  "name1",
			"Count": 50,
		},
	)
	// Pre-existing value persists when Unmarshaler doesn't set it
	if result1.Partial.UsedValue != "pre-existing" {
		t.Errorf("Expected UsedValue 'pre-existing', got '%s'", result1.Partial.UsedValue)
	}
	if result1.Name != "name1" {
		t.Errorf("Expected Name 'name1', got '%s'", result1.Name)
	}
	if result1.Count != 50 {
		t.Errorf("Expected Count 50, got %d", result1.Count)
	}

	// Test 2: With ZeroFields true - Unmarshaler types STILL keep pre-existing values!
	// This is the key test: ZeroFields does NOT affect types with Unmarshaler.
	result2 := Container{
		Partial: PartialUnmarshaler{UsedValue: "pre-existing"},
		Name:    "initial-name",
		Count:   100,
	}
	decodeSuccess(t,
		&DecoderConfig{
			Result:     &result2,
			ZeroFields: true,
		},
		map[string]any{
			"Partial": map[string]any{
				// Not setting "used" - Unmarshaler won't modify UsedValue
			},
			"Name":  "name2",
			"Count": 0,
		},
	)
	// Pre-existing value STILL persists! ZeroFields doesn't affect Unmarshaler types
	if result2.Partial.UsedValue != "pre-existing" {
		t.Errorf("Expected UsedValue 'pre-existing' (ZeroFields doesn't affect Unmarshaler), got '%s'", result2.Partial.UsedValue)
	}
	if result2.Name != "name2" {
		t.Errorf("Expected Name 'name2', got '%s'", result2.Name)
	}
	if result2.Count != 0 {
		t.Errorf("Expected Count 0, got %d", result2.Count)
	}

	// Test 3: Verify Unmarshaler still works - when "used" is provided, value is set
	result3 := Container{
		Partial: PartialUnmarshaler{UsedValue: "pre-existing"},
		Name:    "initial-name",
		Count:   100,
	}
	decodeSuccess(t,
		&DecoderConfig{
			Result:     &result3,
			ZeroFields: true,
		},
		map[string]any{
			"Partial": map[string]any{
				"used": "new-value", // Now providing "used" key
			},
			"Name":  "new-name",
			"Count": 75,
		},
	)
	// Now UsedValue is set by Unmarshaler
	if result3.Partial.UsedValue != "new-value" {
		t.Errorf("Expected UsedValue 'new-value', got '%s'", result3.Partial.UsedValue)
	}
	if result3.Name != "new-name" {
		t.Errorf("Expected Name 'new-name', got '%s'", result3.Name)
	}
	if result3.Count != 75 {
		t.Errorf("Expected Count 75, got %d", result3.Count)
	}
}
