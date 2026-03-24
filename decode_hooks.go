package mapstructure

import (
	"encoding"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// safeInterface safely extracts the interface value from a reflect.Value.
// It returns nil if the value is not valid or is a nil interface.
func safeInterface(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

// decodeHookFuncTyped is an internal interface restricting the types should be a hook function type.
type decodeHookFuncTyped interface {
	// Unify returns a DecodeHookFuncValue that can be used directly in the decoder, don't return a nil plz.
	Unify() DecodeHookFuncValue
}

// unifyDecodeHook takes a raw DecodeHookFunc (an any) and turns it into a DecodeHookFuncValue(most wide form).
// if the type fails to convert we return a closure always erroring to keep the previous behaviour
func unifyDecodeHook(h DecodeHookFunc) DecodeHookFuncValue {
	// Fill in the variables into this interface and the rest is done
	// automatically using the reflect package.
	potential := []decodeHookFuncTyped{
		DecodeHookFuncType(nil),
		DecodeHookFuncKind(nil),
		DecodeHookFuncValue(nil),
	}

	v := reflect.ValueOf(h)
	vt := v.Type()
	for _, raw := range potential {
		pt := reflect.ValueOf(raw).Type()
		// Check if the provided hook is convertible to this type (same signature)
		if !vt.ConvertibleTo(pt) {
			// Not convertible, try the next one
			continue
		}

		anyV := v.Convert(pt).Interface()
		typed, ok := anyV.(decodeHookFuncTyped)
		if !ok {
			// Here should never happen since the types in potential all implement decodeHookFuncTyped.
			continue
		}

		unified := typed.Unify()
		if unified == nil {
			// Unify should never return nil, guards for further safety (maybe custom decodeHookFuncTyped)
			return func(from reflect.Value, to reflect.Value) (any, error) {
				return nil, fmt.Errorf("failed to unify decode hook: (%T).Unify() returned nil", typed)
			}
		}
		return unified
	}

	return func(from reflect.Value, to reflect.Value) (any, error) {
		return nil, errors.New("invalid decode hook signature")
	}
}

func (h DecodeHookFuncType) Unify() DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		if !from.IsValid() {
			return h(reflect.TypeOf((*any)(nil)).Elem(), to.Type(), nil)
		}
		return h(from.Type(), to.Type(), from.Interface())
	}
}

func (h DecodeHookFuncKind) Unify() DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		if !from.IsValid() {
			return h(reflect.Invalid, to.Kind(), nil)
		}
		return h(from.Kind(), to.Kind(), from.Interface())
	}
}

func (h DecodeHookFuncValue) Unify() DecodeHookFuncValue { return h }

// DecodeHookExec executes the given decode hook. This should be used
// since it'll naturally degrade to the older backwards compatible DecodeHookFunc
// that took reflect.Kind instead of reflect.Type.
func DecodeHookExec(
	raw DecodeHookFunc,
	from reflect.Value, to reflect.Value,
) (any, error) {
	unified := unifyDecodeHook(raw)
	return unified(from, to)
}

// ComposeDecodeHookFunc creates a single DecodeHookFunc that
// automatically composes multiple DecodeHookFuncs.
//
// Given hooks should be one of the three function signatures:
//   - [DecodeHookFuncType] func(reflect.Type, reflect.Type, any) (any, error)
//   - [DecodeHookFuncKind] func(reflect.Kind, reflect.Kind, any) (any, error)
//   - [DecodeHookFuncValue] func(reflect.Value, reflect.Value) (any, error)
//
// The composed funcs are called in order, with the result of the
// previous transformation.
func ComposeDecodeHookFunc(fs ...DecodeHookFunc) DecodeHookFuncValue {
	unified := make([]DecodeHookFuncValue, 0, len(fs))
	for _, f := range fs {
		unified = append(unified, unifyDecodeHook(f))
	}
	return func(f reflect.Value, t reflect.Value) (any, error) {
		var err error
		data := safeInterface(f)

		newFrom := f
		for _, c := range unified {
			data, err = c(newFrom, t)
			if err != nil {
				return nil, err
			}
			if v, ok := data.(reflect.Value); ok {
				newFrom = v
			} else if data != nil {
				newFrom = reflect.ValueOf(data)
			} else {
				// Keep newFrom as invalid (zero) Value when data is nil
				newFrom = reflect.Value{}
			}
		}

		return data, nil
	}
}

