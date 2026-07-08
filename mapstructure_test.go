package mapstructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type Basic struct {
	Vstring     string
	Vint        int
	Vint8       int8
	Vint16      int16
	Vint32      int32
	Vint64      int64
	Vuint       uint
	Vbool       bool
	Vfloat      float64
	Vextra      string
	vsilent     bool
	Vdata       any
	VjsonInt    int
	VjsonUint   uint
	VjsonUint64 uint64
	VjsonFloat  float64
	VjsonNumber json.Number
	Vcomplex64  complex64
	Vcomplex128 complex128
}

type BasicPointer struct {
	Vstring     *string
	Vint        *int
	Vuint       *uint
	Vbool       *bool
	Vfloat      *float64
	Vextra      *string
	vsilent     *bool
	Vdata       *any
	VjsonInt    *int
	VjsonFloat  *float64
	VjsonNumber *json.Number
}

type BasicSquash struct {
	Test Basic `mapstructure:",squash"`
}

type BasicJSONInline struct {
	Test Basic `json:",inline"`
}

type Embedded struct {
	Basic
	Vunique string
}

type EmbeddedPointer struct {
	*Basic
	Vunique string
}

type EmbeddedSquash struct {
	Basic   `mapstructure:",squash"`
	Vunique string
}

type EmbeddedPointerSquash struct {
	*Basic  `mapstructure:",squash"`
	Vunique string
}

type BasicMapStructure struct {
	Vunique string     `mapstructure:"vunique"`
	Vtime   *time.Time `mapstructure:"time"`
}

type NestedPointerWithMapstructure struct {
	Vbar *BasicMapStructure `mapstructure:"vbar"`
}

type EmbeddedPointerSquashWithNestedMapstructure struct {
	*NestedPointerWithMapstructure `mapstructure:",squash"`
	Vunique                        string
}

type EmbeddedAndNamed struct {
	Basic
	Named   Basic
	Vunique string
}

type SliceAlias []string

type EmbeddedSlice struct {
	SliceAlias `mapstructure:"slice_alias"`
	Vunique    string
}

type ArrayAlias [2]string

type EmbeddedArray struct {
	ArrayAlias `mapstructure:"array_alias"`
	Vunique    string
}

type SquashOnNonStructType struct {
	InvalidSquashType int `mapstructure:",squash"`
}

type TestInterface interface {
	GetVfoo() string
	GetVbarfoo() string
	GetVfoobar() string
}

type TestInterfaceImpl struct {
	Vfoo string
}

func (t *TestInterfaceImpl) GetVfoo() string {
	return t.Vfoo
}

func (t *TestInterfaceImpl) GetVbarfoo() string {
	return ""
}

func (t *TestInterfaceImpl) GetVfoobar() string {
	return ""
}

type TestNestedInterfaceImpl struct {
	SquashOnNestedInterfaceType `mapstructure:",squash"`
	Vfoo                        string
}

func (t *TestNestedInterfaceImpl) GetVfoo() string {
	return t.Vfoo
}

func (t *TestNestedInterfaceImpl) GetVbarfoo() string {
	return t.Vbarfoo
}

func (t *TestNestedInterfaceImpl) GetVfoobar() string {
	return t.NestedSquash.Vfoobar
}

type SquashOnInterfaceType struct {
	TestInterface `mapstructure:",squash"`
	Vbar          string
}

type NestedSquash struct {
	SquashOnInterfaceType `mapstructure:",squash"`
	Vfoobar               string
}

type SquashOnNestedInterfaceType struct {
	NestedSquash NestedSquash `mapstructure:",squash"`
	Vbarfoo      string
}

type Map struct {
	Vfoo   string
	Vother map[string]string
}

type MapOfStruct struct {
	Value map[string]Basic
}

type Nested struct {
	Vfoo string
	Vbar Basic
}

type NestedPointer struct {
	Vfoo string
	Vbar *Basic
}

type NilInterface struct {
	W io.Writer
}

type NilPointer struct {
	Value *string
}

type Slice struct {
	Vfoo string
	Vbar []string
}

type SliceOfByte struct {
	Vfoo string
	Vbar []byte
}

type SliceOfAlias struct {
	Vfoo string
	Vbar SliceAlias
}

type SliceOfStruct struct {
	Value []Basic
}

type SlicePointer struct {
	Vbar *[]string
}

type Array struct {
	Vfoo string
	Vbar [2]string
}

type ArrayOfStruct struct {
	Value [2]Basic
}

type Func struct {
	Foo func() string
}

type Tagged struct {
	Extra string `mapstructure:"bar,what,what"`
	Value string `mapstructure:"foo"`
}

type Remainder struct {
	A     string
	Extra map[string]any `mapstructure:",remain"`
}

type StructWithOmitEmpty struct {
	VisibleStringField string         `mapstructure:"visible-string"`
	OmitStringField    string         `mapstructure:"omittable-string,omitempty"`
	VisibleIntField    int            `mapstructure:"visible-int"`
	OmitIntField       int            `mapstructure:"omittable-int,omitempty"`
	VisibleFloatField  float64        `mapstructure:"visible-float"`
	OmitFloatField     float64        `mapstructure:"omittable-float,omitempty"`
	VisibleSliceField  []any          `mapstructure:"visible-slice"`
	OmitSliceField     []any          `mapstructure:"omittable-slice,omitempty"`
	VisibleMapField    map[string]any `mapstructure:"visible-map"`
	OmitMapField       map[string]any `mapstructure:"omittable-map,omitempty"`
	NestedField        *Nested        `mapstructure:"visible-nested"`
	OmitNestedField    *Nested        `mapstructure:"omittable-nested,omitempty"`
}

type StructWithOmitZero struct {
	VisibleStringField string         `mapstructure:"visible-string"`
	OmitStringField    string         `mapstructure:"omittable-string,omitzero"`
	VisibleIntField    int            `mapstructure:"visible-int"`
	OmitIntField       int            `mapstructure:"omittable-int,omitzero"`
	VisibleFloatField  float64        `mapstructure:"visible-float"`
	OmitFloatField     float64        `mapstructure:"omittable-float,omitzero"`
	VisibleSliceField  []any          `mapstructure:"visible-slice"`
	OmitSliceField     []any          `mapstructure:"omittable-slice,omitzero"`
	VisibleMapField    map[string]any `mapstructure:"visible-map"`
	OmitMapField       map[string]any `mapstructure:"omittable-map,omitzero"`
	NestedField        *Nested        `mapstructure:"visible-nested"`
	OmitNestedField    *Nested        `mapstructure:"omittable-nested,omitzero"`
}

type TypeConversionResult struct {
	IntToFloat         float32
	IntToUint          uint
	IntToBool          bool
	IntToString        string
	UintToInt          int
	UintToFloat        float32
	UintToBool         bool
	UintToString       string
	BoolToInt          int
	BoolToUint         uint
	BoolToFloat        float32
	BoolToString       string
	FloatToInt         int
	FloatToUint        uint
	FloatToBool        bool
	FloatToString      string
	SliceUint8ToString string
	StringToSliceUint8 []byte
	ArrayUint8ToString string
	StringToInt        int
	StringToUint       uint
	StringToBool       bool
	StringToFloat      float32
	StringToStrSlice   []string
	StringToIntSlice   []int
	StringToStrArray   [1]string
	StringToIntArray   [1]int
	SliceToMap         map[string]any
	MapToSlice         []any
	ArrayToMap         map[string]any
	MapToArray         [1]any
}

func TestBasicTypes(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring":     "foo",
		"vint":        42,
		"vint8":       42,
		"vint16":      42,
		"vint32":      42,
		"vint64":      42,
		"Vuint":       42,
		"vbool":       true,
		"Vfloat":      42.42,
		"vsilent":     true,
		"vdata":       42,
		"vjsonInt":    json.Number("1234"),
		"vjsonUint":   json.Number("1234"),
		"vjsonUint64": json.Number("9223372036854775809"), // 2^63 + 1
		"vjsonFloat":  json.Number("1234.5"),
		"vjsonNumber": json.Number("1234.5"),
		"vcomplex64":  complex(float32(42), float32(42)),
		"vcomplex128": complex(42, 42),
	}

	var result Basic
	err := Decode(input, &result)
	if err != nil {
		t.Errorf("got an err: %s", err.Error())
		t.FailNow()
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vint)
	}
	if result.Vint8 != 42 {
		t.Errorf("vint8 value should be 42: %#v", result.Vint)
	}
	if result.Vint16 != 42 {
		t.Errorf("vint16 value should be 42: %#v", result.Vint)
	}
	if result.Vint32 != 42 {
		t.Errorf("vint32 value should be 42: %#v", result.Vint)
	}
	if result.Vint64 != 42 {
		t.Errorf("vint64 value should be 42: %#v", result.Vint)
	}

	if result.Vuint != 42 {
		t.Errorf("vuint value should be 42: %#v", result.Vuint)
	}

	if result.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbool)
	}

	if result.Vfloat != 42.42 {
		t.Errorf("vfloat value should be 42.42: %#v", result.Vfloat)
	}

	if result.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vextra)
	}

	if result.vsilent != false {
		t.Error("vsilent should not be set, it is unexported")
	}

	if result.Vdata != 42 {
		t.Error("vdata should be valid")
	}

	if result.VjsonInt != 1234 {
		t.Errorf("vjsonint value should be 1234: %#v", result.VjsonInt)
	}

	if result.VjsonUint != 1234 {
		t.Errorf("vjsonuint value should be 1234: %#v", result.VjsonUint)
	}

	if result.VjsonUint64 != 9223372036854775809 {
		t.Errorf("vjsonuint64 value should be 9223372036854775809: %#v", result.VjsonUint64)
	}

	if result.VjsonFloat != 1234.5 {
		t.Errorf("vjsonfloat value should be 1234.5: %#v", result.VjsonFloat)
	}

	if !reflect.DeepEqual(result.VjsonNumber, json.Number("1234.5")) {
		t.Errorf("vjsonnumber value should be '1234.5': %T, %#v", result.VjsonNumber, result.VjsonNumber)
	}

	if real(result.Vcomplex64) != 42 || imag(result.Vcomplex64) != 42 {
		t.Errorf("vcomplex64 value shou be 42+42i: %#v", result.Vcomplex64)
	}

	if real(result.Vcomplex128) != 42 || imag(result.Vcomplex128) != 42 {
		t.Errorf("vcomplex64 value shou be 42+42i: %#v", result.Vcomplex128)
	}
}

func TestBasic_IntWithFloat(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vint": float64(42),
	}

	var result Basic
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}
}

func TestBasic_Merge(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vint": 42,
	}

	var result Basic
	result.Vuint = 100
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	expected := Basic{
		Vint:  42,
		Vuint: 100,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

// Test for issue #46.
func TestBasic_Struct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vdata": map[string]any{
			"vstring": "foo",
		},
	}

	var result, inner Basic
	result.Vdata = &inner
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}
	expected := Basic{
		Vdata: &Basic{
			Vstring: "foo",
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestBasic_interfaceStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
	}

	var iface any = &Basic{}
	err := Decode(input, &iface)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	expected := &Basic{
		Vstring: "foo",
	}
	if !reflect.DeepEqual(iface, expected) {
		t.Fatalf("bad: %#v", iface)
	}
}

// Issue 187
func TestBasic_interfaceStructNonPtr(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
	}

	var iface any = Basic{}
	err := Decode(input, &iface)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	expected := Basic{
		Vstring: "foo",
	}
	if !reflect.DeepEqual(iface, expected) {
		t.Fatalf("bad: %#v", iface)
	}
}

func TestDecode_BasicSquash(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
	}

	var result BasicSquash
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Test.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Test.Vstring)
	}
}

func TestDecodeFrom_BasicSquash(t *testing.T) {
	t.Parallel()

	var v any
	var ok bool

	input := BasicSquash{
		Test: Basic{
			Vstring: "foo",
		},
	}

	var result map[string]any
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if _, ok = result["Test"]; ok {
		t.Error("test should not be present in map")
	}

	v, ok = result["Vstring"]
	if !ok {
		t.Error("vstring should be present in map")
	} else if !reflect.DeepEqual(v, "foo") {
		t.Errorf("vstring value should be 'foo': %#v", v)
	}
}

func TestDecode_BasicJSONInline(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
	}

	var result BasicJSONInline
	d, err := NewDecoder(&DecoderConfig{TagName: "json", SquashTagOption: "inline", Result: &result})
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if err := d.Decode(input); err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Test.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Test.Vstring)
	}
}

func TestDecodeFrom_BasicJSONInline(t *testing.T) {
	t.Parallel()

	var v any
	var ok bool

	input := BasicJSONInline{
		Test: Basic{
			Vstring: "foo",
		},
	}

	var result map[string]any
	d, err := NewDecoder(&DecoderConfig{TagName: "json", SquashTagOption: "inline", Result: &result})
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if err := d.Decode(input); err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if _, ok = result["Test"]; ok {
		t.Error("test should not be present in map")
	}

	v, ok = result["Vstring"]
	if !ok {
		t.Error("vstring should be present in map")
	} else if !reflect.DeepEqual(v, "foo") {
		t.Errorf("vstring value should be 'foo': %#v", v)
	}
}

func TestDecode_Embedded(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"Basic": map[string]any{
			"vstring": "innerfoo",
		},
		"vunique": "bar",
	}

	var result Embedded
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "innerfoo" {
		t.Errorf("vstring value should be 'innerfoo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedPointer(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"Basic": map[string]any{
			"vstring": "innerfoo",
		},
		"vunique": "bar",
	}

	var result EmbeddedPointer
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := EmbeddedPointer{
		Basic: &Basic{
			Vstring: "innerfoo",
		},
		Vunique: "bar",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestDecode_EmbeddedSlice(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"slice_alias": []string{"foo", "bar"},
		"vunique":     "bar",
	}

	var result EmbeddedSlice
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if !reflect.DeepEqual(result.SliceAlias, SliceAlias([]string{"foo", "bar"})) {
		t.Errorf("slice value: %#v", result.SliceAlias)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedArray(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"array_alias": [2]string{"foo", "bar"},
		"vunique":     "bar",
	}

	var result EmbeddedArray
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if !reflect.DeepEqual(result.ArrayAlias, ArrayAlias([2]string{"foo", "bar"})) {
		t.Errorf("array value: %#v", result.ArrayAlias)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_decodeSliceWithArray(t *testing.T) {
	t.Parallel()

	var result []int
	input := [1]int{1}
	expected := []int{1}
	if err := Decode(input, &result); err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %+v, got %+v", expected, result)
	}
}

func TestDecode_EmbeddedNoSquash(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"vunique": "bar",
	}

	var result Embedded
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "" {
		t.Errorf("vstring value should be empty: %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedPointerNoSquash(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"vunique": "bar",
	}

	result := EmbeddedPointer{
		Basic: &Basic{},
	}

	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Vstring != "" {
		t.Errorf("vstring value should be empty: %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedSquash(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"vunique": "bar",
	}

	var result EmbeddedSquash
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecodeFrom_EmbeddedSquash(t *testing.T) {
	t.Parallel()

	var v any
	var ok bool

	input := EmbeddedSquash{
		Basic: Basic{
			Vstring: "foo",
		},
		Vunique: "bar",
	}

	var result map[string]any
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if _, ok = result["Basic"]; ok {
		t.Error("basic should not be present in map")
	}

	v, ok = result["Vstring"]
	if !ok {
		t.Error("vstring should be present in map")
	} else if !reflect.DeepEqual(v, "foo") {
		t.Errorf("vstring value should be 'foo': %#v", v)
	}

	v, ok = result["Vunique"]
	if !ok {
		t.Error("vunique should be present in map")
	} else if !reflect.DeepEqual(v, "bar") {
		t.Errorf("vunique value should be 'bar': %#v", v)
	}
}

func TestDecode_EmbeddedPointerSquash_FromStructToMap(t *testing.T) {
	t.Parallel()

	input := EmbeddedPointerSquash{
		Basic: &Basic{
			Vstring: "foo",
		},
		Vunique: "bar",
	}

	var result map[string]any
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result["Vstring"] != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result["Vstring"])
	}

	if result["Vunique"] != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result["Vunique"])
	}
}

func TestDecode_EmbeddedPointerSquash_FromMapToStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"Vstring": "foo",
		"Vunique": "bar",
	}

	result := EmbeddedPointerSquash{
		Basic: &Basic{},
	}
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedPointerSquash_WithoutPreInitializedStructs_FromMapToStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"Vstring": "foo",
		"Vunique": "bar",
	}

	result := EmbeddedPointerSquash{}
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedPointerSquashWithNestedMapstructure_FromStructToMap(t *testing.T) {
	t.Parallel()

	vTime := time.Now()

	input := EmbeddedPointerSquashWithNestedMapstructure{
		NestedPointerWithMapstructure: &NestedPointerWithMapstructure{
			Vbar: &BasicMapStructure{
				Vunique: "bar",
				Vtime:   &vTime,
			},
		},
		Vunique: "foo",
	}

	var result map[string]any
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}
	expected := map[string]any{
		"vbar": map[string]any{
			"vunique": "bar",
			"time":    &vTime,
		},
		"Vunique": "foo",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result should be %#v: got %#v", expected, result)
	}
}

func TestDecode_EmbeddedPointerSquashWithNestedMapstructure_FromMapToStruct(t *testing.T) {
	t.Parallel()

	vTime := time.Now()

	input := map[string]any{
		"vbar": map[string]any{
			"vunique": "bar",
			"time":    &vTime,
		},
		"Vunique": "foo",
	}

	result := EmbeddedPointerSquashWithNestedMapstructure{
		NestedPointerWithMapstructure: &NestedPointerWithMapstructure{},
	}
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}
	expected := EmbeddedPointerSquashWithNestedMapstructure{
		NestedPointerWithMapstructure: &NestedPointerWithMapstructure{
			Vbar: &BasicMapStructure{
				Vunique: "bar",
				Vtime:   &vTime,
			},
		},
		Vunique: "foo",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result should be %#v: got %#v", expected, result)
	}
}

func TestDecode_EmbeddedSquashConfig(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"vunique": "bar",
		"Named": map[string]any{
			"vstring": "baz",
		},
	}

	var result EmbeddedAndNamed
	config := &DecoderConfig{
		Squash: true,
		Result: &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}

	if result.Named.Vstring != "baz" {
		t.Errorf("Named.vstring value should be 'baz': %#v", result.Named.Vstring)
	}
}

func TestDecodeFrom_EmbeddedSquashConfig(t *testing.T) {
	t.Parallel()

	input := EmbeddedAndNamed{
		Basic:   Basic{Vstring: "foo"},
		Named:   Basic{Vstring: "baz"},
		Vunique: "bar",
	}

	result := map[string]any{}
	config := &DecoderConfig{
		Squash: true,
		Result: &result,
	}
	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if _, ok := result["Basic"]; ok {
		t.Error("basic should not be present in map")
	}

	v, ok := result["Vstring"]
	if !ok {
		t.Error("vstring should be present in map")
	} else if !reflect.DeepEqual(v, "foo") {
		t.Errorf("vstring value should be 'foo': %#v", v)
	}

	v, ok = result["Vunique"]
	if !ok {
		t.Error("vunique should be present in map")
	} else if !reflect.DeepEqual(v, "bar") {
		t.Errorf("vunique value should be 'bar': %#v", v)
	}

	v, ok = result["Named"]
	if !ok {
		t.Error("Named should be present in map")
	} else {
		named := v.(map[string]any)
		v, ok := named["Vstring"]
		if !ok {
			t.Error("Named: vstring should be present in map")
		} else if !reflect.DeepEqual(v, "baz") {
			t.Errorf("Named: vstring should be 'baz': %#v", v)
		}
	}
}

func TestDecodeFrom_EmbeddedSquashConfig_WithTags(t *testing.T) {
	t.Parallel()

	var v any
	var ok bool

	input := EmbeddedSquash{
		Basic: Basic{
			Vstring: "foo",
		},
		Vunique: "bar",
	}

	result := map[string]any{}
	config := &DecoderConfig{
		Squash: true,
		Result: &result,
	}
	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if _, ok = result["Basic"]; ok {
		t.Error("basic should not be present in map")
	}

	v, ok = result["Vstring"]
	if !ok {
		t.Error("vstring should be present in map")
	} else if !reflect.DeepEqual(v, "foo") {
		t.Errorf("vstring value should be 'foo': %#v", v)
	}

	v, ok = result["Vunique"]
	if !ok {
		t.Error("vunique should be present in map")
	} else if !reflect.DeepEqual(v, "bar") {
		t.Errorf("vunique value should be 'bar': %#v", v)
	}
}

func TestDecode_SquashOnNonStructType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"InvalidSquashType": 42,
	}

	var result SquashOnNonStructType
	err := Decode(input, &result)
	if err == nil {
		t.Fatal("unexpected success decoding invalid squash field type")
	} else if !strings.Contains(err.Error(), "unsupported type for squash") {
		t.Fatalf("unexpected error message for invalid squash field type: %s", err)
	}
}

func TestDecode_SquashOnInterfaceType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"VFoo": "42",
		"VBar": "43",
	}

	result := SquashOnInterfaceType{
		TestInterface: &TestInterfaceImpl{},
	}
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	res := result.GetVfoo()
	if res != "42" {
		t.Errorf("unexpected value for VFoo: %s", res)
	}

	res = result.Vbar
	if res != "43" {
		t.Errorf("unexpected value for Vbar: %s", res)
	}
}

func TestDecode_SquashOnOuterNestedInterfaceType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"VFoo":    "42",
		"VBar":    "43",
		"Vfoobar": "44",
		"Vbarfoo": "45",
	}

	result := SquashOnNestedInterfaceType{
		NestedSquash: NestedSquash{
			SquashOnInterfaceType: SquashOnInterfaceType{
				TestInterface: &TestInterfaceImpl{},
			},
		},
	}

	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	res := result.NestedSquash.GetVfoo()
	if res != "42" {
		t.Errorf("unexpected value for VFoo: %s", res)
	}

	res = result.NestedSquash.Vbar
	if res != "43" {
		t.Errorf("unexpected value for Vbar: %s", res)
	}

	res = result.NestedSquash.Vfoobar
	if res != "44" {
		t.Errorf("unexpected value for Vfoobar: %s", res)
	}

	res = result.Vbarfoo
	if res != "45" {
		t.Errorf("unexpected value for Vbarfoo: %s", res)
	}
}

