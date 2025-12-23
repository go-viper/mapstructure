package mapstructure

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// PrimitiveStringConvertible defines the constraint for primitive types that can be converted from strings.
type PrimitiveStringConvertible interface {
	~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 |
		~int | ~uint | ~float32 | ~float64 | ~bool | ~complex64 | ~complex128
}

// ComplexStringConvertible defines the constraint for complex types that can be converted from strings.
type ComplexStringConvertible interface {
	time.Duration | *url.URL | net.IP | *net.IPNet | netip.Addr | netip.AddrPort | netip.Prefix
}

// StringConvertible defines the constraint for all types that can be converted from strings.
type StringConvertible interface {
	PrimitiveStringConvertible | ComplexStringConvertible
}

// StringToHookFuncWithParser creates a DecodeHookFunc that converts strings to type T
// using the provided parseFunc allowing for custom parsing logic.
//
// Unlike [StringToHookFunc], this function supports tilde types (~int8, ~uint8, etc.)
// which allows it to work with custom type aliases at compile time:
//
//	type MyInt int32
//	customParser := func(s string) (MyInt, error) {
//		val, err := strconv.ParseInt(s, 0, 32)
//		return MyInt(val), err
//	}
//	hook := StringParserHookFunc(customParser)
func StringParserHookFunc[T StringConvertible](parseFunc func(string) (T, error)) DecodeHookFunc {
	var zero T
	expectedType := reflect.TypeOf(zero)

	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		// Type checking with special case for net.IPNet
		if expectedType == reflect.TypeOf((*net.IPNet)(nil)) {
			expectedType = reflect.TypeOf(net.IPNet{})
		}

		if t != expectedType {
			return data, nil
		}

		return parseFunc(data.(string))
	}
}

// ExactPrimitiveStringConvertible defines the constraint for primitive types that can be converted from strings.
type ExactPrimitiveStringConvertible interface {
	int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 |
		int | uint | float32 | float64 | bool | complex64 | complex128
}

// ExactStringConvertible defines the constraint for exact types (no tilde) that can be converted from strings.
// This is used by StringToHookFunc to prevent type alias compilation issues.
type ExactStringConvertible interface {
	ExactPrimitiveStringConvertible | ComplexStringConvertible
}

// StringToHookFunc is a generic decode hook for converting strings.
func StringToHookFunc[T ExactStringConvertible]() DecodeHookFunc {
	return StringParserHookFunc(getParseFunc[T]())
}

// getParseFunc returns the appropriate parsing function for the given type T.
// This function encapsulates the type switch logic that determines which parser to use.
func getParseFunc[T ExactStringConvertible]() func(string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case int8:
		return genericParseWrapper[T](parseInt8)
	case uint8:
		return genericParseWrapper[T](parseUint8)
	case int16:
		return genericParseWrapper[T](parseInt16)
	case uint16:
		return genericParseWrapper[T](parseUint16)
	case int32:
		return genericParseWrapper[T](parseInt32)
	case uint32:
		return genericParseWrapper[T](parseUint32)
	case int64:
		return genericParseWrapper[T](parseInt64)
	case uint64:
		return genericParseWrapper[T](parseUint64)
	case int:
		return genericParseWrapper[T](parseInt)
	case uint:
		return genericParseWrapper[T](parseUint)
	case float32:
		return genericParseWrapper[T](parseFloat32)
	case float64:
		return genericParseWrapper[T](parseFloat64)
	case bool:
		return genericParseWrapper[T](parseBool)
	case complex64:
		return genericParseWrapper[T](parseComplex64)
	case complex128:
		return genericParseWrapper[T](parseComplex128)
	case time.Duration:
		return genericParseWrapper[T](parseDuration)
	case *url.URL:
		return genericParseWrapper[T](parseURL)
	case net.IP:
		return genericParseWrapper[T](parseIP)
	case *net.IPNet:
		return genericParseWrapper[T](parseIPNet)
	case netip.Addr:
		return genericParseWrapper[T](parseNetipAddr)
	case netip.AddrPort:
		return genericParseWrapper[T](parseNetipAddrPort)
	case netip.Prefix:
		return genericParseWrapper[T](parseNetipPrefix)
	default:
		// This should never happen due to the type constraint
		panic("unsupported type for string conversion")
	}
}