// OrComposeDecodeHookFunc executes all input hook functions until one of them returns no error. In that case its value is returned.
// If all hooks return an error, OrComposeDecodeHookFunc returns an error concatenating all error messages.
//
// Given hooks should be one of the three function signatures:
//   - [DecodeHookFuncType] func(reflect.Type, reflect.Type, any) (any, error)
//   - [DecodeHookFuncKind] func(reflect.Kind, reflect.Kind, any) (any, error)
//   - [DecodeHookFuncValue] func(reflect.Value, reflect.Value) (any, error)
func OrComposeDecodeHookFunc(ff ...DecodeHookFunc) DecodeHookFuncValue {
	unified := make([]DecodeHookFuncValue, 0, len(ff))
	for _, f := range ff {
		unified = append(unified, unifyDecodeHook(f))
	}
	return func(a, b reflect.Value) (any, error) {
		var allErrs string
		var out any
		var err error

		for _, c := range unified {
			out, err = c(a, b)
			if err != nil {
				allErrs += err.Error() + "\n"
				continue
			}

			return out, nil
		}

		return nil, errors.New(allErrs)
	}
}

// StringToSliceHookFunc returns a DecodeHookFuncType that converts
// string to []string by splitting on the given sep.
func StringToSliceHookFunc(sep string) DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.SliceOf(f) {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.Split(raw, sep), nil
	}
}

// StringToWeakSliceHookFunc brings back the old (pre-v2) behavior of [StringToSliceHookFunc].
//
// As of mapstructure v2.0.0 [StringToSliceHookFunc] checks if the return type is a string slice.
// This function removes that check.
func StringToWeakSliceHookFunc(sep string) DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.Split(raw, sep), nil
	}
}

// StringToTimeDurationHookFunc returns a DecodeHookFuncType that converts
// strings to time.Duration.
func StringToTimeDurationHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Duration(5)) {
			return data, nil
		}

		// Convert it by parsing
		d, err := time.ParseDuration(data.(string))

		return d, wrapTimeParseDurationError(err)
	}
}

// StringToTimeLocationHookFunc returns a DecodeHookFuncType that converts
// strings to *time.Location.
func StringToTimeLocationHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Local) {
			return data, nil
		}
		d, err := time.LoadLocation(data.(string))

		return d, wrapTimeParseLocationError(err)
	}
}

// StringToURLHookFunc returns a DecodeHookFuncType that converts
// strings to *url.URL.
func StringToURLHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(&url.URL{}) {
			return data, nil
		}

		// Convert it by parsing
		u, err := url.Parse(data.(string))

		return u, wrapUrlError(err)
	}
}

// StringToIPHookFunc returns a DecodeHookFuncType that converts
// strings to net.IP
func StringToIPHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(net.IP{}) {
			return data, nil
		}

		// Convert it by parsing
		ip := net.ParseIP(data.(string))
		if ip == nil {
			return net.IP{}, fmt.Errorf("failed parsing ip")
		}

		return ip, nil
	}
}

// StringToIPNetHookFunc returns a DecodeHookFuncType that converts
// strings to net.IPNet
func StringToIPNetHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(net.IPNet{}) {
			return data, nil
		}

		// Convert it by parsing
		_, net, err := net.ParseCIDR(data.(string))
		return net, wrapNetParseError(err)
	}
}

// StringToTimeHookFunc returns a DecodeHookFuncType that converts
// strings to time.Time.
func StringToTimeHookFunc(layout string) DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		// Convert it by parsing
		ti, err := time.Parse(layout, data.(string))

		return ti, wrapTimeParseError(err)
	}
}

// WeaklyTypedHook is a DecodeHookFunc which adds support for weak typing to
// the decoder.
//
// Note that this is significantly different from the WeaklyTypedInput option
// of the DecoderConfig.
func WeaklyTypedHook(
	f reflect.Kind,
	t reflect.Kind,
	data any,
) (any, error) {
	dataVal := reflect.ValueOf(data)
	switch t {
	case reflect.String:
		switch f {
		case reflect.Bool:
			if dataVal.Bool() {
				return "1", nil
			}
			return "0", nil
		case reflect.Float32:
			return strconv.FormatFloat(dataVal.Float(), 'f', -1, 64), nil
		case reflect.Int:
			return strconv.FormatInt(dataVal.Int(), 10), nil
		case reflect.Slice:
			dataType := dataVal.Type()
			elemKind := dataType.Elem().Kind()
			if elemKind == reflect.Uint8 {
				return string(dataVal.Interface().([]uint8)), nil
			}
		case reflect.Uint:
			return strconv.FormatUint(dataVal.Uint(), 10), nil
		}
	}

	return data, nil
}