func TestDecode_SquashOnInnerNestedInterfaceType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"VFoo":    "42",
		"VBar":    "43",
		"Vfoobar": "44",
		"Vbarfoo": "45",
	}

	result := SquashOnInterfaceType{
		TestInterface: &TestNestedInterfaceImpl{
			SquashOnNestedInterfaceType: SquashOnNestedInterfaceType{
				NestedSquash: NestedSquash{
					SquashOnInterfaceType: SquashOnInterfaceType{
						TestInterface: &TestInterfaceImpl{},
					},
				},
			},
		},
	}

	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	res := result.GetVfoo()
	if res != "42" {
		t.Errorf("unexpected value for VFoo: %s", res)
	}

	res = result.Vbar
	if res != "43" {
		t.Errorf("unexpected value for Vbar: %s", res)
	}

	res = result.GetVfoobar()
	if res != "44" {
		t.Errorf("unexpected value for Vfoobar: %s", res)
	}

	res = result.GetVbarfoo()
	if res != "45" {
		t.Errorf("unexpected value for Vbarfoo: %s", res)
	}
}

func TestDecode_SquashOnNilInterfaceType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"VFoo": "42",
		"VBar": "43",
	}

	result := SquashOnInterfaceType{
		TestInterface: nil,
	}
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	res := result.Vbar
	if res != "43" {
		t.Errorf("unexpected value for Vbar: %s", res)
	}
}

func TestDecode_DecodeHook(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vint": "WHAT",
	}

	decodeHook := func(from reflect.Kind, to reflect.Kind, v any) (any, error) {
		if from == reflect.String && to != reflect.String {
			return 5, nil
		}

		return v, nil
	}

	var result Basic
	config := &DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Vint != 5 {
		t.Errorf("vint should be 5: %#v", result.Vint)
	}
}

func TestDecode_DecodeHookType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vint": "WHAT",
	}

	decodeHook := func(from reflect.Type, to reflect.Type, v any) (any, error) {
		if from.Kind() == reflect.String &&
			to.Kind() != reflect.String {
			return 5, nil
		}

		return v, nil
	}

	var result Basic
	config := &DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Vint != 5 {
		t.Errorf("vint should be 5: %#v", result.Vint)
	}
}

func TestDecode_Nil(t *testing.T) {
	t.Parallel()

	var input any
	result := Basic{
		Vstring: "foo",
	}

	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Vstring != "foo" {
		t.Fatalf("bad: %#v", result.Vstring)
	}
}

func TestDecode_NilInterfaceHook(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"w": "",
	}

	decodeHook := func(f, t reflect.Type, v any) (any, error) {
		if t.String() == "io.Writer" {
			return nil, nil
		}

		return v, nil
	}

	var result NilInterface
	config := &DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.W != nil {
		t.Errorf("W should be nil: %#v", result.W)
	}
}

func TestDecode_NilPointerHook(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"value": "",
	}

	decodeHook := func(f, t reflect.Type, v any) (any, error) {
		if typed, ok := v.(string); ok {
			if typed == "" {
				return nil, nil
			}
		}
		return v, nil
	}

	var result NilPointer
	config := &DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Value != nil {
		t.Errorf("W should be nil: %#v", result.Value)
	}
}

func TestDecode_FuncHook(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"foo": "baz",
	}

	decodeHook := func(f, t reflect.Type, v any) (any, error) {
		if t.Kind() != reflect.Func {
			return v, nil
		}
		val := v.(string)
		return func() string { return val }, nil
	}

	var result Func
	config := &DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Foo() != "baz" {
		t.Errorf("Foo call result should be 'baz': %s", result.Foo())
	}
}

func TestDecode_NonStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"foo": "bar",
		"bar": "baz",
	}

	var result map[string]string
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result["foo"] != "bar" {
		t.Fatal("foo is not bar")
	}
}

func TestDecode_StructMatch(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vbar": Basic{
			Vstring: "foo",
		},
	}

	var result Nested
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vbar.Vstring != "foo" {
		t.Errorf("bad: %#v", result)
	}
}

func TestDecode_TypeConversion(t *testing.T) {
	input := map[string]any{
		"IntToFloat":         42,
		"IntToUint":          42,
		"IntToBool":          1,
		"IntToString":        42,
		"UintToInt":          42,
		"UintToFloat":        42,
		"UintToBool":         42,
		"UintToString":       42,
		"BoolToInt":          true,
		"BoolToUint":         true,
		"BoolToFloat":        true,
		"BoolToString":       true,
		"FloatToInt":         42.42,
		"FloatToUint":        42.42,
		"FloatToBool":        42.42,
		"FloatToString":      42.42,
		"SliceUint8ToString": []uint8("foo"),
		"StringToSliceUint8": "foo",
		"ArrayUint8ToString": [3]uint8{'f', 'o', 'o'},
		"StringToInt":        "42",
		"StringToUint":       "42",
		"StringToBool":       "1",
		"StringToFloat":      "42.42",
		"StringToStrSlice":   "A",
		"StringToIntSlice":   "42",
		"StringToStrArray":   "A",
		"StringToIntArray":   "42",
		"SliceToMap":         []any{},
		"MapToSlice":         map[string]any{},
		"ArrayToMap":         []any{},
		"MapToArray":         map[string]any{},
	}

	expectedResultStrict := TypeConversionResult{
		IntToFloat:  42.0,
		IntToUint:   42,
		UintToInt:   42,
		UintToFloat: 42,
		BoolToInt:   0,
		BoolToUint:  0,
		BoolToFloat: 0,
		FloatToInt:  42,
		FloatToUint: 42,
	}

	expectedResultWeak := TypeConversionResult{
		IntToFloat:         42.0,
		IntToUint:          42,
		IntToBool:          true,
		IntToString:        "42",
		UintToInt:          42,
		UintToFloat:        42,
		UintToBool:         true,
		UintToString:       "42",
		BoolToInt:          1,
		BoolToUint:         1,
		BoolToFloat:        1,
		BoolToString:       "1",
		FloatToInt:         42,
		FloatToUint:        42,
		FloatToBool:        true,
		FloatToString:      "42.42",
		SliceUint8ToString: "foo",
		StringToSliceUint8: []byte("foo"),
		ArrayUint8ToString: "foo",
		StringToInt:        42,
		StringToUint:       42,
		StringToBool:       true,
		StringToFloat:      42.42,
		StringToStrSlice:   []string{"A"},
		StringToIntSlice:   []int{42},
		StringToStrArray:   [1]string{"A"},
		StringToIntArray:   [1]int{42},
		SliceToMap:         map[string]any{},
		MapToSlice:         []any{},
		ArrayToMap:         map[string]any{},
		MapToArray:         [1]any{},
	}

	// Test strict type conversion
	var resultStrict TypeConversionResult
	err := Decode(input, &resultStrict)
	if err == nil {
		t.Errorf("should return an error")
	}
	if !reflect.DeepEqual(resultStrict, expectedResultStrict) {
		t.Errorf("expected %v, got: %v", expectedResultStrict, resultStrict)
	}

	// Test weak type conversion
	var decoder *Decoder
	var resultWeak TypeConversionResult

	config := &DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &resultWeak,
	}

	decoder, err = NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if !reflect.DeepEqual(resultWeak, expectedResultWeak) {
		t.Errorf("expected \n%#v, got: \n%#v", expectedResultWeak, resultWeak)
	}
}

func TestDecoder_ErrorUnused(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "hello",
		"foo":     "bar",
	}

	var result Basic
	config := &DecoderConfig{
		ErrorUnused: true,
		Result:      &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDecoder_ErrorUnused_NotSetable(t *testing.T) {
	t.Parallel()

	// lowercase vsilent is unexported and cannot be set
	input := map[string]any{
		"vsilent": "false",
	}

	var result Basic
	config := &DecoderConfig{
		ErrorUnused: true,
		Result:      &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDecoder_ErrorUnset(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "hello",
		"foo":     "bar",
	}

	var result Basic
	config := &DecoderConfig{
		ErrorUnset: true,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDecoder_ErrorUnset_AllowUnsetPointer(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "hello",
		"foo":     "bar",
	}

	var result BasicPointer
	config := &DecoderConfig{
		ErrorUnset:        true,
		AllowUnsetPointer: true,
		Result:            &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatal("error not expected")
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vother": map[any]any{
			"foo": "foo",
			"bar": "bar",
		},
	}

	var result Map
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an error: %s", err)
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vother == nil {
		t.Fatal("vother should not be nil")
	}

	if len(result.Vother) != 2 {
		t.Error("vother should have two items")
	}

	if result.Vother["foo"] != "foo" {
		t.Errorf("'foo' key should be foo, got: %#v", result.Vother["foo"])
	}

	if result.Vother["bar"] != "bar" {
		t.Errorf("'bar' key should be bar, got: %#v", result.Vother["bar"])
	}
}

func TestMapMerge(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vother": map[any]any{
			"foo": "foo",
			"bar": "bar",
		},
	}

	var result Map
	result.Vother = map[string]string{"hello": "world"}
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an error: %s", err)
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	expected := map[string]string{
		"foo":   "foo",
		"bar":   "bar",
		"hello": "world",
	}
	if !reflect.DeepEqual(result.Vother, expected) {
		t.Errorf("bad: %#v", result.Vother)
	}
}

func TestMapOfStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"value": map[string]any{
			"foo": map[string]string{"vstring": "one"},
			"bar": map[string]string{"vstring": "two"},
		},
	}

	var result MapOfStruct
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Value == nil {
		t.Fatal("value should not be nil")
	}

	if len(result.Value) != 2 {
		t.Error("value should have two items")
	}

	if result.Value["foo"].Vstring != "one" {
		t.Errorf("foo value should be 'one', got: %s", result.Value["foo"].Vstring)
	}

	if result.Value["bar"].Vstring != "two" {
		t.Errorf("bar value should be 'two', got: %s", result.Value["bar"].Vstring)
	}
}

func TestNestedType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vbar": map[string]any{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,
		},
	}

	var result Nested
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vbar.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vbar.Vstring)
	}

	if result.Vbar.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vbar.Vint)
	}

	if result.Vbar.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbar.Vbool)
	}

	if result.Vbar.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vbar.Vextra)
	}
}

func TestNestedTypePointer(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vbar": &map[string]any{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,
		},
	}

	var result NestedPointer
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vbar.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vbar.Vstring)
	}

	if result.Vbar.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vbar.Vint)
	}

	if result.Vbar.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbar.Vbool)
	}

	if result.Vbar.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vbar.Vextra)
	}
}

// Test for issue #46.
func TestNestedTypeInterface(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vbar": &map[string]any{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,

			"vdata": map[string]any{
				"vstring": "bar",
			},
		},
	}

	var result NestedPointer
	result.Vbar = new(Basic)
	result.Vbar.Vdata = new(Basic)
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vbar.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vbar.Vstring)
	}

	if result.Vbar.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vbar.Vint)
	}

	if result.Vbar.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbar.Vbool)
	}

	if result.Vbar.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vbar.Vextra)
	}

	if result.Vbar.Vdata.(*Basic).Vstring != "bar" {
		t.Errorf("vstring value should be 'bar': %#v", result.Vbar.Vdata.(*Basic).Vstring)
	}
}

func TestSlice(t *testing.T) {
	t.Parallel()

	inputStringSlice := map[string]any{
		"vfoo": "foo",
		"vbar": []string{"foo", "bar", "baz"},
	}

	inputStringSlicePointer := map[string]any{
		"vfoo": "foo",
		"vbar": &[]string{"foo", "bar", "baz"},
	}

	outputStringSlice := &Slice{
		"foo",
		[]string{"foo", "bar", "baz"},
	}

	testSliceInput(t, inputStringSlice, outputStringSlice)
	testSliceInput(t, inputStringSlicePointer, outputStringSlice)
}

func TestNotEmptyByteSlice(t *testing.T) {
	t.Parallel()

	inputByteSlice := map[string]any{
		"vfoo": "foo",
		"vbar": []byte(`{"bar": "bar"}`),
	}

	result := SliceOfByte{
		Vfoo: "another foo",
		Vbar: []byte(`{"bar": "bar bar bar bar bar bar bar bar"}`),
	}

	err := Decode(inputByteSlice, &result)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	expected := SliceOfByte{
		Vfoo: "foo",
		Vbar: []byte(`{"bar": "bar"}`),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("bad: %#v", result)
	}
}

func TestInvalidSlice(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vbar": 42,
	}

	result := Slice{}
	err := Decode(input, &result)
	if err == nil {
		t.Errorf("expected failure")
	}
}

func TestSliceOfStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"value": []map[string]any{
			{"vstring": "one"},
			{"vstring": "two"},
		},
	}

	var result SliceOfStruct
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if len(result.Value) != 2 {
		t.Fatalf("expected two values, got %d", len(result.Value))
	}

	if result.Value[0].Vstring != "one" {
		t.Errorf("first value should be 'one', got: %s", result.Value[0].Vstring)
	}

	if result.Value[1].Vstring != "two" {
		t.Errorf("second value should be 'two', got: %s", result.Value[1].Vstring)
	}
}

func TestSliceCornerCases(t *testing.T) {
	t.Parallel()

	// Input with a map with zero values
	input := map[string]any{}
	var resultWeak []Basic

	err := WeakDecode(input, &resultWeak)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if len(resultWeak) != 0 {
		t.Errorf("length should be 0")
	}
	// Input with more values
	input = map[string]any{
		"Vstring": "foo",
	}

	resultWeak = nil
	err = WeakDecode(input, &resultWeak)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if resultWeak[0].Vstring != "foo" {
		t.Errorf("value does not match")
	}
}

func TestSliceToMap(t *testing.T) {
	t.Parallel()

	input := []map[string]any{
		{
			"foo": "bar",
		},
		{
			"bar": "baz",
		},
	}

	var result map[string]any
	err := WeakDecode(input, &result)
	if err != nil {
		t.Fatalf("got an error: %s", err)
	}

	expected := map[string]any{
		"foo": "bar",
		"bar": "baz",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("bad: %#v", result)
	}
}

func TestArray(t *testing.T) {
	t.Parallel()

	inputStringArray := map[string]any{
		"vfoo": "foo",
		"vbar": [2]string{"foo", "bar"},
	}

	inputStringArrayPointer := map[string]any{
		"vfoo": "foo",
		"vbar": &[2]string{"foo", "bar"},
	}

	outputStringArray := &Array{
		"foo",
		[2]string{"foo", "bar"},
	}

	testArrayInput(t, inputStringArray, outputStringArray)
	testArrayInput(t, inputStringArrayPointer, outputStringArray)
}

func TestInvalidArray(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vbar": 42,
	}

	result := Array{}
	err := Decode(input, &result)
	if err == nil {
		t.Errorf("expected failure")
	}
}

func TestArrayOfStruct(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"value": []map[string]any{
			{"vstring": "one"},
			{"vstring": "two"},
		},
	}

	var result ArrayOfStruct
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if len(result.Value) != 2 {
		t.Fatalf("expected two values, got %d", len(result.Value))
	}

	if result.Value[0].Vstring != "one" {
		t.Errorf("first value should be 'one', got: %s", result.Value[0].Vstring)
	}

	if result.Value[1].Vstring != "two" {
		t.Errorf("second value should be 'two', got: %s", result.Value[1].Vstring)
	}
}

func TestArrayToMap(t *testing.T) {
	t.Parallel()

	input := []map[string]any{
		{
			"foo": "bar",
		},
		{
			"bar": "baz",
		},
	}

	var result map[string]any
	err := WeakDecode(input, &result)
	if err != nil {
		t.Fatalf("got an error: %s", err)
	}

	expected := map[string]any{
		"foo": "bar",
		"bar": "baz",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("bad: %#v", result)
	}
}

func TestDecodeTable(t *testing.T) {
	t.Parallel()

	// We need to make new types so that we don't get the short-circuit
	// copy functionality. We want to test the deep copying functionality.
	type BasicCopy Basic
	type NestedPointerCopy NestedPointer
	type MapCopy Map

	tests := []struct {
		name    string
		in      any
		target  any
		out     any
		wantErr bool
	}{
		{
			"basic struct input",
			&Basic{
				Vstring: "vstring",
				Vint:    2,
				Vint8:   2,
				Vint16:  2,
				Vint32:  2,
				Vint64:  2,
				Vuint:   3,
				Vbool:   true,
				Vfloat:  4.56,
				Vextra:  "vextra",
				vsilent: true,
				Vdata:   []byte("data"),
			},
			&map[string]any{},
			&map[string]any{
				"Vstring":     "vstring",
				"Vint":        2,
				"Vint8":       int8(2),
				"Vint16":      int16(2),
				"Vint32":      int32(2),
				"Vint64":      int64(2),
				"Vuint":       uint(3),
				"Vbool":       true,
				"Vfloat":      4.56,
				"Vextra":      "vextra",
				"Vdata":       []byte("data"),
				"VjsonInt":    0,
				"VjsonUint":   uint(0),
				"VjsonUint64": uint64(0),
				"VjsonFloat":  0.0,
				"VjsonNumber": json.Number(""),
				"Vcomplex64":  complex64(0),
				"Vcomplex128": complex128(0),
			},
			false,
		},
		{
			"embedded struct input",
			&Embedded{
				Vunique: "vunique",
				Basic: Basic{
					Vstring: "vstring",
					Vint:    2,
					Vint8:   2,
					Vint16:  2,
					Vint32:  2,
					Vint64:  2,
					Vuint:   3,
					Vbool:   true,
					Vfloat:  4.56,
					Vextra:  "vextra",
					vsilent: true,
					Vdata:   []byte("data"),
				},
			},
			&map[string]any{},
			&map[string]any{
				"Vunique": "vunique",
				"Basic": map[string]any{
					"Vstring":     "vstring",
					"Vint":        2,
					"Vint8":       int8(2),
					"Vint16":      int16(2),
					"Vint32":      int32(2),
					"Vint64":      int64(2),
					"Vuint":       uint(3),
					"Vbool":       true,
					"Vfloat":      4.56,
					"Vextra":      "vextra",
					"Vdata":       []byte("data"),
					"VjsonInt":    0,
					"VjsonUint":   uint(0),
					"VjsonUint64": uint64(0),
					"VjsonFloat":  0.0,
					"VjsonNumber": json.Number(""),
					"Vcomplex64":  complex64(0),
					"Vcomplex128": complex128(0),
				},
			},
			false,
		},
		{
			"struct => struct",
			&Basic{
				Vstring: "vstring",
				Vint:    2,
				Vuint:   3,
				Vbool:   true,
				Vfloat:  4.56,
				Vextra:  "vextra",
				Vdata:   []byte("data"),
				vsilent: true,
			},
			&BasicCopy{},
			&BasicCopy{
				Vstring: "vstring",
				Vint:    2,
				Vuint:   3,
				Vbool:   true,
				Vfloat:  4.56,
				Vextra:  "vextra",
				Vdata:   []byte("data"),
			},
			false,
		},
		{
			"struct => struct with pointers",
			&NestedPointer{
				Vfoo: "hello",
				Vbar: nil,
			},
			&NestedPointerCopy{},
			&NestedPointerCopy{
				Vfoo: "hello",
			},
			false,
		},
		{
			"basic pointer to non-pointer",
			&BasicPointer{
				Vstring: stringPtr("vstring"),
				Vint:    intPtr(2),
				Vuint:   uintPtr(3),
				Vbool:   boolPtr(true),
				Vfloat:  floatPtr(4.56),
				Vdata:   interfacePtr([]byte("data")),
			},
			&Basic{},
			&Basic{
				Vstring: "vstring",
				Vint:    2,
				Vuint:   3,
				Vbool:   true,
				Vfloat:  4.56,
				Vdata:   []byte("data"),
			},
			false,
		},
		{
			"slice non-pointer to pointer",
			&Slice{},
			&SlicePointer{},
			&SlicePointer{},
			false,
		},
		{
			"slice non-pointer to pointer, zero field",
			&Slice{},
			&SlicePointer{
				Vbar: &[]string{"yo"},
			},
			&SlicePointer{},
			false,
		},
		{
			"slice to slice alias",
			&Slice{},
			&SliceOfAlias{},
			&SliceOfAlias{},
			false,
		},
		{
			"nil map to map",
			&Map{},
			&MapCopy{},
			&MapCopy{},
			false,
		},
		{
			"nil map to non-empty map",
			&Map{},
			&MapCopy{Vother: map[string]string{"foo": "bar"}},
			&MapCopy{},
			false,
		},

		{
			"slice input - should error",
			[]string{"foo", "bar"},
			&map[string]any{},
			&map[string]any{},
			true,
		},
		{
			"struct with slice property",
			&Slice{
				Vfoo: "vfoo",
				Vbar: []string{"foo", "bar"},
			},
			&map[string]any{},
			&map[string]any{
				"Vfoo": "vfoo",
				"Vbar": []string{"foo", "bar"},
			},
			false,
		},
		{
			"struct with empty slice",
			&map[string]any{
				"Vbar": []string{},
			},
			&Slice{},
			&Slice{
				Vbar: []string{},
			},
			false,
		},
		{
			"struct with slice of struct property",
			&SliceOfStruct{
				Value: []Basic{
					{
						Vstring: "vstring",
						Vint:    2,
						Vuint:   3,
						Vbool:   true,
						Vfloat:  4.56,
						Vextra:  "vextra",
						vsilent: true,
						Vdata:   []byte("data"),
					},
				},
			},
			&map[string]any{},
			&map[string]any{
				"Value": []Basic{
					{
						Vstring: "vstring",
						Vint:    2,
						Vuint:   3,
						Vbool:   true,
						Vfloat:  4.56,
						Vextra:  "vextra",
						vsilent: true,
						Vdata:   []byte("data"),
					},
				},
			},
			false,
		},
		{
			"struct with map property",
			&Map{
				Vfoo:   "vfoo",
				Vother: map[string]string{"vother": "vother"},
			},
			&map[string]any{},
			&map[string]any{
				"Vfoo": "vfoo",
				"Vother": map[string]string{
					"vother": "vother",
				},
			},
			false,
		},
		{
			"tagged struct",
			&Tagged{
				Extra: "extra",
				Value: "value",
			},
			&map[string]string{},
			&map[string]string{
				"bar": "extra",
				"foo": "value",
			},
			false,
		},
		{
			"omit tag struct",
			&struct {
				Value string `mapstructure:"value"`
				Omit  string `mapstructure:"-"`
			}{
				Value: "value",
				Omit:  "omit",
			},
			&map[string]string{},
			&map[string]string{
				"value": "value",
			},
			false,
		},
		{
			"decode to wrong map type",
			&struct {
				Value string
			}{
				Value: "string",
			},
			&map[string]int{},
			&map[string]int{},
			true,
		},
		{
			"remainder",
			map[string]any{
				"A": "hello",
				"B": "goodbye",
				"C": "yo",
			},
			&Remainder{},
			&Remainder{
				A: "hello",
				Extra: map[string]any{
					"B": "goodbye",
					"C": "yo",
				},
			},
			false,
		},
		{
			"remainder with no extra",
			map[string]any{
				"A": "hello",
			},
			&Remainder{},
			&Remainder{
				A:     "hello",
				Extra: nil,
			},
			false,
		},
		{
			"struct with omitempty tag return non-empty values",
			&struct {
				VisibleField any `mapstructure:"visible"`
				OmitField    any `mapstructure:"omittable,omitempty"`
			}{
				VisibleField: nil,
				OmitField:    "string",
			},
			&map[string]any{},
			&map[string]any{"visible": nil, "omittable": "string"},
			false,
		},
		{
			"struct with omitempty tag ignore empty values",
			&struct {
				VisibleField any `mapstructure:"visible"`
				OmitField    any `mapstructure:"omittable,omitempty"`
			}{
				VisibleField: nil,
				OmitField:    nil,
			},
			&map[string]any{},
			&map[string]any{"visible": nil},
			false,
		},
		{
			"remainder with decode to map",
			&Remainder{
				A: "Alabasta",
				Extra: map[string]any{
					"B": "Baratie",
					"C": "Cocoyasi",
				},
			},
			&map[string]any{},
			&map[string]any{
				"A": "Alabasta",
				"B": "Baratie",
				"C": "Cocoyasi",
			},
			false,
		},
		{
			"remainder with decode to map with non-map field",
			&struct {
				A     string
				Extra *struct{} `mapstructure:",remain"`
			}{
				A:     "Alabasta",
				Extra: nil,
			},
			&map[string]any{},
			&map[string]any{
				"A": "Alabasta",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Decode(tt.in, tt.target); (err != nil) != tt.wantErr {
				t.Fatalf("%q: TestMapOutputForStructuredInputs() unexpected error: %s", tt.name, err)
			}

			if !reflect.DeepEqual(tt.out, tt.target) {
				t.Fatalf("%q: TestMapOutputForStructuredInputs() expected: %#v, got: %#v", tt.name, tt.out, tt.target)
			}
		})
	}
}

