package mapstructure

import (
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestDecodeHooksNamedStrings(t *testing.T) {
	type namedString string
	tests := []struct {
		name  string
		hook  DecodeHookFunc
		input string
		want  any
	}{
		{"int8", StringToInt8HookFunc(), "12", int8(12)},
		{"uint8", StringToUint8HookFunc(), "12", uint8(12)},
		{"int16", StringToInt16HookFunc(), "12", int16(12)},
		{"uint16", StringToUint16HookFunc(), "12", uint16(12)},
		{"int32", StringToInt32HookFunc(), "12", int32(12)},
		{"uint32", StringToUint32HookFunc(), "12", uint32(12)},
		{"int64", StringToInt64HookFunc(), "12", int64(12)},
		{"uint64", StringToUint64HookFunc(), "12", uint64(12)},
		{"int", StringToIntHookFunc(), "12", int(12)},
		{"uint", StringToUintHookFunc(), "12", uint(12)},
		{"float32", StringToFloat32HookFunc(), "1.5", float32(1.5)},
		{"float64", StringToFloat64HookFunc(), "1.5", float64(1.5)},
		{"bool", StringToBoolHookFunc(), "true", true},
		{"complex64", StringToComplex64HookFunc(), "1+2i", complex64(1 + 2i)},
		{"complex128", StringToComplex128HookFunc(), "1+2i", complex128(1 + 2i)},
		{"basic", StringToBasicTypeHookFunc(), "12", int(12)},
		{"duration", StringToTimeDurationHookFunc(), "2s", 2 * time.Second},
		{"time", StringToTimeHookFunc(time.RFC3339), "2024-01-01T00:00:00Z", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"location", StringToTimeLocationHookFunc(), "UTC", time.UTC},
		{"url", StringToURLHookFunc(), "https://example.com/path", &url.URL{Scheme: "https", Host: "example.com", Path: "/path"}},
		{"ip", StringToIPHookFunc(), "127.0.0.1", net.ParseIP("127.0.0.1")},
		{"addr", StringToNetIPAddrHookFunc(), "127.0.0.1", netip.MustParseAddr("127.0.0.1")},
		{"addrport", StringToNetIPAddrPortHookFunc(), "127.0.0.1:80", netip.MustParseAddrPort("127.0.0.1:80")},
		{"prefix", StringToNetIPPrefixHookFunc(), "127.0.0.0/8", netip.MustParsePrefix("127.0.0.0/8")},
		{"weak slice", StringToWeakSliceHookFunc(","), "a,b", []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range []any{tt.input, namedString(tt.input)} {
				t.Run(reflect.TypeOf(input).String(), func(t *testing.T) {
					got, err := DecodeHookExec(tt.hook, reflect.ValueOf(input), reflect.ValueOf(tt.want))
					if err != nil || !reflect.DeepEqual(got, tt.want) {
						t.Fatalf("got %#v, %v; want %#v", got, err, tt.want)
					}
				})
			}
		})
	}
}

func TestDecodeNamedStringToStructuredTypes(t *testing.T) {
	type namedString string
	for _, tt := range []struct {
		name  string
		hook  DecodeHookFunc
		input namedString
		want  any
	}{
		{"slice", StringToSliceHookFunc(","), "a,b", []namedString{"a", "b"}},
		{"empty slice", StringToSliceHookFunc(","), "", []namedString{}},
		{"IP network", StringToIPNetHookFunc(), "192.0.2.0/24", net.IPNet{IP: net.IP{192, 0, 2, 0}, Mask: net.CIDRMask(24, 32)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := reflect.New(reflect.TypeOf(tt.want))
			decoder, err := NewDecoder(&DecoderConfig{DecodeHook: tt.hook, Result: result.Interface()})
			if err != nil {
				t.Fatal(err)
			}
			if err := decoder.Decode(tt.input); err != nil {
				t.Fatal(err)
			}
			if got := result.Elem().Interface(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v; want %#v", got, tt.want)
			}
		})
	}
}