func RecursiveStructToMapHookFunc() DecodeHookFuncValue {
	return func(f reflect.Value, t reflect.Value) (any, error) {
		if f.Kind() != reflect.Struct {
			return f.Interface(), nil
		}

		var i any = struct{}{}
		if t.Type() != reflect.TypeOf(&i).Elem() {
			return f.Interface(), nil
		}

		m := make(map[string]any)
		t.Set(reflect.ValueOf(m))

		return f.Interface(), nil
	}
}

// TextUnmarshallerHookFunc returns a DecodeHookFuncType that applies
// strings to the UnmarshalText function, when the target type
// implements the encoding.TextUnmarshaler interface
func TextUnmarshallerHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		result := reflect.New(t).Interface()
		unmarshaller, ok := result.(encoding.TextUnmarshaler)
		if !ok {
			return data, nil
		}
		str, ok := data.(string)
		if !ok {
			str = reflect.Indirect(reflect.ValueOf(&data)).Elem().String()
		}
		if err := unmarshaller.UnmarshalText([]byte(str)); err != nil {
			return nil, err
		}
		return result, nil
	}
}

// StringToNetIPAddrHookFunc returns a DecodeHookFuncType that converts
// strings to netip.Addr.
func StringToNetIPAddrHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(netip.Addr{}) {
			return data, nil
		}

		// Convert it by parsing
		addr, err := netip.ParseAddr(data.(string))

		return addr, wrapNetIPParseAddrError(err)
	}
}

// StringToNetIPAddrPortHookFunc returns a DecodeHookFuncType that converts
// strings to netip.AddrPort.
func StringToNetIPAddrPortHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(netip.AddrPort{}) {
			return data, nil
		}

		// Convert it by parsing
		addrPort, err := netip.ParseAddrPort(data.(string))

		return addrPort, wrapNetIPParseAddrPortError(err)
	}
}

// StringToNetIPPrefixHookFunc returns a DecodeHookFuncType that converts
// strings to netip.Prefix.
func StringToNetIPPrefixHookFunc() DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(netip.Prefix{}) {
			return data, nil
		}

		// Convert it by parsing
		prefix, err := netip.ParsePrefix(data.(string))

		return prefix, wrapNetIPParsePrefixError(err)
	}
}

// StringToBasicTypeHookFunc returns a DecodeHookFuncValue that converts
// strings to basic types.
// int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint, float32, float64, bool, byte, rune, complex64, complex128
func StringToBasicTypeHookFunc() DecodeHookFuncValue {
	return ComposeDecodeHookFunc(
		StringToInt8HookFunc(),
		StringToUint8HookFunc(),
		StringToInt16HookFunc(),
		StringToUint16HookFunc(),
		StringToInt32HookFunc(),
		StringToUint32HookFunc(),
		StringToInt64HookFunc(),
		StringToUint64HookFunc(),
		StringToIntHookFunc(),
		StringToUintHookFunc(),
		StringToFloat32HookFunc(),
		StringToFloat64HookFunc(),
		StringToBoolHookFunc(),
		// byte and rune are aliases for uint8 and int32 respectively
		// StringToByteHookFunc(),
		// StringToRuneHookFunc(),
		StringToComplex64HookFunc(),
		StringToComplex128HookFunc(),
	)
}

// StringToInt8HookFunc returns a DecodeHookFuncType that converts
// strings to int8.
func StringToInt8HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Int8 {
			return data, nil
		}

		// Convert it by parsing
		i64, err := strconv.ParseInt(data.(string), 0, 8)
		return int8(i64), wrapStrconvNumError(err)
	}
}

// StringToUint8HookFunc returns a DecodeHookFuncType that converts
// strings to uint8.
func StringToUint8HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Uint8 {
			return data, nil
		}

		// Convert it by parsing
		u64, err := strconv.ParseUint(data.(string), 0, 8)
		return uint8(u64), wrapStrconvNumError(err)
	}
}

// StringToInt16HookFunc returns a DecodeHookFuncType that converts
// strings to int16.
func StringToInt16HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Int16 {
			return data, nil
		}

		// Convert it by parsing
		i64, err := strconv.ParseInt(data.(string), 0, 16)
		return int16(i64), wrapStrconvNumError(err)
	}
}