func TestInvalidType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": 42,
	}

	var result Basic
	err := Decode(input, &result)
	if err == nil {
		t.Fatal("error should exist")
	}

	var derr interface {
		Unwrap() []error
	}

	if !errors.As(err, &derr) {
		t.Fatalf("error should be a type implementing Unwrap() []error, instead: %#v", err)
	}

	errs := derr.Unwrap()

	var decoderErr *DecodeError
	if !errors.As(errs[0], &decoderErr) {
		t.Errorf("got unexpected error: %s", err)
	} else if errors.Is(decoderErr.Unwrap(), &UnconvertibleTypeError{}) {
		t.Errorf("error should be UnconvertibleTypeError, got: %s", decoderErr.Unwrap())
	}

	inputNegIntUint := map[string]any{
		"vuint": -42,
	}

	err = Decode(inputNegIntUint, &result)
	if err == nil {
		t.Fatal("error should exist")
	}

	if !errors.As(err, &derr) {
		t.Fatalf("error should be a type implementing Unwrap() []error, instead: %#v", err)
	}

	errs = derr.Unwrap()

	if !errors.As(errs[0], &decoderErr) {
		t.Errorf("got unexpected error: %s", err)
	} else if errors.Is(decoderErr.Unwrap(), &ParseError{}) {
		t.Errorf("error should be ParseError, got: %s", decoderErr.Unwrap())
	}

	inputNegFloatUint := map[string]any{
		"vuint": -42.0,
	}

	err = Decode(inputNegFloatUint, &result)
	if err == nil {
		t.Fatal("error should exist")
	}

	if !errors.As(err, &derr) {
		t.Fatalf("error should be a type implementing Unwrap() []error, instead: %#v", err)
	}

	errs = derr.Unwrap()

	if !errors.As(errs[0], &decoderErr) {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestDecodeMetadata(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vfoo": "foo",
		"vbar": map[string]any{
			"vstring": "foo",
			"Vuint":   42,
			"vsilent": "false",
			"foo":     "bar",
		},
		"bar": "nil",
	}

	var md Metadata
	var result Nested

	err := DecodeMetadata(input, &result, &md)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	expectedKeys := []string{"Vbar", "Vbar.Vstring", "Vbar.Vuint", "Vfoo"}
	sort.Strings(md.Keys)
	if !reflect.DeepEqual(md.Keys, expectedKeys) {
		t.Fatalf("bad keys: %#v", md.Keys)
	}

	expectedUnused := []string{"Vbar.foo", "Vbar.vsilent", "bar"}
	sort.Strings(md.Unused)
	if !reflect.DeepEqual(md.Unused, expectedUnused) {
		t.Fatalf("bad unused: %#v", md.Unused)
	}
}

func TestMetadata(t *testing.T) {
	t.Parallel()

	type testResult struct {
		Vfoo string
		Vbar BasicPointer
	}

	input := map[string]any{
		"vfoo": "foo",
		"vbar": map[string]any{
			"vstring": "foo",
			"Vuint":   42,
			"vsilent": "false",
			"foo":     "bar",
		},
		"bar": "nil",
	}

	var md Metadata
	var result testResult
	config := &DecoderConfig{
		Metadata: &md,
		Result:   &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	expectedKeys := []string{"Vbar", "Vbar.Vstring", "Vbar.Vuint", "Vfoo"}
	sort.Strings(md.Keys)
	if !reflect.DeepEqual(md.Keys, expectedKeys) {
		t.Fatalf("bad keys: %#v", md.Keys)
	}

	expectedUnused := []string{"Vbar.foo", "Vbar.vsilent", "bar"}
	sort.Strings(md.Unused)
	if !reflect.DeepEqual(md.Unused, expectedUnused) {
		t.Fatalf("bad unused: %#v", md.Unused)
	}

	expectedUnset := []string{
		"Vbar.Vbool", "Vbar.Vdata", "Vbar.Vextra", "Vbar.Vfloat", "Vbar.Vint",
		"Vbar.VjsonFloat", "Vbar.VjsonInt", "Vbar.VjsonNumber",
	}
	sort.Strings(md.Unset)
	if !reflect.DeepEqual(md.Unset, expectedUnset) {
		t.Fatalf("bad unset: %#v", md.Unset)
	}
}

func TestMetadata_Embedded(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"vstring": "foo",
		"vunique": "bar",
	}

	var md Metadata
	var result EmbeddedSquash
	config := &DecoderConfig{
		Metadata: &md,
		Result:   &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	expectedKeys := []string{"Vstring", "Vunique"}

	sort.Strings(md.Keys)
	if !reflect.DeepEqual(md.Keys, expectedKeys) {
		t.Fatalf("bad keys: %#v", md.Keys)
	}

	expectedUnused := []string{}
	if !reflect.DeepEqual(md.Unused, expectedUnused) {
		t.Fatalf("bad unused: %#v", md.Unused)
	}
}

func TestNonPtrValue(t *testing.T) {
	t.Parallel()

	err := Decode(map[string]any{}, Basic{})
	if err == nil {
		t.Fatal("error should exist")
	}

	if err.Error() != "result must be a pointer" {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestTagged(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"foo": "bar",
		"bar": "value",
	}

	var result Tagged
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Value != "bar" {
		t.Errorf("value should be 'bar', got: %#v", result.Value)
	}

	if result.Extra != "value" {
		t.Errorf("extra should be 'value', got: %#v", result.Extra)
	}
}

func TestWeakDecode(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"foo": "4",
		"bar": "value",
	}

	var result struct {
		Foo int
		Bar string
	}

	if err := WeakDecode(input, &result); err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.Foo != 4 {
		t.Fatalf("bad: %#v", result)
	}
	if result.Bar != "value" {
		t.Fatalf("bad: %#v", result)
	}
}

func TestWeakDecodeMetadata(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"foo":        "4",
		"bar":        "value",
		"unused":     "value",
		"unexported": "value",
	}

	var md Metadata
	var result struct {
		Foo        int
		Bar        string
		unexported string
	}

	if err := WeakDecodeMetadata(input, &result, &md); err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.Foo != 4 {
		t.Fatalf("bad: %#v", result)
	}
	if result.Bar != "value" {
		t.Fatalf("bad: %#v", result)
	}

	expectedKeys := []string{"Bar", "Foo"}
	sort.Strings(md.Keys)
	if !reflect.DeepEqual(md.Keys, expectedKeys) {
		t.Fatalf("bad keys: %#v", md.Keys)
	}

	expectedUnused := []string{"unexported", "unused"}
	sort.Strings(md.Unused)
	if !reflect.DeepEqual(md.Unused, expectedUnused) {
		t.Fatalf("bad unused: %#v", md.Unused)
	}
}

func TestDecode_StructTaggedWithOmitempty_OmitEmptyValues(t *testing.T) {
	t.Parallel()

	input := &StructWithOmitEmpty{}

	var emptySlice []any
	var emptyMap map[string]any
	var emptyNested *Nested
	expected := &map[string]any{
		"visible-string": "",
		"visible-int":    0,
		"visible-float":  0.0,
		"visible-slice":  emptySlice,
		"visible-map":    emptyMap,
		"visible-nested": emptyNested,
	}

	actual := &map[string]any{}
	Decode(input, actual)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Decode() expected: %#v, got: %#v", expected, actual)
	}
}

func TestDecode_StructTaggedWithOmitempty_KeepNonEmptyValues(t *testing.T) {
	t.Parallel()

	input := &StructWithOmitEmpty{
		VisibleStringField: "",
		OmitStringField:    "string",
		VisibleIntField:    0,
		OmitIntField:       1,
		VisibleFloatField:  0.0,
		OmitFloatField:     1.0,
		VisibleSliceField:  nil,
		OmitSliceField:     []any{1},
		VisibleMapField:    nil,
		OmitMapField:       map[string]any{"k": "v"},
		NestedField:        nil,
		OmitNestedField:    &Nested{},
	}

	var emptySlice []any
	var emptyMap map[string]any
	var emptyNested *Nested
	expected := &map[string]any{
		"visible-string":   "",
		"omittable-string": "string",
		"visible-int":      0,
		"omittable-int":    1,
		"visible-float":    0.0,
		"omittable-float":  1.0,
		"visible-slice":    emptySlice,
		"omittable-slice":  []any{1},
		"visible-map":      emptyMap,
		"omittable-map":    map[string]any{"k": "v"},
		"visible-nested":   emptyNested,
		"omittable-nested": &Nested{},
	}

	actual := &map[string]any{}
	Decode(input, actual)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Decode() expected: %#v, got: %#v", expected, actual)
	}
}

func TestDecode_StructTaggedWithOmitzero_KeepNonZeroValues(t *testing.T) {
	t.Parallel()

	input := &StructWithOmitZero{
		VisibleStringField: "",
		OmitStringField:    "string",
		VisibleIntField:    0,
		OmitIntField:       1,
		VisibleFloatField:  0.0,
		OmitFloatField:     1.0,
		VisibleSliceField:  nil,
		OmitSliceField:     []any{},
		VisibleMapField:    nil,
		OmitMapField:       map[string]any{},
		NestedField:        nil,
		OmitNestedField:    &Nested{},
	}

	var emptySlice []any
	var emptyMap map[string]any
	var emptyNested *Nested
	expected := &map[string]any{
		"visible-string":   "",
		"omittable-string": "string",
		"visible-int":      0,
		"omittable-int":    1,
		"visible-float":    0.0,
		"omittable-float":  1.0,
		"visible-slice":    emptySlice,
		"omittable-slice":  []any{},
		"visible-map":      emptyMap,
		"omittable-map":    map[string]any{},
		"visible-nested":   emptyNested,
		"omittable-nested": &Nested{},
	}

	actual := &map[string]any{}
	Decode(input, actual)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Decode() expected: %#v, got: %#v", expected, actual)
	}
}

func TestDecode_StructTaggedWithOmitzero_DropZeroValues(t *testing.T) {
	t.Parallel()

	input := &StructWithOmitZero{
		VisibleStringField: "",
		OmitStringField:    "",
		VisibleIntField:    0,
		OmitIntField:       0,
		VisibleFloatField:  0.0,
		OmitFloatField:     0.0,
		VisibleSliceField:  nil,
		OmitSliceField:     nil,
		VisibleMapField:    nil,
		OmitMapField:       nil,
		NestedField:        nil,
		OmitNestedField:    nil,
	}

	var emptySlice []any
	var emptyMap map[string]any
	var emptyNested *Nested
	expected := &map[string]any{
		"visible-string": "",
		"visible-int":    0,
		"visible-float":  0.0,
		"visible-slice":  emptySlice,
		"visible-map":    emptyMap,
		"visible-nested": emptyNested,
	}

	actual := &map[string]any{}
	Decode(input, actual)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Decode() expected: %#v, got: %#v", expected, actual)
	}
}

func TestDecode_mapToStruct(t *testing.T) {
	type Target struct {
		String    string
		StringPtr *string
	}

	expected := Target{
		String: "hello",
	}

	var target Target
	err := Decode(map[string]any{
		"string":    "hello",
		"StringPtr": "goodbye",
	}, &target)
	if err != nil {
		t.Fatalf("got error: %s", err)
	}

	// Pointers fail reflect test so do those manually
	if target.StringPtr == nil || *target.StringPtr != "goodbye" {
		t.Fatalf("bad: %#v", target)
	}
	target.StringPtr = nil

	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("bad: %#v", target)
	}
}

func TestDecoder_MatchName(t *testing.T) {
	t.Parallel()

	type Target struct {
		FirstMatch  string `mapstructure:"first_match"`
		SecondMatch string
		NoMatch     string `mapstructure:"no_match"`
	}

	input := map[string]any{
		"first_match": "foo",
		"SecondMatch": "bar",
		"NO_MATCH":    "baz",
	}

	expected := Target{
		FirstMatch:  "foo",
		SecondMatch: "bar",
	}

	var actual Target
	config := &DecoderConfig{
		Result: &actual,
		MatchName: func(mapKey, fieldName string) bool {
			return mapKey == fieldName
		},
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Decode() expected: %#v, got: %#v", expected, actual)
	}
}