// genericParseWrapper creates a generic wrapper for the specific parse functions
func genericParseWrapper[T StringConvertible, U any](parseFunc func(string) (U, error)) func(string) (T, error) {
	return func(str string) (T, error) {
		val, err := parseFunc(str)
		return any(val).(T), err
	}
}

func parseInt8(str string) (int8, error) {
	v, err := strconv.ParseInt(str, 0, 8)
	return int8(v), wrapStrconvNumError(err)
}

func parseUint8(str string) (uint8, error) {
	v, err := strconv.ParseUint(str, 0, 8)
	return uint8(v), wrapStrconvNumError(err)
}

func parseInt16(str string) (int16, error) {
	v, err := strconv.ParseInt(str, 0, 16)
	return int16(v), wrapStrconvNumError(err)
}

func parseUint16(str string) (uint16, error) {
	v, err := strconv.ParseUint(str, 0, 16)
	return uint16(v), wrapStrconvNumError(err)
}

func parseInt32(str string) (int32, error) {
	v, err := strconv.ParseInt(str, 0, 32)
	return int32(v), wrapStrconvNumError(err)
}

func parseUint32(str string) (uint32, error) {
	v, err := strconv.ParseUint(str, 0, 32)
	return uint32(v), wrapStrconvNumError(err)
}

func parseInt64(str string) (int64, error) {
	v, err := strconv.ParseInt(str, 0, 64)
	return int64(v), wrapStrconvNumError(err)
}

func parseUint64(str string) (uint64, error) {
	v, err := strconv.ParseUint(str, 0, 64)
	return uint64(v), wrapStrconvNumError(err)
}

func parseInt(str string) (int, error) {
	v, err := strconv.ParseInt(str, 0, 0)
	return int(v), wrapStrconvNumError(err)
}

func parseUint(str string) (uint, error) {
	v, err := strconv.ParseUint(str, 0, 0)
	return uint(v), wrapStrconvNumError(err)
}

func parseFloat32(str string) (float32, error) {
	v, err := strconv.ParseFloat(str, 32)
	return float32(v), wrapStrconvNumError(err)
}

func parseFloat64(str string) (float64, error) {
	v, err := strconv.ParseFloat(str, 64)
	return v, wrapStrconvNumError(err)
}

func parseBool(str string) (bool, error) {
	v, err := strconv.ParseBool(str)
	return v, wrapStrconvNumError(err)
}

func parseComplex64(str string) (complex64, error) {
	v, err := strconv.ParseComplex(str, 64)
	return complex64(v), wrapStrconvNumError(err)
}

func parseComplex128(str string) (complex128, error) {
	v, err := strconv.ParseComplex(str, 128)
	return v, wrapStrconvNumError(err)
}

func parseDuration(str string) (time.Duration, error) {
	v, err := time.ParseDuration(str)
	return v, wrapTimeParseDurationError(err)
}

func parseURL(str string) (*url.URL, error) {
	v, err := url.Parse(str)
	return v, wrapUrlError(err)
}

func parseIP(str string) (net.IP, error) {
	v := net.ParseIP(str)
	if v == nil {
		return net.IP{}, fmt.Errorf("failed parsing ip")
	}
	return v, nil
}

func parseIPNet(str string) (*net.IPNet, error) {
	_, v, err := net.ParseCIDR(str)
	return v, wrapNetParseError(err)
}

func parseNetipAddr(str string) (netip.Addr, error) {
	v, err := netip.ParseAddr(str)
	return v, wrapNetIPParseAddrError(err)
}

func parseNetipAddrPort(str string) (netip.AddrPort, error) {
	v, err := netip.ParseAddrPort(str)
	return v, wrapNetIPParseAddrPortError(err)
}

func parseNetipPrefix(str string) (netip.Prefix, error) {
	v, err := netip.ParsePrefix(str)
	return v, wrapNetIPParsePrefixError(err)
}