// StringToUint16HookFunc returns a DecodeHookFuncType that converts
// strings to uint16.
func StringToUint16HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Uint16 {
			return data, nil
		}

		// Convert it by parsing
		u64, err := strconv.ParseUint(data.(string), 0, 16)
		return uint16(u64), wrapStrconvNumError(err)
	}
}

// StringToInt32HookFunc returns a DecodeHookFuncType that converts
// strings to int32.
func StringToInt32HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Int32 {
			return data, nil
		}

		// Convert it by parsing
		i64, err := strconv.ParseInt(data.(string), 0, 32)
		return int32(i64), wrapStrconvNumError(err)
	}
}

// StringToUint32HookFunc returns a DecodeHookFuncType that converts
// strings to uint32.
func StringToUint32HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Uint32 {
			return data, nil
		}

		// Convert it by parsing
		u64, err := strconv.ParseUint(data.(string), 0, 32)
		return uint32(u64), wrapStrconvNumError(err)
	}
}

// StringToInt64HookFunc returns a DecodeHookFuncType that converts
// strings to int64.
func StringToInt64HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Int64 {
			return data, nil
		}

		// Convert it by parsing
		i64, err := strconv.ParseInt(data.(string), 0, 64)
		return int64(i64), wrapStrconvNumError(err)
	}
}

// StringToUint64HookFunc returns a DecodeHookFuncType that converts
// strings to uint64.
func StringToUint64HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Uint64 {
			return data, nil
		}

		// Convert it by parsing
		u64, err := strconv.ParseUint(data.(string), 0, 64)
		return uint64(u64), wrapStrconvNumError(err)
	}
}

// StringToIntHookFunc returns a DecodeHookFuncType that converts
// strings to int.
func StringToIntHookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Int {
			return data, nil
		}

		// Convert it by parsing
		i64, err := strconv.ParseInt(data.(string), 0, 0)
		return int(i64), wrapStrconvNumError(err)
	}
}

// StringToUintHookFunc returns a DecodeHookFuncType that converts
// strings to uint.
func StringToUintHookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Uint {
			return data, nil
		}

		// Convert it by parsing
		u64, err := strconv.ParseUint(data.(string), 0, 0)
		return uint(u64), wrapStrconvNumError(err)
	}
}

// StringToFloat32HookFunc returns a DecodeHookFuncType that converts
// strings to float32.
func StringToFloat32HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Float32 {
			return data, nil
		}

		// Convert it by parsing
		f64, err := strconv.ParseFloat(data.(string), 32)
		return float32(f64), wrapStrconvNumError(err)
	}
}

// StringToFloat64HookFunc returns a DecodeHookFuncType that converts
// strings to float64.
func StringToFloat64HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Float64 {
			return data, nil
		}

		// Convert it by parsing
		f64, err := strconv.ParseFloat(data.(string), 64)
		return f64, wrapStrconvNumError(err)
	}
}

// StringToBoolHookFunc returns a DecodeHookFuncType that converts
// strings to bool.
func StringToBoolHookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Bool {
			return data, nil
		}

		// Convert it by parsing
		b, err := strconv.ParseBool(data.(string))
		return b, wrapStrconvNumError(err)
	}
}

// StringToByteHookFunc returns a DecodeHookFuncType that converts
// strings to byte.
func StringToByteHookFunc() DecodeHookFuncType {
	return StringToUint8HookFunc()
}

// StringToRuneHookFunc returns a DecodeHookFuncType that converts
// strings to rune.
func StringToRuneHookFunc() DecodeHookFuncType {
	return StringToInt32HookFunc()
}

// StringToComplex64HookFunc returns a DecodeHookFuncType that converts
// strings to complex64.
func StringToComplex64HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Complex64 {
			return data, nil
		}

		// Convert it by parsing
		c128, err := strconv.ParseComplex(data.(string), 64)
		return complex64(c128), wrapStrconvNumError(err)
	}
}

// StringToComplex128HookFunc returns a DecodeHookFuncType that converts
// strings to complex128.
func StringToComplex128HookFunc() DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Complex128 {
			return data, nil
		}

		// Convert it by parsing
		c128, err := strconv.ParseComplex(data.(string), 128)
		return c128, wrapStrconvNumError(err)
	}
}