func TestDecoder_IgnoreUntaggedFields(t *testing.T) {
	type Input struct {
		UntaggedNumber int
		TaggedNumber   int `mapstructure:"tagged_number"`
		UntaggedString string
		TaggedString   string `mapstructure:"tagged_string"`
	}
	input := &Input{
		UntaggedNumber: 31,
		TaggedNumber:   42,
		UntaggedString: "hidden",
		TaggedString:   "visible",
	}

	actual := make(map[string]any)
	config := &DecoderConfig{
		Result:               &actual,
		IgnoreUntaggedFields: true,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := map[string]any{
		"tagged_number": 42,
		"tagged_string": "visible",
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Decode() expected: %#v\ngot: %#v", expected, actual)
	}
}

func TestDecoder_IgnoreUntaggedFieldsWithStruct(t *testing.T) {
	type Output struct {
		UntaggedInt    int
		TaggedNumber   int `mapstructure:"tagged_number"`
		UntaggedString string
		TaggedString   string `mapstructure:"tagged_string"`
	}
	input := map[any]any{
		"untaggedint":     31,
		"tagged_number":   41,
		"untagged_string": "hidden",
		"tagged_string":   "visible",
	}
	actual := Output{}
	config := &DecoderConfig{
		Result:               &actual,
		IgnoreUntaggedFields: true,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Output{
		TaggedNumber: 41,
		TaggedString: "visible",
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Decode() expected: %#v\ngot: %#v", expected, actual)
	}
}

func TestDecoder_MultiTagInline(t *testing.T) {
	type Inner struct {
		A int `yaml:"a"`
	}

	type Wrap struct {
		Inner `yaml:",inline"`
	}

	input := map[string]any{"a": 1}
	var result Wrap

	dec, err := NewDecoder(&DecoderConfig{
		TagName:          "config,yaml",
		SquashTagOption:  "inline",
		WeaklyTypedInput: true,
		Result:           &result,
	})
	if err != nil {
		t.Fatalf("NewDecoder error: %v", err)
	}

	if err := dec.Decode(input); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if result.Inner.A != 1 {
		t.Fatalf("expected inline field A=1, got %d", result.Inner.A)
	}
}

func TestDecoder_MultiTagRemain(t *testing.T) {
	type Wrap struct {
		Known string         `yaml:"known"`
		Extra map[string]any `yaml:",remain"`
	}

	input := map[string]any{
		"known":  "ok",
		"extra1": "v1",
		"extra2": 2,
	}
	var result Wrap

	dec, err := NewDecoder(&DecoderConfig{
		TagName:          "config,yaml",
		WeaklyTypedInput: true,
		Result:           &result,
	})
	if err != nil {
		t.Fatalf("NewDecoder error: %v", err)
	}

	if err := dec.Decode(input); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if result.Known != "ok" {
		t.Fatalf("expected Known=ok, got %q", result.Known)
	}
	if result.Extra == nil || len(result.Extra) != 2 {
		t.Fatalf("expected Extra to contain 2 items, got %v", result.Extra)
	}
	if result.Extra["extra1"] != "v1" {
		t.Fatalf("expected extra1=v1, got %v", result.Extra["extra1"])
	}
}

func TestDecoder_MultiTagBasic(t *testing.T) {
	type Person struct {
		Name  string `yaml:"name"`
		Age   int    `json:"age"`
		Email string `config:"email_address"`
	}

	input := map[string]any{
		"name":          "Alice",
		"age":           30,
		"email_address": "alice@example.com",
	}
	var result Person

	dec, err := NewDecoder(&DecoderConfig{
		TagName: "yaml,json,config",
		Result:  &result,
	})
	if err != nil {
		t.Fatalf("NewDecoder error: %v", err)
	}

	if err := dec.Decode(input); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if result.Name != "Alice" {
		t.Fatalf("expected Name=Alice, got %q", result.Name)
	}
	if result.Age != 30 {
		t.Fatalf("expected Age=30, got %d", result.Age)
	}
	if result.Email != "alice@example.com" {
		t.Fatalf("expected Email=alice@example.com, got %q", result.Email)
	}
}

func TestDecoder_MultiTagPriority(t *testing.T) {
	// When both tags exist, the first tag name in the list takes precedence
	type Item struct {
		Value string `yaml:"yaml_value" json:"json_value"`
	}

	input := map[string]any{
		"yaml_value": "from_yaml",
		"json_value": "from_json",
	}

	// Test yaml,json order - should use yaml tag
	var result1 Item
	dec1, err := NewDecoder(&DecoderConfig{
		TagName: "yaml,json",
		Result:  &result1,
	})
	if err != nil {
		t.Fatalf("NewDecoder error: %v", err)
	}
	if err := dec1.Decode(input); err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if result1.Value != "from_yaml" {
		t.Fatalf("with yaml,json expected Value=from_yaml, got %q", result1.Value)
	}

	// Test json,yaml order - should use json tag
	var result2 Item
	dec2, err := NewDecoder(&DecoderConfig{
		TagName: "json,yaml",
		Result:  &result2,
	})
	if err != nil {
		t.Fatalf("NewDecoder error: %v", err)
	}
	if err := dec2.Decode(input); err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if result2.Value != "from_json" {
		t.Fatalf("with json,yaml expected Value=from_json, got %q", result2.Value)
	}
}

func TestDecoder_MultiTagWhitespace(t *testing.T) {
	type Person struct {
		Name string `yaml:"name"`
		Age  int    `json:"age"`
	}

	input := map[string]any{
		"name": "Bob",
		"age":  25,
	}
	var result Person

	// Test with whitespace around tag names
	dec, err := NewDecoder(&DecoderConfig{
		TagName: " yaml , json ",
		Result:  &result,
	})
	if err != nil {
		t.Fatalf("NewDecoder error: %v", err)
	}

	if err := dec.Decode(input); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if result.Name != "Bob" {
		t.Fatalf("expected Name=Bob, got %q", result.Name)
	}
	if result.Age != 25 {
		t.Fatalf("expected Age=25, got %d", result.Age)
	}
}

func TestDecoder_MultiTagEmptyNames(t *testing.T) {
	type Person struct {
		Name string `mapstructure:"name"`
	}

	input := map[string]any{
		"name": "Charlie",
	}

	tests := []struct {
		name    string
		tagName string
	}{
		{"leading comma", ",yaml"},
		{"trailing comma", "yaml,"},
		{"multiple commas", ",,yaml,,"},
		{"only commas", ",,,"},
		{"empty with spaces", " , , "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result Person
			dec, err := NewDecoder(&DecoderConfig{
				TagName: tc.tagName,
				Result:  &result,
			})
			if err != nil {
				t.Fatalf("NewDecoder error: %v", err)
			}

			if err := dec.Decode(input); err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			// With invalid/empty tag names, should fall back to mapstructure
			if result.Name != "Charlie" {
				t.Fatalf("expected Name=Charlie (fallback to mapstructure), got %q", result.Name)
			}
		})
	}
}

func TestDecoder_DecodeNilOption(t *testing.T) {
	t.Parallel()

	type Transformed struct {
		Message string
		When    string
	}

	helloHook := func(reflect.Type, reflect.Type, any) (any, error) {
		return Transformed{Message: "hello"}, nil
	}
	goodbyeHook := func(reflect.Type, reflect.Type, any) (any, error) {
		return Transformed{Message: "goodbye"}, nil
	}
	appendHook := func(from reflect.Value, to reflect.Value) (any, error) {
		if from.Kind() == reflect.Map {
			stringMap := from.Interface().(map[string]any)
			if stringMap == nil {
				stringMap = make(map[string]any)
			}
			stringMap["when"] = "see you later"
			return stringMap, nil
		}
		return from.Interface(), nil
	}

	tests := []struct {
		name           string
		decodeNil      bool
		input          any
		result         Transformed
		expectedResult Transformed
		decodeHook     DecodeHookFunc
	}{
		{
			name:           "decodeNil=true for nil input with hook",
			decodeNil:      true,
			input:          nil,
			decodeHook:     helloHook,
			expectedResult: Transformed{Message: "hello"},
		},
		{
			name:           "decodeNil=true for nil input without hook",
			decodeNil:      true,
			input:          nil,
			expectedResult: Transformed{Message: ""},
		},
		{
			name:           "decodeNil=false for nil input with hook",
			decodeNil:      false,
			input:          nil,
			decodeHook:     helloHook,
			expectedResult: Transformed{Message: ""},
		},
		{
			name:           "decodeNil=false for nil input without hook",
			decodeNil:      false,
			input:          nil,
			expectedResult: Transformed{Message: ""},
		},
		{
			name:           "decodeNil=true for non-nil input without hook",
			decodeNil:      true,
			input:          map[string]any{"message": "bar"},
			expectedResult: Transformed{Message: "bar"},
		},
		{
			name:           "decodeNil=true for non-nil input with hook",
			decodeNil:      true,
			input:          map[string]any{"message": "bar"},
			decodeHook:     goodbyeHook,
			expectedResult: Transformed{Message: "goodbye"},
		},
		{
			name:           "decodeNil=false for non-nil input without hook",
			decodeNil:      false,
			input:          map[string]any{"message": "bar"},
			expectedResult: Transformed{Message: "bar"},
		},
		{
			name:           "decodeNil=false for non-nil input with hook",
			decodeNil:      false,
			input:          map[string]any{"message": "bar"},
			decodeHook:     goodbyeHook,
			expectedResult: Transformed{Message: "goodbye"},
		},
		{
			name:           "decodeNil=true for nil input without hook and non-empty result",
			decodeNil:      true,
			input:          nil,
			result:         Transformed{Message: "foo"},
			expectedResult: Transformed{Message: "foo"},
		},
		{
			name:           "decodeNil=true for nil input with hook and non-empty result",
			decodeNil:      true,
			input:          nil,
			result:         Transformed{Message: "foo"},
			decodeHook:     helloHook,
			expectedResult: Transformed{Message: "hello"},
		},
		{
			name:           "decodeNil=false for nil input without hook and non-empty result",
			decodeNil:      false,
			input:          nil,
			result:         Transformed{Message: "foo"},
			expectedResult: Transformed{Message: "foo"},
		},
		{
			name:           "decodeNil=false for nil input with hook and non-empty result",
			decodeNil:      false,
			input:          nil,
			result:         Transformed{Message: "foo"},
			decodeHook:     helloHook,
			expectedResult: Transformed{Message: "foo"},
		},
		{
			name:           "decodeNil=false for non-nil input with hook that appends a value",
			decodeNil:      false,
			input:          map[string]any{"message": "bar"},
			decodeHook:     appendHook,
			expectedResult: Transformed{Message: "bar", When: "see you later"},
		},
		{
			name:           "decodeNil=true for non-nil input with hook that appends a value",
			decodeNil:      true,
			input:          map[string]any{"message": "bar"},
			decodeHook:     appendHook,
			expectedResult: Transformed{Message: "bar", When: "see you later"},
		},
		{
			name:           "decodeNil=true for nil input with hook that appends a value",
			decodeNil:      true,
			decodeHook:     appendHook,
			expectedResult: Transformed{When: "see you later"},
		},
		{
			name:           "decodeNil=false for nil input with hook that appends a value",
			decodeNil:      false,
			decodeHook:     appendHook,
			expectedResult: Transformed{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &DecoderConfig{
				Result:     &test.result,
				DecodeNil:  test.decodeNil,
				DecodeHook: test.decodeHook,
			}

			decoder, err := NewDecoder(config)
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			if err := decoder.Decode(test.input); err != nil {
				t.Fatalf("got an err: %s", err)
			}

			if test.result != test.expectedResult {
				t.Errorf("result should be: %#v, got %#v", test.expectedResult, test.result)
			}
		})
	}
}

func TestDecoder_ExpandNilStructPointersHookFunc(t *testing.T) {
	// a decoder hook that expands nil pointers in a struct to their zero value
	// if the input map contains the corresponding key.
	decodeHook := func(from reflect.Value, to reflect.Value) (any, error) {
		if from.Kind() == reflect.Map && to.Kind() == reflect.Map {
			toElem := to.Type().Elem()
			if toElem.Kind() == reflect.Ptr && toElem.Elem().Kind() == reflect.Struct {
				fromRange := from.MapRange()
				for fromRange.Next() {
					fromKey := fromRange.Key()
					fromValue := fromRange.Value()
					if fromValue.IsNil() {
						newFromValue := reflect.New(toElem.Elem())
						from.SetMapIndex(fromKey, newFromValue)
					}
				}
			}
		}
		return from.Interface(), nil
	}
	type Struct struct {
		Name string
	}
	type TestConfig struct {
		Boolean   *bool              `mapstructure:"boolean"`
		Struct    *Struct            `mapstructure:"struct"`
		MapStruct map[string]*Struct `mapstructure:"map_struct"`
	}
	stringMap := map[string]any{
		"boolean": nil,
		"struct":  nil,
		"map_struct": map[string]any{
			"struct": nil,
		},
	}
	var result TestConfig
	decoder, err := NewDecoder(&DecoderConfig{
		Result:     &result,
		DecodeNil:  true,
		DecodeHook: decodeHook,
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := decoder.Decode(stringMap); err != nil {
		t.Fatalf("got an err: %s", err)
	}
	if result.Boolean != nil {
		t.Errorf("nil Boolean expected, got '%#v'", result.Boolean)
	}
	if result.Struct != nil {
		t.Errorf("nil Struct expected, got '%#v'", result.Struct)
	}
	if len(result.MapStruct) == 0 {
		t.Fatalf("not-empty MapStruct expected, got '%#v'", result.MapStruct)
	}
	if _, ok := result.MapStruct["struct"]; !ok {
		t.Errorf("MapStruct['struct'] expected")
	}
}

func TestErrorLeakageWeakDecode(t *testing.T) {
	cases := []struct {
		value         any
		target        any
		allowNilError bool
	}{
		// case 0
		{"testing", new(bool), false},
		{"testing", new(int8), false},
		{"testing", new(uint8), false},
		{"testing", new(int16), false},
		{"testing", new(uint16), false},
		// case 5
		{"testing", new(int32), false},
		{"testing", new(uint32), false},
		{"testing", new(int64), false},
		{"testing", new(uint64), false},
		{"testing", new(int), false},
		// case 10
		{"testing", new(uint), false},
		{"testing", new(float32), false},
		{"testing", new(float64), false},
		{"testing", new(complex64), false},
		{"testing", new(complex128), false},
		// case 15
		{"testing", new(time.Duration), true},
		{"testing", new(time.Time), true},
		{map[string]string{"key": "secret-testing"}, new(map[string]int), false},
		{map[string]string{"key": "secret-testing"}, new(struct{ Key int }), false},
		{"secret-testing", new([]int), false},
		// case 20
		{"secret-testing", new([4]int), false},
	}

	for i, tc := range cases {
		err := WeakDecode(tc.value, tc.target)
		if err == nil {
			if tc.allowNilError {
				continue
			}

			t.Fatalf("case %d: expected error from input %v:\n\toutput (%T): %#v\n\toutput (string): %v", i, tc.value, tc.target, tc.target, tc.target)
		}

		strValue := fmt.Sprintf("%v", tc.value)
		if strings.Contains(strValue, "secret-testing") {
			strValue = "secret-testing"
		}

		if strings.Contains(err.Error(), strValue) {
			t.Errorf("case %d: error contains input value\n\terr: %v\n\tinput: %v", i, err, strValue)
		} else {
			t.Logf("case %d: got safe error: %v", i, err)
		}
	}
}

func testSliceInput(t *testing.T, input map[string]any, expected *Slice) {
	var result Slice
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got error: %s", err)
	}

	if result.Vfoo != expected.Vfoo {
		t.Errorf("Vfoo expected '%s', got '%s'", expected.Vfoo, result.Vfoo)
	}

	if result.Vbar == nil {
		t.Fatalf("Vbar a slice, got '%#v'", result.Vbar)
	}

	if len(result.Vbar) != len(expected.Vbar) {
		t.Errorf("Vbar length should be %d, got %d", len(expected.Vbar), len(result.Vbar))
	}

	for i, v := range result.Vbar {
		if v != expected.Vbar[i] {
			t.Errorf(
				"Vbar[%d] should be '%#v', got '%#v'",
				i, expected.Vbar[i], v)
		}
	}
}

func testArrayInput(t *testing.T, input map[string]any, expected *Array) {
	var result Array
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got error: %s", err)
	}

	if result.Vfoo != expected.Vfoo {
		t.Errorf("Vfoo expected '%s', got '%s'", expected.Vfoo, result.Vfoo)
	}

	if result.Vbar == [2]string{} {
		t.Fatalf("Vbar a slice, got '%#v'", result.Vbar)
	}

	if len(result.Vbar) != len(expected.Vbar) {
		t.Errorf("Vbar length should be %d, got %d", len(expected.Vbar), len(result.Vbar))
	}

	for i, v := range result.Vbar {
		if v != expected.Vbar[i] {
			t.Errorf(
				"Vbar[%d] should be '%#v', got '%#v'",
				i, expected.Vbar[i], v)
		}
	}
}

// Test struct for embedded error message testing
type TestDatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
}

type TestServerConfig struct {
	TestDatabaseConfig `mapstructure:",squash"`
	AppName            string `mapstructure:"app_name"`
	Debug              bool   `mapstructure:"debug"`
}

func TestDecoder_ErrorUnused_EmbeddedStruct_QualifiedTypeName(t *testing.T) {
	t.Parallel()

	// Input with an invalid key that should cause an error
	input := map[string]any{
		"host":        "localhost",
		"port":        5432,
		"username":    "admin",
		"app_name":    "myapp",
		"debug":       true,
		"invalid_key": "this should cause an error", // This key doesn't exist in the struct
	}

	var config TestServerConfig

	decoder, err := NewDecoder(&DecoderConfig{
		ErrorUnused: true, // Enable error on unused keys
		Result:      &config,
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err == nil {
		t.Fatal("expected error due to unused keys")
	}

	errorMessage := err.Error()
	t.Logf("Error message: %s", errorMessage)

	// Check that the error message contains the qualified struct type name
	if !strings.Contains(errorMessage, "'mapstructure.TestServerConfig'") {
		t.Errorf("Expected error message to contain qualified struct type 'mapstructure.TestServerConfig', got: %s", errorMessage)
	}

	// Check that unused keys are mentioned
	if !strings.Contains(errorMessage, "invalid_key") {
		t.Errorf("Expected error message to contain 'invalid_key', got: %s", errorMessage)
	}
}

func TestDecode_structArrayDeepMap(t *testing.T) {
	type SourceChild struct {
		String string `mapstructure:"field"`
	}

	type SourceParent struct {
		SE         []SourceChild                      `mapstructure:"s-e,deep"`
		SPE        *[]SourceChild                     `mapstructure:"s-p-e,deep"`
		SPPE       **[]SourceChild                    `mapstructure:"s-p-p-e,deep"`
		SEP        []*SourceChild                     `mapstructure:"s-e-p,deep"`
		SPEP       *[]*SourceChild                    `mapstructure:"s-p-e-p,deep"`
		SPPEP      **[]*SourceChild                   `mapstructure:"s-p-p-e-p,deep"`
		SEPP       []**SourceChild                    `mapstructure:"s-e-p-p,deep"`
		SPEPP      *[]**SourceChild                   `mapstructure:"s-p-e-p-p,deep"`
		SPPEPP     **[]**SourceChild                  `mapstructure:"s-p-p-e-p-p,deep"`
		AE         [1]SourceChild                     `mapstructure:"a-e,deep"`
		APE        *[1]SourceChild                    `mapstructure:"a-p-e,deep"`
		APPE       **[1]SourceChild                   `mapstructure:"a-p-p-e,deep"`
		AEP        [1]*SourceChild                    `mapstructure:"a-e-p,deep"`
		APEP       *[1]*SourceChild                   `mapstructure:"a-p-e-p,deep"`
		APPEP      **[1]*SourceChild                  `mapstructure:"a-p-p-e-p,deep"`
		AEPP       [1]**SourceChild                   `mapstructure:"a-e-p-p,deep"`
		APEPP      *[1]**SourceChild                  `mapstructure:"a-p-e-p-p,deep"`
		APPEPP     **[1]**SourceChild                 `mapstructure:"a-p-p-e-p-p,deep"`
		ME         map[string]SourceChild             `mapstructure:"m-e,deep"`
		MPE        *map[string]SourceChild            `mapstructure:"m-p-e,deep"`
		MPPE       **map[string]SourceChild           `mapstructure:"m-p-p-e,deep"`
		MEP        map[string]*SourceChild            `mapstructure:"m-e-p,deep"`
		MPEP       *map[string]*SourceChild           `mapstructure:"m-p-e-p,deep"`
		MPPEP      **map[string]*SourceChild          `mapstructure:"m-p-p-e-p,deep"`
		MEPP       map[string]**SourceChild           `mapstructure:"m-e-p-p,deep"`
		MPEPP      *map[string]**SourceChild          `mapstructure:"m-p-e-p-p,deep"`
		MPPEPP     **map[string]**SourceChild         `mapstructure:"m-p-p-e-p-p,deep"`
		MSE        map[string][]SourceChild           `mapstructure:"m-s-e,deep"`
		MPSE       *map[string][]SourceChild          `mapstructure:"m-p-s-e,deep"`
		MPPSE      **map[string][]SourceChild         `mapstructure:"m-p-p-s-e,deep"`
		MSPE       map[string]*[]SourceChild          `mapstructure:"m-s-p-e,deep"`
		MPSPE      *map[string]*[]SourceChild         `mapstructure:"m-p-s-p-e,deep"`
		MPPSPE     **map[string]*[]SourceChild        `mapstructure:"m-p-p-s-p-e,deep"`
		MSPPE      map[string]**[]SourceChild         `mapstructure:"m-s-p-p-e,deep"`
		MPSPPE     *map[string]**[]SourceChild        `mapstructure:"m-p-s-p-p-e,deep"`
		MPPSPPE    **map[string]**[]SourceChild       `mapstructure:"m-p-p-s-p-p-e,deep"`
		MAE        map[string][1]SourceChild          `mapstructure:"m-a-e,deep"`
		MPAE       *map[string][1]SourceChild         `mapstructure:"m-p-a-e,deep"`
		MPPAE      **map[string][1]SourceChild        `mapstructure:"m-p-p-a-e,deep"`
		MAPE       map[string]*[1]SourceChild         `mapstructure:"m-a-p-e,deep"`
		MPAPE      *map[string]*[1]SourceChild        `mapstructure:"m-p-a-p-e,deep"`
		MPPAPE     **map[string]*[1]SourceChild       `mapstructure:"m-p-p-a-p-e,deep"`
		MAPPE      map[string]**[1]SourceChild        `mapstructure:"m-a-p-p-e,deep"`
		MPAPPE     *map[string]**[1]SourceChild       `mapstructure:"m-p-a-p-p-e,deep"`
		MPPAPPE    **map[string]**[1]SourceChild      `mapstructure:"m-p-p-a-p-p-e,deep"`
		MSEP       map[string][]*SourceChild          `mapstructure:"m-s-e-p,deep"`
		MPSEP      *map[string][]*SourceChild         `mapstructure:"m-p-s-e-p,deep"`
		MPPSEP     **map[string][]*SourceChild        `mapstructure:"m-p-p-s-e-p,deep"`
		MSPEP      map[string]*[]*SourceChild         `mapstructure:"m-s-p-e-p,deep"`
		MPSPEP     *map[string]*[]*SourceChild        `mapstructure:"m-p-s-p-e-p,deep"`
		MPPSPEP    **map[string]*[]*SourceChild       `mapstructure:"m-p-p-s-p-e-p,deep"`
		MSPPEP     map[string]**[]*SourceChild        `mapstructure:"m-s-p-p-e-p,deep"`
		MPSPPEP    *map[string]**[]*SourceChild       `mapstructure:"m-p-s-p-p-e-p,deep"`
		MPPSPPEP   **map[string]**[]*SourceChild      `mapstructure:"m-p-p-s-p-p-e-p,deep"`
		MAEP       map[string][1]*SourceChild         `mapstructure:"m-a-e-p,deep"`
		MPAEP      *map[string][1]*SourceChild        `mapstructure:"m-p-a-e-p,deep"`
		MPPAEP     **map[string][1]*SourceChild       `mapstructure:"m-p-p-a-e-p,deep"`
		MAPEP      map[string]*[1]*SourceChild        `mapstructure:"m-a-p-e-p,deep"`
		MPAPEP     *map[string]*[1]*SourceChild       `mapstructure:"m-p-a-p-e-p,deep"`
		MPPAPEP    **map[string]*[1]*SourceChild      `mapstructure:"m-p-p-a-p-e-p,deep"`
		MAPPEP     map[string]**[1]*SourceChild       `mapstructure:"m-a-p-p-e-p,deep"`
		MPAPPEP    *map[string]**[1]*SourceChild      `mapstructure:"m-p-a-p-p-e-p,deep"`
		MPPAPPEP   **map[string]**[1]*SourceChild     `mapstructure:"m-p-p-a-p-p-e-p,deep"`
		MSEPP      map[string][]**SourceChild         `mapstructure:"m-s-e-p-p,deep"`
		MPSEPP     *map[string][]**SourceChild        `mapstructure:"m-p-s-e-p-p,deep"`
		MPPSEPP    **map[string][]**SourceChild       `mapstructure:"m-p-p-s-e-p-p,deep"`
		MSPEPP     map[string]*[]**SourceChild        `mapstructure:"m-s-p-e-p-p,deep"`
		MPSPEPP    *map[string]*[]**SourceChild       `mapstructure:"m-p-s-p-e-p-p,deep"`
		MPPSPEPP   **map[string]*[]**SourceChild      `mapstructure:"m-p-p-s-p-e-p-p,deep"`
		MSPPEPP    map[string]**[]**SourceChild       `mapstructure:"m-s-p-p-e-p-p,deep"`
		MPSPPEPP   *map[string]**[]**SourceChild      `mapstructure:"m-p-s-p-p-e-p-p,deep"`
		MPPSPPEPP  **map[string]**[]**SourceChild     `mapstructure:"m-p-p-s-p-p-e-p-p,deep"`
		MAEPP      map[string][1]**SourceChild        `mapstructure:"m-a-e-p-p,deep"`
		MPAEPP     *map[string][1]**SourceChild       `mapstructure:"m-p-a-e-p-p,deep"`
		MPPAEPP    **map[string][1]**SourceChild      `mapstructure:"m-p-p-a-e-p-p,deep"`
		MAPEPP     map[string]*[1]**SourceChild       `mapstructure:"m-a-p-e-p-p,deep"`
		MPAPEPP    *map[string]*[1]**SourceChild      `mapstructure:"m-p-a-p-e-p-p,deep"`
		MPPAPEPP   **map[string]*[1]**SourceChild     `mapstructure:"m-p-p-a-p-e-p-p,deep"`
		MAPPEPP    map[string]**[1]**SourceChild      `mapstructure:"m-a-p-p-e-p-p,deep"`
		MPAPPEPP   *map[string]**[1]**SourceChild     `mapstructure:"m-p-a-p-p-e-p-p,deep"`
		MPPAPPEPP  **map[string]**[1]**SourceChild    `mapstructure:"m-p-p-a-p-p-e-p-p,deep"`
		SEM        []map[string]any                   `mapstructure:"s-e-m,deep"`
		SPEM       *[]map[string]any                  `mapstructure:"s-p-e-m,deep"`
		SPPEM      **[]map[string]any                 `mapstructure:"s-p-p-e-m,deep"`
		SEPM       []*map[string]any                  `mapstructure:"s-e-p-m,deep"`
		SPEPM      *[]*map[string]any                 `mapstructure:"s-p-e-p-m,deep"`
		SPPEPM     **[]*map[string]any                `mapstructure:"s-p-p-e-p-m,deep"`
		SEPPM      []**map[string]any                 `mapstructure:"s-e-p-p-m,deep"`
		SPEPPM     *[]**map[string]any                `mapstructure:"s-p-e-p-p-m,deep"`
		SPPEPPM    **[]**map[string]any               `mapstructure:"s-p-p-e-p-p-m,deep"`
		AEM        [1]map[string]any                  `mapstructure:"a-e-m,deep"`
		APEM       *[1]map[string]any                 `mapstructure:"a-p-e-m,deep"`
		APPEM      **[1]map[string]any                `mapstructure:"a-p-p-e-m,deep"`
		AEPM       [1]*map[string]any                 `mapstructure:"a-e-p-m,deep"`
		APEPM      *[1]*map[string]any                `mapstructure:"a-p-e-p-m,deep"`
		APPEPM     **[1]*map[string]any               `mapstructure:"a-p-p-e-p-m,deep"`
		AEPPM      [1]**map[string]any                `mapstructure:"a-e-p-p-m,deep"`
		APEPPM     *[1]**map[string]any               `mapstructure:"a-p-e-p-p-m,deep"`
		APPEPPM    **[1]**map[string]any              `mapstructure:"a-p-p-e-p-p-m,deep"`
		MEM        map[string]map[string]any          `mapstructure:"m-e-m,deep"`
		MPEM       *map[string]map[string]any         `mapstructure:"m-p-e-m,deep"`
		MPPEM      **map[string]map[string]any        `mapstructure:"m-p-p-e-m,deep"`
		MEPM       map[string]*map[string]any         `mapstructure:"m-e-p-m,deep"`
		MPEPM      *map[string]*map[string]any        `mapstructure:"m-p-e-p-m,deep"`
		MPPEPM     **map[string]*map[string]any       `mapstructure:"m-p-p-e-p-m,deep"`
		MEPPM      map[string]**map[string]any        `mapstructure:"m-e-p-p-m,deep"`
		MPEPPM     *map[string]**map[string]any       `mapstructure:"m-p-e-p-p-m,deep"`
		MPPEPPM    **map[string]**map[string]any      `mapstructure:"m-p-p-e-p-p-m,deep"`
		MSEM       map[string][]map[string]any        `mapstructure:"m-s-e-m,deep"`
		MPSEM      *map[string][]map[string]any       `mapstructure:"m-p-s-e-m,deep"`
		MPPSEM     **map[string][]map[string]any      `mapstructure:"m-p-p-s-e-m,deep"`
		MSPEM      map[string]*[]map[string]any       `mapstructure:"m-s-p-e-m,deep"`
		MPSPEM     *map[string]*[]map[string]any      `mapstructure:"m-p-s-p-e-m,deep"`
		MPPSPEM    **map[string]*[]map[string]any     `mapstructure:"m-p-p-s-p-e-m,deep"`
		MSPPEM     map[string]**[]map[string]any      `mapstructure:"m-s-p-p-e-m,deep"`
		MPSPPEM    *map[string]**[]map[string]any     `mapstructure:"m-p-s-p-p-e-m,deep"`
		MPPSPPEM   **map[string]**[]map[string]any    `mapstructure:"m-p-p-s-p-p-e-m,deep"`
		MAEM       map[string][1]map[string]any       `mapstructure:"m-a-e-m,deep"`
		MPAEM      *map[string][1]map[string]any      `mapstructure:"m-p-a-e-m,deep"`
		MPPAEM     **map[string][1]map[string]any     `mapstructure:"m-p-p-a-e-m,deep"`
		MAPEM      map[string]*[1]map[string]any      `mapstructure:"m-a-p-e-m,deep"`
		MPAPEM     *map[string]*[1]map[string]any     `mapstructure:"m-p-a-p-e-m,deep"`
		MPPAPEM    **map[string]*[1]map[string]any    `mapstructure:"m-p-p-a-p-e-m,deep"`
		MAPPEM     map[string]**[1]map[string]any     `mapstructure:"m-a-p-p-e-m,deep"`
		MPAPPEM    *map[string]**[1]map[string]any    `mapstructure:"m-p-a-p-p-e-m,deep"`
		MPPAPPEM   **map[string]**[1]map[string]any   `mapstructure:"m-p-p-a-p-p-e-m,deep"`
		MSEPM      map[string][]*map[string]any       `mapstructure:"m-s-e-p-m,deep"`
		MPSEPM     *map[string][]*map[string]any      `mapstructure:"m-p-s-e-p-m,deep"`
		MPPSEPM    **map[string][]*map[string]any     `mapstructure:"m-p-p-s-e-p-m,deep"`
		MSPEPM     map[string]*[]*map[string]any      `mapstructure:"m-s-p-e-p-m,deep"`
		MPSPEPM    *map[string]*[]*map[string]any     `mapstructure:"m-p-s-p-e-p-m,deep"`
		MPPSPEPM   **map[string]*[]*map[string]any    `mapstructure:"m-p-p-s-p-e-p-m,deep"`
		MSPPEPM    map[string]**[]*map[string]any     `mapstructure:"m-s-p-p-e-p-m,deep"`
		MPSPPEPM   *map[string]**[]*map[string]any    `mapstructure:"m-p-s-p-p-e-p-m,deep"`
		MPPSPPEPM  **map[string]**[]*map[string]any   `mapstructure:"m-p-p-s-p-p-e-p-m,deep"`
		MAEPM      map[string][1]*map[string]any      `mapstructure:"m-a-e-p-m,deep"`
		MPAEPM     *map[string][1]*map[string]any     `mapstructure:"m-p-a-e-p-m,deep"`
		MPPAEPM    **map[string][1]*map[string]any    `mapstructure:"m-p-p-a-e-p-m,deep"`
		MAPEPM     map[string]*[1]*map[string]any     `mapstructure:"m-a-p-e-p-m,deep"`
		MPAPEPM    *map[string]*[1]*map[string]any    `mapstructure:"m-p-a-p-e-p-m,deep"`
		MPPAPEPM   **map[string]*[1]*map[string]any   `mapstructure:"m-p-p-a-p-e-p-m,deep"`
		MAPPEPM    map[string]**[1]*map[string]any    `mapstructure:"m-a-p-p-e-p-m,deep"`
		MPAPPEPM   *map[string]**[1]*map[string]any   `mapstructure:"m-p-a-p-p-e-p-m,deep"`
		MPPAPPEPM  **map[string]**[1]*map[string]any  `mapstructure:"m-p-p-a-p-p-e-p-m,deep"`
		MSEPPM     map[string][]**map[string]any      `mapstructure:"m-s-e-p-p-m,deep"`
		MPSEPPM    *map[string][]**map[string]any     `mapstructure:"m-p-s-e-p-p-m,deep"`
		MPPSEPPM   **map[string][]**map[string]any    `mapstructure:"m-p-p-s-e-p-p-m,deep"`
		MSPEPPM    map[string]*[]**map[string]any     `mapstructure:"m-s-p-e-p-p-m,deep"`
		MPSPEPPM   *map[string]*[]**map[string]any    `mapstructure:"m-p-s-p-e-p-p-m,deep"`
		MPPSPEPPM  **map[string]*[]**map[string]any   `mapstructure:"m-p-p-s-p-e-p-p-m,deep"`
		MSPPEPPM   map[string]**[]**map[string]any    `mapstructure:"m-s-p-p-e-p-p-m,deep"`
		MPSPPEPPM  *map[string]**[]**map[string]any   `mapstructure:"m-p-s-p-p-e-p-p-m,deep"`
		MPPSPPEPPM **map[string]**[]**map[string]any  `mapstructure:"m-p-p-s-p-p-e-p-p-m,deep"`
		MAEPPM     map[string][1]**map[string]any     `mapstructure:"m-a-e-p-p-m,deep"`
		MPAEPPM    *map[string][1]**map[string]any    `mapstructure:"m-p-a-e-p-p-m,deep"`
		MPPAEPPM   **map[string][1]**map[string]any   `mapstructure:"m-p-p-a-e-p-p-m,deep"`
		MAPEPPM    map[string]*[1]**map[string]any    `mapstructure:"m-a-p-e-p-p-m,deep"`
		MPAPEPPM   *map[string]*[1]**map[string]any   `mapstructure:"m-p-a-p-e-p-p-m,deep"`
		MPPAPEPPM  **map[string]*[1]**map[string]any  `mapstructure:"m-p-p-a-p-e-p-p-m,deep"`
		MAPPEPPM   map[string]**[1]**map[string]any   `mapstructure:"m-a-p-p-e-p-p-m,deep"`
		MPAPPEPPM  *map[string]**[1]**map[string]any  `mapstructure:"m-p-a-p-p-e-p-p-m,deep"`
		MPPAPPEPPM **map[string]**[1]**map[string]any `mapstructure:"m-p-p-a-p-p-e-p-p-m,deep"`
	}

	e := SourceChild{String: "el"}
	ep := &e
	epp := &ep

	em := map[string]any{"field": "el"}
	epm := &em
	eppm := &epm

	se := []SourceChild{e}
	spe := &se
	sppe := &spe

	sem := []map[string]any{em}
	spem := &sem
	sppem := &spem

	sep := []*SourceChild{ep}
	spep := &sep
	sppep := &spep

	sepm := []*map[string]any{epm}
	spepm := &sepm
	sppepm := &spepm

	sepp := []**SourceChild{epp}
	spepp := &sepp
	sppepp := &spepp

	seppm := []**map[string]any{eppm}
	speppm := &seppm
	sppeppm := &speppm

	ae := [1]SourceChild{e}
	ape := &ae
	appe := &ape

	aem := [1]map[string]any{em}
	apem := &aem
	appem := &apem

	aep := [1]*SourceChild{ep}
	apep := &aep
	appep := &apep

	aepm := [1]*map[string]any{epm}
	apepm := &aepm
	appepm := &apepm

	aepp := [1]**SourceChild{epp}
	apepp := &aepp
	appepp := &apepp

	aeppm := [1]**map[string]any{eppm}
	apeppm := &aeppm
	appeppm := &apeppm

	me := map[string]SourceChild{"foo": e}
	mpe := &me
	mppe := &mpe

	mem := map[string]map[string]any{"foo": em}
	mpem := &mem
	mppem := &mpem

	mep := map[string]*SourceChild{"foo": ep}
	mpep := &mep
	mppep := &mpep

	mepm := map[string]*map[string]any{"foo": epm}
	mpepm := &mepm
	mppepm := &mpepm

	mepp := map[string]**SourceChild{"foo": epp}
	mpepp := &mepp
	mppepp := &mpepp

	meppm := map[string]**map[string]any{"foo": eppm}
	mpeppm := &meppm
	mppeppm := &mpeppm

	mse := map[string][]SourceChild{"foo": se}
	mpse := &mse
	mppse := &mpse

	msem := map[string][]map[string]any{"foo": sem}
	mpsem := &msem
	mppsem := &mpsem

	mspe := map[string]*[]SourceChild{"foo": spe}
	mpspe := &mspe
	mppspe := &mpspe

	mspem := map[string]*[]map[string]any{"foo": spem}
	mpspem := &mspem
	mppspem := &mpspem

	msppe := map[string]**[]SourceChild{"foo": sppe}
	mpsppe := &msppe
	mppsppe := &mpsppe

	msppem := map[string]**[]map[string]any{"foo": sppem}
	mpsppem := &msppem
	mppsppem := &mpsppem

	mae := map[string][1]SourceChild{"foo": ae}
	mpae := &mae
	mppae := &mpae

	maem := map[string][1]map[string]any{"foo": aem}
	mpaem := &maem
	mppaem := &mpaem

	mape := map[string]*[1]SourceChild{"foo": ape}
	mpape := &mape
	mppape := &mpape

	mapem := map[string]*[1]map[string]any{"foo": apem}
	mpapem := &mapem
	mppapem := &mpapem

	mappe := map[string]**[1]SourceChild{"foo": appe}
	mpappe := &mappe
	mppappe := &mpappe

	mappem := map[string]**[1]map[string]any{"foo": appem}
	mpappem := &mappem
	mppappem := &mpappem

	msep := map[string][]*SourceChild{"foo": sep}
	mpsep := &msep
	mppsep := &mpsep

	msepm := map[string][]*map[string]any{"foo": sepm}
	mpsepm := &msepm
	mppsepm := &mpsepm

	mspep := map[string]*[]*SourceChild{"foo": spep}
	mpspep := &mspep
	mppspep := &mpspep

	mspepm := map[string]*[]*map[string]any{"foo": spepm}
	mpspepm := &mspepm
	mppspepm := &mpspepm

	msppep := map[string]**[]*SourceChild{"foo": sppep}
	mpsppep := &msppep
	mppsppep := &mpsppep

	msppepm := map[string]**[]*map[string]any{"foo": sppepm}
	mpsppepm := &msppepm
	mppsppepm := &mpsppepm

	maep := map[string][1]*SourceChild{"foo": aep}
	mpaep := &maep
	mppaep := &mpaep

	maepm := map[string][1]*map[string]any{"foo": aepm}
	mpaepm := &maepm
	mppaepm := &mpaepm

	mapep := map[string]*[1]*SourceChild{"foo": apep}
	mpapep := &mapep
	mppapep := &mpapep

	mapepm := map[string]*[1]*map[string]any{"foo": apepm}
	mpapepm := &mapepm
	mppapepm := &mpapepm

	mappep := map[string]**[1]*SourceChild{"foo": appep}
	mpappep := &mappep
	mppappep := &mpappep

	mappepm := map[string]**[1]*map[string]any{"foo": appepm}
	mpappepm := &mappepm
	mppappepm := &mpappepm

	msepp := map[string][]**SourceChild{"foo": sepp}
	mpsepp := &msepp
	mppsepp := &mpsepp

	mseppm := map[string][]**map[string]any{"foo": seppm}
	mpseppm := &mseppm
	mppseppm := &mpseppm

	mspepp := map[string]*[]**SourceChild{"foo": spepp}
	mpspepp := &mspepp
	mppspepp := &mpspepp

	mspeppm := map[string]*[]**map[string]any{"foo": speppm}
	mpspeppm := &mspeppm
	mppspeppm := &mpspeppm

	msppepp := map[string]**[]**SourceChild{"foo": sppepp}
	mpsppepp := &msppepp
	mppsppepp := &mpsppepp

	msppeppm := map[string]**[]**map[string]any{"foo": sppeppm}
	mpsppeppm := &msppeppm
	mppsppeppm := &mpsppeppm

	maepp := map[string][1]**SourceChild{"foo": aepp}
	mpaepp := &maepp
	mppaepp := &mpaepp

	maeppm := map[string][1]**map[string]any{"foo": aeppm}
	mpaeppm := &maeppm
	mppaeppm := &mpaeppm

	mapepp := map[string]*[1]**SourceChild{"foo": apepp}
	mpapepp := &mapepp
	mppapepp := &mpapepp

	mapeppm := map[string]*[1]**map[string]any{"foo": apeppm}
	mpapeppm := &mapeppm
	mppapeppm := &mpapeppm

	mappepp := map[string]**[1]**SourceChild{"foo": appepp}
	mpappepp := &mappepp
	mppappepp := &mpappepp

	mappeppm := map[string]**[1]**map[string]any{"foo": appeppm}
	mpappeppm := &mappeppm
	mppappeppm := &mpappeppm

	var target map[string]any

	source := SourceParent{
		SE:         se,
		SPE:        spe,
		SPPE:       sppe,
		SEP:        sep,
		SPEP:       spep,
		SPPEP:      sppep,
		SEPP:       sepp,
		SPEPP:      spepp,
		SPPEPP:     sppepp,
		AE:         ae,
		APE:        ape,
		APPE:       appe,
		AEP:        aep,
		APEP:       apep,
		APPEP:      appep,
		AEPP:       aepp,
		APEPP:      apepp,
		APPEPP:     appepp,
		ME:         me,
		MPE:        mpe,
		MPPE:       mppe,
		MEP:        mep,
		MPEP:       mpep,
		MPPEP:      mppep,
		MEPP:       mepp,
		MPEPP:      mpepp,
		MPPEPP:     mppepp,
		MSE:        mse,
		MPSE:       mpse,
		MPPSE:      mppse,
		MSPE:       mspe,
		MPSPE:      mpspe,
		MPPSPE:     mppspe,
		MSPPE:      msppe,
		MPSPPE:     mpsppe,
		MPPSPPE:    mppsppe,
		MAE:        mae,
		MPAE:       mpae,
		MPPAE:      mppae,
		MAPE:       mape,
		MPAPE:      mpape,
		MPPAPE:     mppape,
		MAPPE:      mappe,
		MPAPPE:     mpappe,
		MPPAPPE:    mppappe,
		MSEP:       msep,
		MPSEP:      mpsep,
		MPPSEP:     mppsep,
		MSPEP:      mspep,
		MPSPEP:     mpspep,
		MPPSPEP:    mppspep,
		MSPPEP:     msppep,
		MPSPPEP:    mpsppep,
		MPPSPPEP:   mppsppep,
		MAEP:       maep,
		MPAEP:      mpaep,
		MPPAEP:     mppaep,
		MAPEP:      mapep,
		MPAPEP:     mpapep,
		MPPAPEP:    mppapep,
		MAPPEP:     mappep,
		MPAPPEP:    mpappep,
		MPPAPPEP:   mppappep,
		MSEPP:      msepp,
		MPSEPP:     mpsepp,
		MPPSEPP:    mppsepp,
		MSPEPP:     mspepp,
		MPSPEPP:    mpspepp,
		MPPSPEPP:   mppspepp,
		MSPPEPP:    msppepp,
		MPSPPEPP:   mpsppepp,
		MPPSPPEPP:  mppsppepp,
		MAEPP:      maepp,
		MPAEPP:     mpaepp,
		MPPAEPP:    mppaepp,
		MAPEPP:     mapepp,
		MPAPEPP:    mpapepp,
		MPPAPEPP:   mppapepp,
		MAPPEPP:    mappepp,
		MPAPPEPP:   mpappepp,
		MPPAPPEPP:  mppappepp,
		SEM:        sem,
		SPEM:       spem,
		SPPEM:      sppem,
		SEPM:       sepm,
		SPEPM:      spepm,
		SPPEPM:     sppepm,
		SEPPM:      seppm,
		SPEPPM:     speppm,
		SPPEPPM:    sppeppm,
		AEM:        aem,
		APEM:       apem,
		APPEM:      appem,
		AEPM:       aepm,
		APEPM:      apepm,
		APPEPM:     appepm,
		AEPPM:      aeppm,
		APEPPM:     apeppm,
		APPEPPM:    appeppm,
		MEM:        mem,
		MPEM:       mpem,
		MPPEM:      mppem,
		MEPM:       mepm,
		MPEPM:      mpepm,
		MPPEPM:     mppepm,
		MEPPM:      meppm,
		MPEPPM:     mpeppm,
		MPPEPPM:    mppeppm,
		MSEM:       msem,
		MPSEM:      mpsem,
		MPPSEM:     mppsem,
		MSPEM:      mspem,
		MPSPEM:     mpspem,
		MPPSPEM:    mppspem,
		MSPPEM:     msppem,
		MPSPPEM:    mpsppem,
		MPPSPPEM:   mppsppem,
		MAEM:       maem,
		MPAEM:      mpaem,
		MPPAEM:     mppaem,
		MAPEM:      mapem,
		MPAPEM:     mpapem,
		MPPAPEM:    mppapem,
		MAPPEM:     mappem,
		MPAPPEM:    mpappem,
		MPPAPPEM:   mppappem,
		MSEPM:      msepm,
		MPSEPM:     mpsepm,
		MPPSEPM:    mppsepm,
		MSPEPM:     mspepm,
		MPSPEPM:    mpspepm,
		MPPSPEPM:   mppspepm,
		MSPPEPM:    msppepm,
		MPSPPEPM:   mpsppepm,
		MPPSPPEPM:  mppsppepm,
		MAEPM:      maepm,
		MPAEPM:     mpaepm,
		MPPAEPM:    mppaepm,
		MAPEPM:     mapepm,
		MPAPEPM:    mpapepm,
		MPPAPEPM:   mppapepm,
		MAPPEPM:    mappepm,
		MPAPPEPM:   mpappepm,
		MPPAPPEPM:  mppappepm,
		MSEPPM:     mseppm,
		MPSEPPM:    mpseppm,
		MPPSEPPM:   mppseppm,
		MSPEPPM:    mspeppm,
		MPSPEPPM:   mpspeppm,
		MPPSPEPPM:  mppspeppm,
		MSPPEPPM:   msppeppm,
		MPSPPEPPM:  mpsppeppm,
		MPPSPPEPPM: mppsppeppm,
		MAEPPM:     maeppm,
		MPAEPPM:    mpaeppm,
		MPPAEPPM:   mppaeppm,
		MAPEPPM:    mapeppm,
		MPAPEPPM:   mpapeppm,
		MPPAPEPPM:  mppapeppm,
		MAPPEPPM:   mappeppm,
		MPAPPEPPM:  mpappeppm,
		MPPAPPEPPM: mppappeppm,
	}

	if err := Decode(source, &target); err != nil {
		t.Fatalf("got error: %s", err)
	}

	sliceValue := []map[string]any{{"field": "el"}}
	arrayValue := [1]map[string]any{{"field": "el"}}
	mapValue := map[string]map[string]any{"foo": {"field": "el"}}
	mapSliceValue := map[string][]map[string]any{"foo": {{"field": "el"}}}
	mapArrayValue := map[string][1]map[string]any{"foo": {{"field": "el"}}}

	expected := map[string]any{
		"s-e":                 sliceValue,
		"s-p-e":               sliceValue,
		"s-p-p-e":             sliceValue,
		"s-e-p":               sliceValue,
		"s-p-e-p":             sliceValue,
		"s-p-p-e-p":           sliceValue,
		"s-e-p-p":             sliceValue,
		"s-p-e-p-p":           sliceValue,
		"s-p-p-e-p-p":         sliceValue,
		"a-e":                 arrayValue,
		"a-p-e":               arrayValue,
		"a-p-p-e":             arrayValue,
		"a-e-p":               arrayValue,
		"a-p-e-p":             arrayValue,
		"a-p-p-e-p":           arrayValue,
		"a-e-p-p":             arrayValue,
		"a-p-e-p-p":           arrayValue,
		"a-p-p-e-p-p":         arrayValue,
		"m-e":                 mapValue,
		"m-p-e":               mapValue,
		"m-p-p-e":             mapValue,
		"m-e-p":               mapValue,
		"m-p-e-p":             mapValue,
		"m-p-p-e-p":           mapValue,
		"m-e-p-p":             mapValue,
		"m-p-e-p-p":           mapValue,
		"m-p-p-e-p-p":         mapValue,
		"m-s-e":               mapSliceValue,
		"m-p-s-e":             mapSliceValue,
		"m-p-p-s-e":           mapSliceValue,
		"m-s-p-e":             mapSliceValue,
		"m-p-s-p-e":           mapSliceValue,
		"m-p-p-s-p-e":         mapSliceValue,
		"m-s-p-p-e":           mapSliceValue,
		"m-p-s-p-p-e":         mapSliceValue,
		"m-p-p-s-p-p-e":       mapSliceValue,
		"m-a-e":               mapArrayValue,
		"m-p-a-e":             mapArrayValue,
		"m-p-p-a-e":           mapArrayValue,
		"m-a-p-e":             mapArrayValue,
		"m-p-a-p-e":           mapArrayValue,
		"m-p-p-a-p-e":         mapArrayValue,
		"m-a-p-p-e":           mapArrayValue,
		"m-p-a-p-p-e":         mapArrayValue,
		"m-p-p-a-p-p-e":       mapArrayValue,
		"m-s-e-p":             mapSliceValue,
		"m-p-s-e-p":           mapSliceValue,
		"m-p-p-s-e-p":         mapSliceValue,
		"m-s-p-e-p":           mapSliceValue,
		"m-p-s-p-e-p":         mapSliceValue,
		"m-p-p-s-p-e-p":       mapSliceValue,
		"m-s-p-p-e-p":         mapSliceValue,
		"m-p-s-p-p-e-p":       mapSliceValue,
		"m-p-p-s-p-p-e-p":     mapSliceValue,
		"m-a-e-p":             mapArrayValue,
		"m-p-a-e-p":           mapArrayValue,
		"m-p-p-a-e-p":         mapArrayValue,
		"m-a-p-e-p":           mapArrayValue,
		"m-p-a-p-e-p":         mapArrayValue,
		"m-p-p-a-p-e-p":       mapArrayValue,
		"m-a-p-p-e-p":         mapArrayValue,
		"m-p-a-p-p-e-p":       mapArrayValue,
		"m-p-p-a-p-p-e-p":     mapArrayValue,
		"m-s-e-p-p":           mapSliceValue,
		"m-p-s-e-p-p":         mapSliceValue,
		"m-p-p-s-e-p-p":       mapSliceValue,
		"m-s-p-e-p-p":         mapSliceValue,
		"m-p-s-p-e-p-p":       mapSliceValue,
		"m-p-p-s-p-e-p-p":     mapSliceValue,
		"m-s-p-p-e-p-p":       mapSliceValue,
		"m-p-s-p-p-e-p-p":     mapSliceValue,
		"m-p-p-s-p-p-e-p-p":   mapSliceValue,
		"m-a-e-p-p":           mapArrayValue,
		"m-p-a-e-p-p":         mapArrayValue,
		"m-p-p-a-e-p-p":       mapArrayValue,
		"m-a-p-e-p-p":         mapArrayValue,
		"m-p-a-p-e-p-p":       mapArrayValue,
		"m-p-p-a-p-e-p-p":     mapArrayValue,
		"m-a-p-p-e-p-p":       mapArrayValue,
		"m-p-a-p-p-e-p-p":     mapArrayValue,
		"m-p-p-a-p-p-e-p-p":   mapArrayValue,
		"s-e-m":               sliceValue,
		"s-p-e-m":             sliceValue,
		"s-p-p-e-m":           sliceValue,
		"s-e-p-m":             sliceValue,
		"s-p-e-p-m":           sliceValue,
		"s-p-p-e-p-m":         sliceValue,
		"s-e-p-p-m":           sliceValue,
		"s-p-e-p-p-m":         sliceValue,
		"s-p-p-e-p-p-m":       sliceValue,
		"a-e-m":               arrayValue,
		"a-p-e-m":             arrayValue,
		"a-p-p-e-m":           arrayValue,
		"a-e-p-m":             arrayValue,
		"a-p-e-p-m":           arrayValue,
		"a-p-p-e-p-m":         arrayValue,
		"a-e-p-p-m":           arrayValue,
		"a-p-e-p-p-m":         arrayValue,
		"a-p-p-e-p-p-m":       arrayValue,
		"m-e-m":               mapValue,
		"m-p-e-m":             mapValue,
		"m-p-p-e-m":           mapValue,
		"m-e-p-m":             mapValue,
		"m-p-e-p-m":           mapValue,
		"m-p-p-e-p-m":         mapValue,
		"m-e-p-p-m":           mapValue,
		"m-p-e-p-p-m":         mapValue,
		"m-p-p-e-p-p-m":       mapValue,
		"m-s-e-m":             mapSliceValue,
		"m-p-s-e-m":           mapSliceValue,
		"m-p-p-s-e-m":         mapSliceValue,
		"m-s-p-e-m":           mapSliceValue,
		"m-p-s-p-e-m":         mapSliceValue,
		"m-p-p-s-p-e-m":       mapSliceValue,
		"m-s-p-p-e-m":         mapSliceValue,
		"m-p-s-p-p-e-m":       mapSliceValue,
		"m-p-p-s-p-p-e-m":     mapSliceValue,
		"m-a-e-m":             mapArrayValue,
		"m-p-a-e-m":           mapArrayValue,
		"m-p-p-a-e-m":         mapArrayValue,
		"m-a-p-e-m":           mapArrayValue,
		"m-p-a-p-e-m":         mapArrayValue,
		"m-p-p-a-p-e-m":       mapArrayValue,
		"m-a-p-p-e-m":         mapArrayValue,
		"m-p-a-p-p-e-m":       mapArrayValue,
		"m-p-p-a-p-p-e-m":     mapArrayValue,
		"m-s-e-p-m":           mapSliceValue,
		"m-p-s-e-p-m":         mapSliceValue,
		"m-p-p-s-e-p-m":       mapSliceValue,
		"m-s-p-e-p-m":         mapSliceValue,
		"m-p-s-p-e-p-m":       mapSliceValue,
		"m-p-p-s-p-e-p-m":     mapSliceValue,
		"m-s-p-p-e-p-m":       mapSliceValue,
		"m-p-s-p-p-e-p-m":     mapSliceValue,
		"m-p-p-s-p-p-e-p-m":   mapSliceValue,
		"m-a-e-p-m":           mapArrayValue,
		"m-p-a-e-p-m":         mapArrayValue,
		"m-p-p-a-e-p-m":       mapArrayValue,
		"m-a-p-e-p-m":         mapArrayValue,
		"m-p-a-p-e-p-m":       mapArrayValue,
		"m-p-p-a-p-e-p-m":     mapArrayValue,
		"m-a-p-p-e-p-m":       mapArrayValue,
		"m-p-a-p-p-e-p-m":     mapArrayValue,
		"m-p-p-a-p-p-e-p-m":   mapArrayValue,
		"m-s-e-p-p-m":         mapSliceValue,
		"m-p-s-e-p-p-m":       mapSliceValue,
		"m-p-p-s-e-p-p-m":     mapSliceValue,
		"m-s-p-e-p-p-m":       mapSliceValue,
		"m-p-s-p-e-p-p-m":     mapSliceValue,
		"m-p-p-s-p-e-p-p-m":   mapSliceValue,
		"m-s-p-p-e-p-p-m":     mapSliceValue,
		"m-p-s-p-p-e-p-p-m":   mapSliceValue,
		"m-p-p-s-p-p-e-p-p-m": mapSliceValue,
		"m-a-e-p-p-m":         mapArrayValue,
		"m-p-a-e-p-p-m":       mapArrayValue,
		"m-p-p-a-e-p-p-m":     mapArrayValue,
		"m-a-p-e-p-p-m":       mapArrayValue,
		"m-p-a-p-e-p-p-m":     mapArrayValue,
		"m-p-p-a-p-e-p-p-m":   mapArrayValue,
		"m-a-p-p-e-p-p-m":     mapArrayValue,
		"m-p-a-p-p-e-p-p-m":   mapArrayValue,
		"m-p-p-a-p-p-e-p-p-m": mapArrayValue,
	}

	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("failed: \nexpected: %#v\nresult:   %#v", expected, target)
	}

	var reverse SourceParent

	if err := Decode(target, &reverse); err != nil {
		t.Fatalf("got error: %s", err)
	}

	if !reflect.DeepEqual(source, reverse) {
		t.Fatalf("failed: \nexpected: %#v\nresult:   %#v", source, reverse)
	}
}

func TestDecode_structArrayDeepMapIndirect(t *testing.T) {
	type SourceChild struct {
		String string `mapstructure:"field"`
	}

	type SourceParent struct {
		SE         []SourceChild                      `mapstructure:"s-e,deep,indirect"`
		SPE        *[]SourceChild                     `mapstructure:"s-p-e,deep,indirect"`
		SPPE       **[]SourceChild                    `mapstructure:"s-p-p-e,deep,indirect"`
		SEP        []*SourceChild                     `mapstructure:"s-e-p,deep,indirect"`
		SPEP       *[]*SourceChild                    `mapstructure:"s-p-e-p,deep,indirect"`
		SPPEP      **[]*SourceChild                   `mapstructure:"s-p-p-e-p,deep,indirect"`
		SEPP       []**SourceChild                    `mapstructure:"s-e-p-p,deep,indirect"`
		SPEPP      *[]**SourceChild                   `mapstructure:"s-p-e-p-p,deep,indirect"`
		SPPEPP     **[]**SourceChild                  `mapstructure:"s-p-p-e-p-p,deep,indirect"`
		AE         [1]SourceChild                     `mapstructure:"a-e,deep,indirect"`
		APE        *[1]SourceChild                    `mapstructure:"a-p-e,deep,indirect"`
		APPE       **[1]SourceChild                   `mapstructure:"a-p-p-e,deep,indirect"`
		AEP        [1]*SourceChild                    `mapstructure:"a-e-p,deep,indirect"`
		APEP       *[1]*SourceChild                   `mapstructure:"a-p-e-p,deep,indirect"`
		APPEP      **[1]*SourceChild                  `mapstructure:"a-p-p-e-p,deep,indirect"`
		AEPP       [1]**SourceChild                   `mapstructure:"a-e-p-p,deep,indirect"`
		APEPP      *[1]**SourceChild                  `mapstructure:"a-p-e-p-p,deep,indirect"`
		APPEPP     **[1]**SourceChild                 `mapstructure:"a-p-p-e-p-p,deep,indirect"`
		ME         map[string]SourceChild             `mapstructure:"m-e,deep,indirect"`
		MPE        *map[string]SourceChild            `mapstructure:"m-p-e,deep,indirect"`
		MPPE       **map[string]SourceChild           `mapstructure:"m-p-p-e,deep,indirect"`
		MEP        map[string]*SourceChild            `mapstructure:"m-e-p,deep,indirect"`
		MPEP       *map[string]*SourceChild           `mapstructure:"m-p-e-p,deep,indirect"`
		MPPEP      **map[string]*SourceChild          `mapstructure:"m-p-p-e-p,deep,indirect"`
		MEPP       map[string]**SourceChild           `mapstructure:"m-e-p-p,deep,indirect"`
		MPEPP      *map[string]**SourceChild          `mapstructure:"m-p-e-p-p,deep,indirect"`
		MPPEPP     **map[string]**SourceChild         `mapstructure:"m-p-p-e-p-p,deep,indirect"`
		MSE        map[string][]SourceChild           `mapstructure:"m-s-e,deep,indirect"`
		MPSE       *map[string][]SourceChild          `mapstructure:"m-p-s-e,deep,indirect"`
		MPPSE      **map[string][]SourceChild         `mapstructure:"m-p-p-s-e,deep,indirect"`
		MSPE       map[string]*[]SourceChild          `mapstructure:"m-s-p-e,deep,indirect"`
		MPSPE      *map[string]*[]SourceChild         `mapstructure:"m-p-s-p-e,deep,indirect"`
		MPPSPE     **map[string]*[]SourceChild        `mapstructure:"m-p-p-s-p-e,deep,indirect"`
		MSPPE      map[string]**[]SourceChild         `mapstructure:"m-s-p-p-e,deep,indirect"`
		MPSPPE     *map[string]**[]SourceChild        `mapstructure:"m-p-s-p-p-e,deep,indirect"`
		MPPSPPE    **map[string]**[]SourceChild       `mapstructure:"m-p-p-s-p-p-e,deep,indirect"`
		MAE        map[string][1]SourceChild          `mapstructure:"m-a-e,deep,indirect"`
		MPAE       *map[string][1]SourceChild         `mapstructure:"m-p-a-e,deep,indirect"`
		MPPAE      **map[string][1]SourceChild        `mapstructure:"m-p-p-a-e,deep,indirect"`
		MAPE       map[string]*[1]SourceChild         `mapstructure:"m-a-p-e,deep,indirect"`
		MPAPE      *map[string]*[1]SourceChild        `mapstructure:"m-p-a-p-e,deep,indirect"`
		MPPAPE     **map[string]*[1]SourceChild       `mapstructure:"m-p-p-a-p-e,deep,indirect"`
		MAPPE      map[string]**[1]SourceChild        `mapstructure:"m-a-p-p-e,deep,indirect"`
		MPAPPE     *map[string]**[1]SourceChild       `mapstructure:"m-p-a-p-p-e,deep,indirect"`
		MPPAPPE    **map[string]**[1]SourceChild      `mapstructure:"m-p-p-a-p-p-e,deep,indirect"`
		MSEP       map[string][]*SourceChild          `mapstructure:"m-s-e-p,deep,indirect"`
		MPSEP      *map[string][]*SourceChild         `mapstructure:"m-p-s-e-p,deep,indirect"`
		MPPSEP     **map[string][]*SourceChild        `mapstructure:"m-p-p-s-e-p,deep,indirect"`
		MSPEP      map[string]*[]*SourceChild         `mapstructure:"m-s-p-e-p,deep,indirect"`
		MPSPEP     *map[string]*[]*SourceChild        `mapstructure:"m-p-s-p-e-p,deep,indirect"`
		MPPSPEP    **map[string]*[]*SourceChild       `mapstructure:"m-p-p-s-p-e-p,deep,indirect"`
		MSPPEP     map[string]**[]*SourceChild        `mapstructure:"m-s-p-p-e-p,deep,indirect"`
		MPSPPEP    *map[string]**[]*SourceChild       `mapstructure:"m-p-s-p-p-e-p,deep,indirect"`
		MPPSPPEP   **map[string]**[]*SourceChild      `mapstructure:"m-p-p-s-p-p-e-p,deep,indirect"`
		MAEP       map[string][1]*SourceChild         `mapstructure:"m-a-e-p,deep,indirect"`
		MPAEP      *map[string][1]*SourceChild        `mapstructure:"m-p-a-e-p,deep,indirect"`
		MPPAEP     **map[string][1]*SourceChild       `mapstructure:"m-p-p-a-e-p,deep,indirect"`
		MAPEP      map[string]*[1]*SourceChild        `mapstructure:"m-a-p-e-p,deep,indirect"`
		MPAPEP     *map[string]*[1]*SourceChild       `mapstructure:"m-p-a-p-e-p,deep,indirect"`
		MPPAPEP    **map[string]*[1]*SourceChild      `mapstructure:"m-p-p-a-p-e-p,deep,indirect"`
		MAPPEP     map[string]**[1]*SourceChild       `mapstructure:"m-a-p-p-e-p,deep,indirect"`
		MPAPPEP    *map[string]**[1]*SourceChild      `mapstructure:"m-p-a-p-p-e-p,deep,indirect"`
		MPPAPPEP   **map[string]**[1]*SourceChild     `mapstructure:"m-p-p-a-p-p-e-p,deep,indirect"`
		MSEPP      map[string][]**SourceChild         `mapstructure:"m-s-e-p-p,deep,indirect"`
		MPSEPP     *map[string][]**SourceChild        `mapstructure:"m-p-s-e-p-p,deep,indirect"`
		MPPSEPP    **map[string][]**SourceChild       `mapstructure:"m-p-p-s-e-p-p,deep,indirect"`
		MSPEPP     map[string]*[]**SourceChild        `mapstructure:"m-s-p-e-p-p,deep,indirect"`
		MPSPEPP    *map[string]*[]**SourceChild       `mapstructure:"m-p-s-p-e-p-p,deep,indirect"`
		MPPSPEPP   **map[string]*[]**SourceChild      `mapstructure:"m-p-p-s-p-e-p-p,deep,indirect"`
		MSPPEPP    map[string]**[]**SourceChild       `mapstructure:"m-s-p-p-e-p-p,deep,indirect"`
		MPSPPEPP   *map[string]**[]**SourceChild      `mapstructure:"m-p-s-p-p-e-p-p,deep,indirect"`
		MPPSPPEPP  **map[string]**[]**SourceChild     `mapstructure:"m-p-p-s-p-p-e-p-p,deep,indirect"`
		MAEPP      map[string][1]**SourceChild        `mapstructure:"m-a-e-p-p,deep,indirect"`
		MPAEPP     *map[string][1]**SourceChild       `mapstructure:"m-p-a-e-p-p,deep,indirect"`
		MPPAEPP    **map[string][1]**SourceChild      `mapstructure:"m-p-p-a-e-p-p,deep,indirect"`
		MAPEPP     map[string]*[1]**SourceChild       `mapstructure:"m-a-p-e-p-p,deep,indirect"`
		MPAPEPP    *map[string]*[1]**SourceChild      `mapstructure:"m-p-a-p-e-p-p,deep,indirect"`
		MPPAPEPP   **map[string]*[1]**SourceChild     `mapstructure:"m-p-p-a-p-e-p-p,deep,indirect"`
		MAPPEPP    map[string]**[1]**SourceChild      `mapstructure:"m-a-p-p-e-p-p,deep,indirect"`
		MPAPPEPP   *map[string]**[1]**SourceChild     `mapstructure:"m-p-a-p-p-e-p-p,deep,indirect"`
		MPPAPPEPP  **map[string]**[1]**SourceChild    `mapstructure:"m-p-p-a-p-p-e-p-p,deep,indirect"`
		SEM        []map[string]any                   `mapstructure:"s-e-m,deep,indirect"`
		SPEM       *[]map[string]any                  `mapstructure:"s-p-e-m,deep,indirect"`
		SPPEM      **[]map[string]any                 `mapstructure:"s-p-p-e-m,deep,indirect"`
		SEPM       []*map[string]any                  `mapstructure:"s-e-p-m,deep,indirect"`
		SPEPM      *[]*map[string]any                 `mapstructure:"s-p-e-p-m,deep,indirect"`
		SPPEPM     **[]*map[string]any                `mapstructure:"s-p-p-e-p-m,deep,indirect"`
		SEPPM      []**map[string]any                 `mapstructure:"s-e-p-p-m,deep,indirect"`
		SPEPPM     *[]**map[string]any                `mapstructure:"s-p-e-p-p-m,deep,indirect"`
		SPPEPPM    **[]**map[string]any               `mapstructure:"s-p-p-e-p-p-m,deep,indirect"`
		AEM        [1]map[string]any                  `mapstructure:"a-e-m,deep,indirect"`
		APEM       *[1]map[string]any                 `mapstructure:"a-p-e-m,deep,indirect"`
		APPEM      **[1]map[string]any                `mapstructure:"a-p-p-e-m,deep,indirect"`
		AEPM       [1]*map[string]any                 `mapstructure:"a-e-p-m,deep,indirect"`
		APEPM      *[1]*map[string]any                `mapstructure:"a-p-e-p-m,deep,indirect"`
		APPEPM     **[1]*map[string]any               `mapstructure:"a-p-p-e-p-m,deep,indirect"`
		AEPPM      [1]**map[string]any                `mapstructure:"a-e-p-p-m,deep,indirect"`
		APEPPM     *[1]**map[string]any               `mapstructure:"a-p-e-p-p-m,deep,indirect"`
		APPEPPM    **[1]**map[string]any              `mapstructure:"a-p-p-e-p-p-m,deep,indirect"`
		MEM        map[string]map[string]any          `mapstructure:"m-e-m,deep,indirect"`
		MPEM       *map[string]map[string]any         `mapstructure:"m-p-e-m,deep,indirect"`
		MPPEM      **map[string]map[string]any        `mapstructure:"m-p-p-e-m,deep,indirect"`
		MEPM       map[string]*map[string]any         `mapstructure:"m-e-p-m,deep,indirect"`
		MPEPM      *map[string]*map[string]any        `mapstructure:"m-p-e-p-m,deep,indirect"`
		MPPEPM     **map[string]*map[string]any       `mapstructure:"m-p-p-e-p-m,deep,indirect"`
		MEPPM      map[string]**map[string]any        `mapstructure:"m-e-p-p-m,deep,indirect"`
		MPEPPM     *map[string]**map[string]any       `mapstructure:"m-p-e-p-p-m,deep,indirect"`
		MPPEPPM    **map[string]**map[string]any      `mapstructure:"m-p-p-e-p-p-m,deep,indirect"`
		MSEM       map[string][]map[string]any        `mapstructure:"m-s-e-m,deep,indirect"`
		MPSEM      *map[string][]map[string]any       `mapstructure:"m-p-s-e-m,deep,indirect"`
		MPPSEM     **map[string][]map[string]any      `mapstructure:"m-p-p-s-e-m,deep,indirect"`
		MSPEM      map[string]*[]map[string]any       `mapstructure:"m-s-p-e-m,deep,indirect"`
		MPSPEM     *map[string]*[]map[string]any      `mapstructure:"m-p-s-p-e-m,deep,indirect"`
		MPPSPEM    **map[string]*[]map[string]any     `mapstructure:"m-p-p-s-p-e-m,deep,indirect"`
		MSPPEM     map[string]**[]map[string]any      `mapstructure:"m-s-p-p-e-m,deep,indirect"`
		MPSPPEM    *map[string]**[]map[string]any     `mapstructure:"m-p-s-p-p-e-m,deep,indirect"`
		MPPSPPEM   **map[string]**[]map[string]any    `mapstructure:"m-p-p-s-p-p-e-m,deep,indirect"`
		MAEM       map[string][1]map[string]any       `mapstructure:"m-a-e-m,deep,indirect"`
		MPAEM      *map[string][1]map[string]any      `mapstructure:"m-p-a-e-m,deep,indirect"`
		MPPAEM     **map[string][1]map[string]any     `mapstructure:"m-p-p-a-e-m,deep,indirect"`
		MAPEM      map[string]*[1]map[string]any      `mapstructure:"m-a-p-e-m,deep,indirect"`
		MPAPEM     *map[string]*[1]map[string]any     `mapstructure:"m-p-a-p-e-m,deep,indirect"`
		MPPAPEM    **map[string]*[1]map[string]any    `mapstructure:"m-p-p-a-p-e-m,deep,indirect"`
		MAPPEM     map[string]**[1]map[string]any     `mapstructure:"m-a-p-p-e-m,deep,indirect"`
		MPAPPEM    *map[string]**[1]map[string]any    `mapstructure:"m-p-a-p-p-e-m,deep,indirect"`
		MPPAPPEM   **map[string]**[1]map[string]any   `mapstructure:"m-p-p-a-p-p-e-m,deep,indirect"`
		MSEPM      map[string][]*map[string]any       `mapstructure:"m-s-e-p-m,deep,indirect"`
		MPSEPM     *map[string][]*map[string]any      `mapstructure:"m-p-s-e-p-m,deep,indirect"`
		MPPSEPM    **map[string][]*map[string]any     `mapstructure:"m-p-p-s-e-p-m,deep,indirect"`
		MSPEPM     map[string]*[]*map[string]any      `mapstructure:"m-s-p-e-p-m,deep,indirect"`
		MPSPEPM    *map[string]*[]*map[string]any     `mapstructure:"m-p-s-p-e-p-m,deep,indirect"`
		MPPSPEPM   **map[string]*[]*map[string]any    `mapstructure:"m-p-p-s-p-e-p-m,deep,indirect"`
		MSPPEPM    map[string]**[]*map[string]any     `mapstructure:"m-s-p-p-e-p-m,deep,indirect"`
		MPSPPEPM   *map[string]**[]*map[string]any    `mapstructure:"m-p-s-p-p-e-p-m,deep,indirect"`
		MPPSPPEPM  **map[string]**[]*map[string]any   `mapstructure:"m-p-p-s-p-p-e-p-m,deep,indirect"`
		MAEPM      map[string][1]*map[string]any      `mapstructure:"m-a-e-p-m,deep,indirect"`
		MPAEPM     *map[string][1]*map[string]any     `mapstructure:"m-p-a-e-p-m,deep,indirect"`
		MPPAEPM    **map[string][1]*map[string]any    `mapstructure:"m-p-p-a-e-p-m,deep,indirect"`
		MAPEPM     map[string]*[1]*map[string]any     `mapstructure:"m-a-p-e-p-m,deep,indirect"`
		MPAPEPM    *map[string]*[1]*map[string]any    `mapstructure:"m-p-a-p-e-p-m,deep,indirect"`
		MPPAPEPM   **map[string]*[1]*map[string]any   `mapstructure:"m-p-p-a-p-e-p-m,deep,indirect"`
		MAPPEPM    map[string]**[1]*map[string]any    `mapstructure:"m-a-p-p-e-p-m,deep,indirect"`
		MPAPPEPM   *map[string]**[1]*map[string]any   `mapstructure:"m-p-a-p-p-e-p-m,deep,indirect"`
		MPPAPPEPM  **map[string]**[1]*map[string]any  `mapstructure:"m-p-p-a-p-p-e-p-m,deep,indirect"`
		MSEPPM     map[string][]**map[string]any      `mapstructure:"m-s-e-p-p-m,deep,indirect"`
		MPSEPPM    *map[string][]**map[string]any     `mapstructure:"m-p-s-e-p-p-m,deep,indirect"`
		MPPSEPPM   **map[string][]**map[string]any    `mapstructure:"m-p-p-s-e-p-p-m,deep,indirect"`
		MSPEPPM    map[string]*[]**map[string]any     `mapstructure:"m-s-p-e-p-p-m,deep,indirect"`
		MPSPEPPM   *map[string]*[]**map[string]any    `mapstructure:"m-p-s-p-e-p-p-m,deep,indirect"`
		MPPSPEPPM  **map[string]*[]**map[string]any   `mapstructure:"m-p-p-s-p-e-p-p-m,deep,indirect"`
		MSPPEPPM   map[string]**[]**map[string]any    `mapstructure:"m-s-p-p-e-p-p-m,deep,indirect"`
		MPSPPEPPM  *map[string]**[]**map[string]any   `mapstructure:"m-p-s-p-p-e-p-p-m,deep,indirect"`
		MPPSPPEPPM **map[string]**[]**map[string]any  `mapstructure:"m-p-p-s-p-p-e-p-p-m,deep,indirect"`
		MAEPPM     map[string][1]**map[string]any     `mapstructure:"m-a-e-p-p-m,deep,indirect"`
		MPAEPPM    *map[string][1]**map[string]any    `mapstructure:"m-p-a-e-p-p-m,deep,indirect"`
		MPPAEPPM   **map[string][1]**map[string]any   `mapstructure:"m-p-p-a-e-p-p-m,deep,indirect"`
		MAPEPPM    map[string]*[1]**map[string]any    `mapstructure:"m-a-p-e-p-p-m,deep,indirect"`
		MPAPEPPM   *map[string]*[1]**map[string]any   `mapstructure:"m-p-a-p-e-p-p-m,deep,indirect"`
		MPPAPEPPM  **map[string]*[1]**map[string]any  `mapstructure:"m-p-p-a-p-e-p-p-m,deep,indirect"`
		MAPPEPPM   map[string]**[1]**map[string]any   `mapstructure:"m-a-p-p-e-p-p-m,deep,indirect"`
		MPAPPEPPM  *map[string]**[1]**map[string]any  `mapstructure:"m-p-a-p-p-e-p-p-m,deep,indirect"`
		MPPAPPEPPM **map[string]**[1]**map[string]any `mapstructure:"m-p-p-a-p-p-e-p-p-m,deep,indirect"`
	}

	e := SourceChild{String: "el"}
	ep := &e
	epp := &ep

	em := map[string]any{"field": "el"}
	epm := &em
	eppm := &epm

	se := []SourceChild{e}
	spe := &se
	sppe := &spe

	sem := []map[string]any{em}
	spem := &sem
	sppem := &spem

	sep := []*SourceChild{ep}
	spep := &sep
	sppep := &spep

	sepm := []*map[string]any{epm}
	spepm := &sepm
	sppepm := &spepm

	sepp := []**SourceChild{epp}
	spepp := &sepp
	sppepp := &spepp

	seppm := []**map[string]any{eppm}
	speppm := &seppm
	sppeppm := &speppm

	ae := [1]SourceChild{e}
	ape := &ae
	appe := &ape

	aem := [1]map[string]any{em}
	apem := &aem
	appem := &apem

	aep := [1]*SourceChild{ep}
	apep := &aep
	appep := &apep

	aepm := [1]*map[string]any{epm}
	apepm := &aepm
	appepm := &apepm

	aepp := [1]**SourceChild{epp}
	apepp := &aepp
	appepp := &apepp

	aeppm := [1]**map[string]any{eppm}
	apeppm := &aeppm
	appeppm := &apeppm

	me := map[string]SourceChild{"foo": e}
	mpe := &me
	mppe := &mpe

	mem := map[string]map[string]any{"foo": em}
	mpem := &mem
	mppem := &mpem

	mep := map[string]*SourceChild{"foo": ep}
	mpep := &mep
	mppep := &mpep

	mepm := map[string]*map[string]any{"foo": epm}
	mpepm := &mepm
	mppepm := &mpepm

	mepp := map[string]**SourceChild{"foo": epp}
	mpepp := &mepp
	mppepp := &mpepp

	meppm := map[string]**map[string]any{"foo": eppm}
	mpeppm := &meppm
	mppeppm := &mpeppm

	mse := map[string][]SourceChild{"foo": se}
	mpse := &mse
	mppse := &mpse

	msem := map[string][]map[string]any{"foo": sem}
	mpsem := &msem
	mppsem := &mpsem

	mspe := map[string]*[]SourceChild{"foo": spe}
	mpspe := &mspe
	mppspe := &mpspe

	mspem := map[string]*[]map[string]any{"foo": spem}
	mpspem := &mspem
	mppspem := &mpspem

	msppe := map[string]**[]SourceChild{"foo": sppe}
	mpsppe := &msppe
	mppsppe := &mpsppe

	msppem := map[string]**[]map[string]any{"foo": sppem}
	mpsppem := &msppem
	mppsppem := &mpsppem

	mae := map[string][1]SourceChild{"foo": ae}
	mpae := &mae
	mppae := &mpae

	maem := map[string][1]map[string]any{"foo": aem}
	mpaem := &maem
	mppaem := &mpaem

	mape := map[string]*[1]SourceChild{"foo": ape}
	mpape := &mape
	mppape := &mpape

	mapem := map[string]*[1]map[string]any{"foo": apem}
	mpapem := &mapem
	mppapem := &mpapem

	mappe := map[string]**[1]SourceChild{"foo": appe}
	mpappe := &mappe
	mppappe := &mpappe

	mappem := map[string]**[1]map[string]any{"foo": appem}
	mpappem := &mappem
	mppappem := &mpappem

	msep := map[string][]*SourceChild{"foo": sep}
	mpsep := &msep
	mppsep := &mpsep

	msepm := map[string][]*map[string]any{"foo": sepm}
	mpsepm := &msepm
	mppsepm := &mpsepm

	mspep := map[string]*[]*SourceChild{"foo": spep}
	mpspep := &mspep
	mppspep := &mpspep

	mspepm := map[string]*[]*map[string]any{"foo": spepm}
	mpspepm := &mspepm
	mppspepm := &mpspepm

	msppep := map[string]**[]*SourceChild{"foo": sppep}
	mpsppep := &msppep
	mppsppep := &mpsppep

	msppepm := map[string]**[]*map[string]any{"foo": sppepm}
	mpsppepm := &msppepm
	mppsppepm := &mpsppepm

	maep := map[string][1]*SourceChild{"foo": aep}
	mpaep := &maep
	mppaep := &mpaep

	maepm := map[string][1]*map[string]any{"foo": aepm}
	mpaepm := &maepm
	mppaepm := &mpaepm

	mapep := map[string]*[1]*SourceChild{"foo": apep}
	mpapep := &mapep
	mppapep := &mpapep

	mapepm := map[string]*[1]*map[string]any{"foo": apepm}
	mpapepm := &mapepm
	mppapepm := &mpapepm

	mappep := map[string]**[1]*SourceChild{"foo": appep}
	mpappep := &mappep
	mppappep := &mpappep

	mappepm := map[string]**[1]*map[string]any{"foo": appepm}
	mpappepm := &mappepm
	mppappepm := &mpappepm

	msepp := map[string][]**SourceChild{"foo": sepp}
	mpsepp := &msepp
	mppsepp := &mpsepp

	mseppm := map[string][]**map[string]any{"foo": seppm}
	mpseppm := &mseppm
	mppseppm := &mpseppm

	mspepp := map[string]*[]**SourceChild{"foo": spepp}
	mpspepp := &mspepp
	mppspepp := &mpspepp

	mspeppm := map[string]*[]**map[string]any{"foo": speppm}
	mpspeppm := &mspeppm
	mppspeppm := &mpspeppm

	msppepp := map[string]**[]**SourceChild{"foo": sppepp}
	mpsppepp := &msppepp
	mppsppepp := &mpsppepp

	msppeppm := map[string]**[]**map[string]any{"foo": sppeppm}
	mpsppeppm := &msppeppm
	mppsppeppm := &mpsppeppm

	maepp := map[string][1]**SourceChild{"foo": aepp}
	mpaepp := &maepp
	mppaepp := &mpaepp

	maeppm := map[string][1]**map[string]any{"foo": aeppm}
	mpaeppm := &maeppm
	mppaeppm := &mpaeppm

	mapepp := map[string]*[1]**SourceChild{"foo": apepp}
	mpapepp := &mapepp
	mppapepp := &mpapepp

	mapeppm := map[string]*[1]**map[string]any{"foo": apeppm}
	mpapeppm := &mapeppm
	mppapeppm := &mpapeppm

	mappepp := map[string]**[1]**SourceChild{"foo": appepp}
	mpappepp := &mappepp
	mppappepp := &mpappepp

	mappeppm := map[string]**[1]**map[string]any{"foo": appeppm}
	mpappeppm := &mappeppm
	mppappeppm := &mpappeppm

	var target map[string]any

	source := SourceParent{
		SE:         se,
		SPE:        spe,
		SPPE:       sppe,
		SEP:        sep,
		SPEP:       spep,
		SPPEP:      sppep,
		SEPP:       sepp,
		SPEPP:      spepp,
		SPPEPP:     sppepp,
		AE:         ae,
		APE:        ape,
		APPE:       appe,
		AEP:        aep,
		APEP:       apep,
		APPEP:      appep,
		AEPP:       aepp,
		APEPP:      apepp,
		APPEPP:     appepp,
		ME:         me,
		MPE:        mpe,
		MPPE:       mppe,
		MEP:        mep,
		MPEP:       mpep,
		MPPEP:      mppep,
		MEPP:       mepp,
		MPEPP:      mpepp,
		MPPEPP:     mppepp,
		MSE:        mse,
		MPSE:       mpse,
		MPPSE:      mppse,
		MSPE:       mspe,
		MPSPE:      mpspe,
		MPPSPE:     mppspe,
		MSPPE:      msppe,
		MPSPPE:     mpsppe,
		MPPSPPE:    mppsppe,
		MAE:        mae,
		MPAE:       mpae,
		MPPAE:      mppae,
		MAPE:       mape,
		MPAPE:      mpape,
		MPPAPE:     mppape,
		MAPPE:      mappe,
		MPAPPE:     mpappe,
		MPPAPPE:    mppappe,
		MSEP:       msep,
		MPSEP:      mpsep,
		MPPSEP:     mppsep,
		MSPEP:      mspep,
		MPSPEP:     mpspep,
		MPPSPEP:    mppspep,
		MSPPEP:     msppep,
		MPSPPEP:    mpsppep,
		MPPSPPEP:   mppsppep,
		MAEP:       maep,
		MPAEP:      mpaep,
		MPPAEP:     mppaep,
		MAPEP:      mapep,
		MPAPEP:     mpapep,
		MPPAPEP:    mppapep,
		MAPPEP:     mappep,
		MPAPPEP:    mpappep,
		MPPAPPEP:   mppappep,
		MSEPP:      msepp,
		MPSEPP:     mpsepp,
		MPPSEPP:    mppsepp,
		MSPEPP:     mspepp,
		MPSPEPP:    mpspepp,
		MPPSPEPP:   mppspepp,
		MSPPEPP:    msppepp,
		MPSPPEPP:   mpsppepp,
		MPPSPPEPP:  mppsppepp,
		MAEPP:      maepp,
		MPAEPP:     mpaepp,
		MPPAEPP:    mppaepp,
		MAPEPP:     mapepp,
		MPAPEPP:    mpapepp,
		MPPAPEPP:   mppapepp,
		MAPPEPP:    mappepp,
		MPAPPEPP:   mpappepp,
		MPPAPPEPP:  mppappepp,
		SEM:        sem,
		SPEM:       spem,
		SPPEM:      sppem,
		SEPM:       sepm,
		SPEPM:      spepm,
		SPPEPM:     sppepm,
		SEPPM:      seppm,
		SPEPPM:     speppm,
		SPPEPPM:    sppeppm,
		AEM:        aem,
		APEM:       apem,
		APPEM:      appem,
		AEPM:       aepm,
		APEPM:      apepm,
		APPEPM:     appepm,
		AEPPM:      aeppm,
		APEPPM:     apeppm,
		APPEPPM:    appeppm,
		MEM:        mem,
		MPEM:       mpem,
		MPPEM:      mppem,
		MEPM:       mepm,
		MPEPM:      mpepm,
		MPPEPM:     mppepm,
		MEPPM:      meppm,
		MPEPPM:     mpeppm,
		MPPEPPM:    mppeppm,
		MSEM:       msem,
		MPSEM:      mpsem,
		MPPSEM:     mppsem,
		MSPEM:      mspem,
		MPSPEM:     mpspem,
		MPPSPEM:    mppspem,
		MSPPEM:     msppem,
		MPSPPEM:    mpsppem,
		MPPSPPEM:   mppsppem,
		MAEM:       maem,
		MPAEM:      mpaem,
		MPPAEM:     mppaem,
		MAPEM:      mapem,
		MPAPEM:     mpapem,
		MPPAPEM:    mppapem,
		MAPPEM:     mappem,
		MPAPPEM:    mpappem,
		MPPAPPEM:   mppappem,
		MSEPM:      msepm,
		MPSEPM:     mpsepm,
		MPPSEPM:    mppsepm,
		MSPEPM:     mspepm,
		MPSPEPM:    mpspepm,
		MPPSPEPM:   mppspepm,
		MSPPEPM:    msppepm,
		MPSPPEPM:   mpsppepm,
		MPPSPPEPM:  mppsppepm,
		MAEPM:      maepm,
		MPAEPM:     mpaepm,
		MPPAEPM:    mppaepm,
		MAPEPM:     mapepm,
		MPAPEPM:    mpapepm,
		MPPAPEPM:   mppapepm,
		MAPPEPM:    mappepm,
		MPAPPEPM:   mpappepm,
		MPPAPPEPM:  mppappepm,
		MSEPPM:     mseppm,
		MPSEPPM:    mpseppm,
		MPPSEPPM:   mppseppm,
		MSPEPPM:    mspeppm,
		MPSPEPPM:   mpspeppm,
		MPPSPEPPM:  mppspeppm,
		MSPPEPPM:   msppeppm,
		MPSPPEPPM:  mpsppeppm,
		MPPSPPEPPM: mppsppeppm,
		MAEPPM:     maeppm,
		MPAEPPM:    mpaeppm,
		MPPAEPPM:   mppaeppm,
		MAPEPPM:    mapeppm,
		MPAPEPPM:   mpapeppm,
		MPPAPEPPM:  mppapeppm,
		MAPPEPPM:   mappeppm,
		MPAPPEPPM:  mpappeppm,
		MPPAPPEPPM: mppappeppm,
	}

	if err := Decode(source, &target); err != nil {
		t.Fatalf("got error: %s", err)
	}

	sliceValue := []any{map[string]any{"field": "el"}}
	mapValue := map[string]any{"foo": map[string]any{"field": "el"}}
	mapSliceValue := map[string]any{"foo": []any{map[string]any{"field": "el"}}}

	expected := map[string]any{
		"s-e":                 sliceValue,
		"s-p-e":               sliceValue,
		"s-p-p-e":             sliceValue,
		"s-e-p":               sliceValue,
		"s-p-e-p":             sliceValue,
		"s-p-p-e-p":           sliceValue,
		"s-e-p-p":             sliceValue,
		"s-p-e-p-p":           sliceValue,
		"s-p-p-e-p-p":         sliceValue,
		"a-e":                 sliceValue,
		"a-p-e":               sliceValue,
		"a-p-p-e":             sliceValue,
		"a-e-p":               sliceValue,
		"a-p-e-p":             sliceValue,
		"a-p-p-e-p":           sliceValue,
		"a-e-p-p":             sliceValue,
		"a-p-e-p-p":           sliceValue,
		"a-p-p-e-p-p":         sliceValue,
		"m-e":                 mapValue,
		"m-p-e":               mapValue,
		"m-p-p-e":             mapValue,
		"m-e-p":               mapValue,
		"m-p-e-p":             mapValue,
		"m-p-p-e-p":           mapValue,
		"m-e-p-p":             mapValue,
		"m-p-e-p-p":           mapValue,
		"m-p-p-e-p-p":         mapValue,
		"m-s-e":               mapSliceValue,
		"m-p-s-e":             mapSliceValue,
		"m-p-p-s-e":           mapSliceValue,
		"m-s-p-e":             mapSliceValue,
		"m-p-s-p-e":           mapSliceValue,
		"m-p-p-s-p-e":         mapSliceValue,
		"m-s-p-p-e":           mapSliceValue,
		"m-p-s-p-p-e":         mapSliceValue,
		"m-p-p-s-p-p-e":       mapSliceValue,
		"m-a-e":               mapSliceValue,
		"m-p-a-e":             mapSliceValue,
		"m-p-p-a-e":           mapSliceValue,
		"m-a-p-e":             mapSliceValue,
		"m-p-a-p-e":           mapSliceValue,
		"m-p-p-a-p-e":         mapSliceValue,
		"m-a-p-p-e":           mapSliceValue,
		"m-p-a-p-p-e":         mapSliceValue,
		"m-p-p-a-p-p-e":       mapSliceValue,
		"m-s-e-p":             mapSliceValue,
		"m-p-s-e-p":           mapSliceValue,
		"m-p-p-s-e-p":         mapSliceValue,
		"m-s-p-e-p":           mapSliceValue,
		"m-p-s-p-e-p":         mapSliceValue,
		"m-p-p-s-p-e-p":       mapSliceValue,
		"m-s-p-p-e-p":         mapSliceValue,
		"m-p-s-p-p-e-p":       mapSliceValue,
		"m-p-p-s-p-p-e-p":     mapSliceValue,
		"m-a-e-p":             mapSliceValue,
		"m-p-a-e-p":           mapSliceValue,
		"m-p-p-a-e-p":         mapSliceValue,
		"m-a-p-e-p":           mapSliceValue,
		"m-p-a-p-e-p":         mapSliceValue,
		"m-p-p-a-p-e-p":       mapSliceValue,
		"m-a-p-p-e-p":         mapSliceValue,
		"m-p-a-p-p-e-p":       mapSliceValue,
		"m-p-p-a-p-p-e-p":     mapSliceValue,
		"m-s-e-p-p":           mapSliceValue,
		"m-p-s-e-p-p":         mapSliceValue,
		"m-p-p-s-e-p-p":       mapSliceValue,
		"m-s-p-e-p-p":         mapSliceValue,
		"m-p-s-p-e-p-p":       mapSliceValue,
		"m-p-p-s-p-e-p-p":     mapSliceValue,
		"m-s-p-p-e-p-p":       mapSliceValue,
		"m-p-s-p-p-e-p-p":     mapSliceValue,
		"m-p-p-s-p-p-e-p-p":   mapSliceValue,
		"m-a-e-p-p":           mapSliceValue,
		"m-p-a-e-p-p":         mapSliceValue,
		"m-p-p-a-e-p-p":       mapSliceValue,
		"m-a-p-e-p-p":         mapSliceValue,
		"m-p-a-p-e-p-p":       mapSliceValue,
		"m-p-p-a-p-e-p-p":     mapSliceValue,
		"m-a-p-p-e-p-p":       mapSliceValue,
		"m-p-a-p-p-e-p-p":     mapSliceValue,
		"m-p-p-a-p-p-e-p-p":   mapSliceValue,
		"s-e-m":               sliceValue,
		"s-p-e-m":             sliceValue,
		"s-p-p-e-m":           sliceValue,
		"s-e-p-m":             sliceValue,
		"s-p-e-p-m":           sliceValue,
		"s-p-p-e-p-m":         sliceValue,
		"s-e-p-p-m":           sliceValue,
		"s-p-e-p-p-m":         sliceValue,
		"s-p-p-e-p-p-m":       sliceValue,
		"a-e-m":               sliceValue,
		"a-p-e-m":             sliceValue,
		"a-p-p-e-m":           sliceValue,
		"a-e-p-m":             sliceValue,
		"a-p-e-p-m":           sliceValue,
		"a-p-p-e-p-m":         sliceValue,
		"a-e-p-p-m":           sliceValue,
		"a-p-e-p-p-m":         sliceValue,
		"a-p-p-e-p-p-m":       sliceValue,
		"m-e-m":               mapValue,
		"m-p-e-m":             mapValue,
		"m-p-p-e-m":           mapValue,
		"m-e-p-m":             mapValue,
		"m-p-e-p-m":           mapValue,
		"m-p-p-e-p-m":         mapValue,
		"m-e-p-p-m":           mapValue,
		"m-p-e-p-p-m":         mapValue,
		"m-p-p-e-p-p-m":       mapValue,
		"m-s-e-m":             mapSliceValue,
		"m-p-s-e-m":           mapSliceValue,
		"m-p-p-s-e-m":         mapSliceValue,
		"m-s-p-e-m":           mapSliceValue,
		"m-p-s-p-e-m":         mapSliceValue,
		"m-p-p-s-p-e-m":       mapSliceValue,
		"m-s-p-p-e-m":         mapSliceValue,
		"m-p-s-p-p-e-m":       mapSliceValue,
		"m-p-p-s-p-p-e-m":     mapSliceValue,
		"m-a-e-m":             mapSliceValue,
		"m-p-a-e-m":           mapSliceValue,
		"m-p-p-a-e-m":         mapSliceValue,
		"m-a-p-e-m":           mapSliceValue,
		"m-p-a-p-e-m":         mapSliceValue,
		"m-p-p-a-p-e-m":       mapSliceValue,
		"m-a-p-p-e-m":         mapSliceValue,
		"m-p-a-p-p-e-m":       mapSliceValue,
		"m-p-p-a-p-p-e-m":     mapSliceValue,
		"m-s-e-p-m":           mapSliceValue,
		"m-p-s-e-p-m":         mapSliceValue,
		"m-p-p-s-e-p-m":       mapSliceValue,
		"m-s-p-e-p-m":         mapSliceValue,
		"m-p-s-p-e-p-m":       mapSliceValue,
		"m-p-p-s-p-e-p-m":     mapSliceValue,
		"m-s-p-p-e-p-m":       mapSliceValue,
		"m-p-s-p-p-e-p-m":     mapSliceValue,
		"m-p-p-s-p-p-e-p-m":   mapSliceValue,
		"m-a-e-p-m":           mapSliceValue,
		"m-p-a-e-p-m":         mapSliceValue,
		"m-p-p-a-e-p-m":       mapSliceValue,
		"m-a-p-e-p-m":         mapSliceValue,
		"m-p-a-p-e-p-m":       mapSliceValue,
		"m-p-p-a-p-e-p-m":     mapSliceValue,
		"m-a-p-p-e-p-m":       mapSliceValue,
		"m-p-a-p-p-e-p-m":     mapSliceValue,
		"m-p-p-a-p-p-e-p-m":   mapSliceValue,
		"m-s-e-p-p-m":         mapSliceValue,
		"m-p-s-e-p-p-m":       mapSliceValue,
		"m-p-p-s-e-p-p-m":     mapSliceValue,
		"m-s-p-e-p-p-m":       mapSliceValue,
		"m-p-s-p-e-p-p-m":     mapSliceValue,
		"m-p-p-s-p-e-p-p-m":   mapSliceValue,
		"m-s-p-p-e-p-p-m":     mapSliceValue,
		"m-p-s-p-p-e-p-p-m":   mapSliceValue,
		"m-p-p-s-p-p-e-p-p-m": mapSliceValue,
		"m-a-e-p-p-m":         mapSliceValue,
		"m-p-a-e-p-p-m":       mapSliceValue,
		"m-p-p-a-e-p-p-m":     mapSliceValue,
		"m-a-p-e-p-p-m":       mapSliceValue,
		"m-p-a-p-e-p-p-m":     mapSliceValue,
		"m-p-p-a-p-e-p-p-m":   mapSliceValue,
		"m-a-p-p-e-p-p-m":     mapSliceValue,
		"m-p-a-p-p-e-p-p-m":   mapSliceValue,
		"m-p-p-a-p-p-e-p-p-m": mapSliceValue,
	}

	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("failed: \nexpected: %#v\nresult:   %#v", expected, target)
	}

	var reverse SourceParent

	if err := Decode(target, &reverse); err != nil {
		t.Fatalf("got error: %s", err)
	}

	if !reflect.DeepEqual(source, reverse) {
		t.Fatalf("failed: \nexpected: %#v\nresult:   %#v", source, reverse)
	}
}

func TestDecode_structNestedArrayDeepMapIndirect(t *testing.T) {
	type SourceChild struct {
		Foo string `mapstructure:"foo"`
	}

	type Source struct {
		Map             map[string]int         `mapstructure:"map,deep"`
		MapIndirect     map[string]int         `mapstructure:"map_indirect,deep,indirect"`
		MapStruct       map[string]SourceChild `mapstructure:"map_struct"`
		MapStructDeep   map[string]SourceChild `mapstructure:"map_struct_deep,deep"`
		Array           [1]int                 `mapstructure:"array,deep"`
		ArrayIndirect   [1]int                 `mapstructure:"array_indirect,deep,indirect"`
		ArrayStruct     [1]SourceChild         `mapstructure:"array_struct"`
		ArrayStructDeep [1]SourceChild         `mapstructure:"array_struct_deep,deep"`
		Slice           []int                  `mapstructure:"slice,deep"`
		SliceIndirect   []int                  `mapstructure:"slice_indirect,deep,indirect"`
		SliceStruct     []SourceChild          `mapstructure:"slice_struct"`
		SliceStructDeep []SourceChild          `mapstructure:"slice_struct_deep,deep"`
	}

	var (
		source Source
		target map[string]any
	)

	if err := Decode(source, &target); err != nil {
		t.Fatalf("got error: %s", err)
	}

	cases := []struct {
		field    string
		expected reflect.Type
	}{
		{"map", reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(0))},
		{"map_indirect", reflect.MapOf(reflect.TypeOf(""), interfaceType)},
		{"map_struct", reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(SourceChild{}))},
		{"map_struct_deep", reflect.MapOf(reflect.TypeOf(""), reflect.MapOf(reflect.TypeOf(""), interfaceType))},
		{"array", reflect.ArrayOf(1, reflect.TypeOf(0))},
		{"array_indirect", reflect.SliceOf(interfaceType)},
		{"array_struct", reflect.ArrayOf(1, reflect.TypeOf(SourceChild{}))},
		{"array_struct_deep", reflect.ArrayOf(1, reflect.MapOf(reflect.TypeOf(""), interfaceType))},
		{"slice", reflect.SliceOf(reflect.TypeOf(0))},
		{"slice_indirect", reflect.SliceOf(interfaceType)},
		{"slice_struct", reflect.SliceOf(reflect.TypeOf(SourceChild{}))},
		{"slice_struct_deep", reflect.SliceOf(reflect.MapOf(reflect.TypeOf(""), interfaceType))},
	}

	for _, c := range cases {
		actual := reflect.TypeOf(target[c.field])
		if actual != c.expected {
			t.Errorf("invalid %s type: %v", c.field, actual)
		}
	}
}

func TestDecoder_RootName(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"surname": "green",
		"relation": map[string]any{
			"surname": "black",
		},
	}

	var result struct {
		Name     string `mapstructure:"name"`
		Relation struct {
			Name string `mapstructure:"name"`
		} `mapstructure:"relation"`
	}

	decoder, err := NewDecoder(&DecoderConfig{
		ErrorUnset:  true,
		ErrorUnused: true,
		Result:      &result,
		RootName:    "root",
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err == nil {
		t.Fatal("expected error")
	}

	expectedErrors := []string{
		`'root.relation' has invalid keys: surname`,
		`'root.relation' has unset fields: name`,
		`'root' has invalid keys: surname`,
		`'root' has unset fields: name`,
	}

	failed := false

	for _, expectedErr := range expectedErrors {
		if !strings.Contains(err.Error(), expectedErr) {
			failed = true
		}
	}

	if failed {
		t.Errorf("unexpected error message, got: %s", err.Error())
	}
}

func stringPtr(v string) *string  { return &v }
func intPtr(v int) *int           { return &v }
func uintPtr(v uint) *uint        { return &v }
func boolPtr(v bool) *bool        { return &v }
func floatPtr(v float64) *float64 { return &v }
func interfacePtr(v any) *any     { return &v }

type TestMapFieldName struct {
	HostName string
	Username string
}

func TestDecoder_MapFieldNameMapFromStruct(t *testing.T) {
	t.Parallel()

	var structKeys map[string]any

	decoder, err := NewDecoder(&DecoderConfig{
		ErrorUnused: true, // Enable error on unused keys
		Result:      &structKeys,
		MapFieldName: func(s string) string {
			if s == "HostName" {
				return "host_name"
			}
			return s
		},
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	var input TestMapFieldName

	err = decoder.Decode(&input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, ok := structKeys["host_name"]
	if !ok {
		t.Fatal("expected host_name to exist")
	}

	_, ok = structKeys["Username"]
	if !ok {
		t.Fatal("expected Username to exist")
	}
}

func TestDecoder_MapFieldNameStructFromMap(t *testing.T) {
	t.Parallel()

	foo := TestMapFieldName{}

	decoder, err := NewDecoder(&DecoderConfig{
		Result: &foo,
		MapFieldName: func(s string) string {
			if s == "HostName" {
				return "host_name"
			}
			if s == "Username" {
				return "user_name"
			}
			return s
		},
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	structKeys := map[string]any{
		"host_name": "foo",
		"user_name": "bar",
	}

	err = decoder.Decode(&structKeys)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if foo.HostName != "foo" {
		t.Fatal("expected HostName to be foo")
	}

	if foo.Username != "bar" {
		t.Fatal("expected Username to be bar")
	}
}

func TestDecoder_MapFieldNameWithMatchName(t *testing.T) {
	t.Parallel()

	type Target struct {
		HostName string
		Username string
	}

	input := map[string]any{
		"HOST_NAME": "server1",
		"user_name": "admin",
	}

	var result Target
	decoder, err := NewDecoder(&DecoderConfig{
		Result: &result,
		MapFieldName: func(s string) string {
			// Convert HostName -> host_name, Username -> user_name
			if s == "HostName" {
				return "host_name"
			}
			if s == "Username" {
				return "user_name"
			}
			return s
		},
		MatchName: strings.EqualFold, // Case-insensitive fallback
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// HOST_NAME should match host_name via case-insensitive MatchName
	if result.HostName != "server1" {
		t.Fatalf("expected HostName to be 'server1', got '%s'", result.HostName)
	}

	// user_name matches exactly
	if result.Username != "admin" {
		t.Fatalf("expected Username to be 'admin', got '%s'", result.Username)
	}
}

func TestDecoder_MapFieldNameWithIgnoreUntaggedFields(t *testing.T) {
	t.Parallel()

	type Target struct {
		TaggedField   string `mapstructure:"tagged_field"`
		UntaggedField string
	}

	input := map[string]any{
		"tagged_field":   "tagged_value",
		"untagged_field": "untagged_value",
	}

	var result Target
	decoder, err := NewDecoder(&DecoderConfig{
		Result:               &result,
		IgnoreUntaggedFields: true,
		MapFieldName: func(s string) string {
			// This would convert UntaggedField -> untagged_field,
			// but IgnoreUntaggedFields takes precedence
			if s == "UntaggedField" {
				return "untagged_field"
			}
			return s
		},
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.TaggedField != "tagged_value" {
		t.Fatalf("expected TaggedField to be 'tagged_value', got '%s'", result.TaggedField)
	}

	// UntaggedField should remain empty because IgnoreUntaggedFields is true
	if result.UntaggedField != "" {
		t.Fatalf("expected UntaggedField to be empty due to IgnoreUntaggedFields, got '%s'", result.UntaggedField)
	}
}

// Test types for Unmarshaler
type PersonWithUnmarshaler struct {
	Name string
	Age  int
	raw  map[string]any // stores raw data for verification
}

func (p *PersonWithUnmarshaler) UnmarshalMapstructure(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map[string]any, got %T", data)
	}
	p.raw = m
	if name, ok := m["name"].(string); ok {
		p.Name = name
	}
	if age, ok := m["age"].(int); ok {
		p.Age = age
	}
	return nil
}

func TestUnmarshaler_Basic(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	var result PersonWithUnmarshaler
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %q", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("expected Age 30, got %d", result.Age)
	}
	if result.raw == nil {
		t.Error("expected raw to be set, but it was nil")
	}
}

func TestUnmarshaler_WithDecodeHook(t *testing.T) {
	t.Parallel()

	// DecodeHook that transforms the data before Unmarshaler receives it
	hook := func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if m, ok := data.(map[string]any); ok {
			if name, ok := m["name"].(string); ok {
				m["name"] = strings.ToUpper(name)
			}
		}
		return data, nil
	}

	input := map[string]any{
		"name": "alice",
		"age":  30,
	}

	var result PersonWithUnmarshaler
	config := &DecoderConfig{
		DecodeHook: hook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("failed to create decoder: %s", err)
	}

	if err := decoder.Decode(input); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Unmarshaler should receive transformed data
	if result.Name != "ALICE" {
		t.Errorf("expected Name 'ALICE' (transformed), got %q", result.Name)
	}
}

func TestUnmarshaler_Disabled(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	var result PersonWithUnmarshaler
	config := &DecoderConfig{
		DisableUnmarshaler: true,
		Result:             &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("failed to create decoder: %s", err)
	}

	if err := decoder.Decode(input); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// With Unmarshaler disabled, standard decoding should be used
	// raw should be nil since UnmarshalMapstructure wasn't called
	if result.raw != nil {
		t.Error("Unmarshaler should not have been called when disabled")
	}

	// But the fields should still be decoded via normal path
	if result.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %q", result.Name)
	}
}

// AddressWithUnmarshaler demonstrates nested struct with Unmarshaler
type AddressWithUnmarshaler struct {
	City    string
	Country string
}

func (a *AddressWithUnmarshaler) UnmarshalMapstructure(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map, got %T", data)
	}
	if city, ok := m["city"].(string); ok {
		a.City = strings.ToUpper(city)
	}
	if country, ok := m["country"].(string); ok {
		a.Country = strings.ToUpper(country)
	}
	return nil
}

type PersonWithAddress struct {
	Name    string
	Address AddressWithUnmarshaler
}

func TestUnmarshaler_Nested(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"name": "Alice",
		"address": map[string]any{
			"city":    "new york",
			"country": "usa",
		},
	}

	var result PersonWithAddress
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %q", result.Name)
	}
	// Address should use its Unmarshaler, uppercasing the values
	if result.Address.City != "NEW YORK" {
		t.Errorf("expected City 'NEW YORK', got %q", result.Address.City)
	}
	if result.Address.Country != "USA" {
		t.Errorf("expected Country 'USA', got %q", result.Address.Country)
	}
}

// SquashInnerNoUnmarshaler is a simple struct without Unmarshaler for squash testing
type SquashInnerNoUnmarshaler struct {
	City string
}

// PersonWithSquashedNoUnmarshaler tests that squash works normally when embedded type
// doesn't implement Unmarshaler
type PersonWithSquashedNoUnmarshaler struct {
	Name                     string
	SquashInnerNoUnmarshaler `mapstructure:",squash"`
}

func TestUnmarshaler_SquashWorksNormally(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"name": "Alice",
		"city": "New York",
	}

	var result PersonWithSquashedNoUnmarshaler
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %q", result.Name)
	}
	// Squashed field should be decoded via standard path
	if result.City != "New York" {
		t.Errorf("expected City 'New York', got %q", result.City)
	}
}

// EmbeddedWithUnmarshaler tests embedded struct with Unmarshaler
type EmbeddedWithUnmarshaler struct {
	City string
}

var embeddedUnmarshalerCalled = false

func (e *EmbeddedWithUnmarshaler) UnmarshalMapstructure(data any) error {
	embeddedUnmarshalerCalled = true
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map, got %T", data)
	}
	if city, ok := m["city"].(string); ok {
		e.City = strings.ToUpper(city)
	}
	return nil
}

// Note: When a type embeds another type that implements Unmarshaler,
// Go's embedding rules cause the parent type to also implement Unmarshaler.
// This test documents that behavior.
type PersonWithSquashedEmbedded struct {
	Name                    string
	EmbeddedWithUnmarshaler `mapstructure:",squash"`
}

func TestUnmarshaler_EmbeddingInheritsInterface(t *testing.T) {
	// Note: not parallel because we're using a package-level flag
	embeddedUnmarshalerCalled = false

	input := map[string]any{
		"name": "Alice",
		"city": "New York",
	}

	var result PersonWithSquashedEmbedded
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Due to Go's embedding rules, PersonWithSquashedEmbedded implements
	// Unmarshaler because EmbeddedWithUnmarshaler does. So the Unmarshaler
	// IS called, and it only knows about the "city" field.
	if !embeddedUnmarshalerCalled {
		t.Error("Unmarshaler should have been called due to embedding inheritance")
	}
	// City should be uppercased because Unmarshaler was called
	if result.City != "NEW YORK" {
		t.Errorf("expected City 'NEW YORK', got %q", result.City)
	}
}

// FailingUnmarshaler tests error handling
type FailingUnmarshaler struct {
	Value string
}

func (f *FailingUnmarshaler) UnmarshalMapstructure(data any) error {
	return fmt.Errorf("intentional error")
}

func TestUnmarshaler_ErrorHandling(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"value": "test",
	}

	var result FailingUnmarshaler
	err := Decode(input, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Check that error is wrapped in DecodeError
	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected DecodeError, got %T: %v", err, err)
	}
}

// PointerReceiverUnmarshaler tests pointer receiver only
type PointerReceiverUnmarshaler struct {
	Value string
}

func (p *PointerReceiverUnmarshaler) UnmarshalMapstructure(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map, got %T", data)
	}
	if v, ok := m["value"].(string); ok {
		p.Value = v + "_unmarshaled"
	}
	return nil
}

func TestUnmarshaler_PointerReceiver(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"value": "test",
	}

	var result PointerReceiverUnmarshaler
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Value != "test_unmarshaled" {
		t.Errorf("expected 'test_unmarshaled', got %q", result.Value)
	}
}

func TestUnmarshaler_Metadata(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	var md Metadata
	var result PersonWithUnmarshaler
	config := &DecoderConfig{
		Metadata: &md,
		Result:   &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("failed to create decoder: %s", err)
	}

	if err := decoder.Decode(input); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Verify Unmarshaler was called
	if result.raw == nil {
		t.Error("expected Unmarshaler to be called")
	}
}

// NonSquashedEmbeddedWithUnmarshaler tests that non-squashed embedded structs DO use Unmarshaler
type PersonWithNonSquashedEmbedded struct {
	Name    string
	Address EmbeddedWithUnmarshaler
}

func TestUnmarshaler_NonSquashedEmbedded(t *testing.T) {
	// Note: not parallel because we're using a package-level flag
	embeddedUnmarshalerCalled = false

	input := map[string]any{
		"name": "Alice",
		"address": map[string]any{
			"city": "new york",
		},
	}

	var result PersonWithNonSquashedEmbedded
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %q", result.Name)
	}
	// Non-squashed embedded field SHOULD use Unmarshaler (uppercased)
	if result.Address.City != "NEW YORK" {
		t.Errorf("expected City 'NEW YORK', got %q", result.Address.City)
	}
	if !embeddedUnmarshalerCalled {
		t.Error("Unmarshaler should have been called for non-squashed embedded field")
	}
}

// CustomString is a non-struct type that implements Unmarshaler
type CustomString string

func (c *CustomString) UnmarshalMapstructure(data any) error {
	s, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", data)
	}
	*c = CustomString("custom:" + s)
	return nil
}

func TestUnmarshaler_NonStructType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"name": "test-value",
	}

	type Result struct {
		Name CustomString
	}

	var result Result
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Name != "custom:test-value" {
		t.Errorf("expected 'custom:test-value', got %q", result.Name)
	}
}

// CustomSlice is a slice type that implements Unmarshaler
type CustomSlice []string

func (c *CustomSlice) UnmarshalMapstructure(data any) error {
	switch v := data.(type) {
	case []any:
		*c = make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				(*c)[i] = "item:" + s
			}
		}
	case string:
		// Also accept a single string
		*c = []string{"single:" + v}
	default:
		return fmt.Errorf("expected slice or string, got %T", data)
	}
	return nil
}

func TestUnmarshaler_SliceType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"tags": []any{"a", "b", "c"},
	}

	type Result struct {
		Tags CustomSlice
	}

	var result Result
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := CustomSlice{"item:a", "item:b", "item:c"}
	if !reflect.DeepEqual(result.Tags, expected) {
		t.Errorf("expected %v, got %v", expected, result.Tags)
	}
}

// CustomMap is a map type that implements Unmarshaler
type CustomMap map[string]int

func (c *CustomMap) UnmarshalMapstructure(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map, got %T", data)
	}
	*c = make(map[string]int)
	for k, v := range m {
		if num, ok := v.(int); ok {
			(*c)["key:"+k] = num * 10
		}
	}
	return nil
}

func TestUnmarshaler_MapType(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"scores": map[string]any{
			"alice": 5,
			"bob":   10,
		},
	}

	type Result struct {
		Scores CustomMap
	}

	var result Result
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Scores["key:alice"] != 50 {
		t.Errorf("expected key:alice=50, got %d", result.Scores["key:alice"])
	}
	if result.Scores["key:bob"] != 100 {
		t.Errorf("expected key:bob=100, got %d", result.Scores["key:bob"])
	}
}

func TestUnmarshaler_StructToMap(t *testing.T) {
	t.Parallel()

	// Input is a struct that implements Unmarshaler
	// Output is a map - Unmarshaler should NOT be involved
	input := PersonWithUnmarshaler{
		Name: "Alice",
		Age:  30,
	}

	var result map[string]any
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Should use normal struct-to-map decoding
	if result["Name"] != "Alice" {
		t.Errorf("expected Name 'Alice', got %v", result["Name"])
	}
	if result["Age"] != 30 {
		t.Errorf("expected Age 30, got %v", result["Age"])
	}
}
