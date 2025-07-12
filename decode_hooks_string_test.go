package mapstructure

import (
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStringParserHookFunc(t *testing.T) {
	t.Run("CustomInt32Parser", func(t *testing.T) {
		customParser := func(s string) (int32, error) {
			// Custom parser that multiplies by 2
			val, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return int32(val * 2), nil
		}

		hook := StringParserHookFunc(customParser)

		strValue := reflect.ValueOf("21")
		int32Value := reflect.ValueOf(int32(0))

		result, err := DecodeHookExec(hook, strValue, int32Value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := int32(42)
		if result != expected {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("CustomStringToURL", func(t *testing.T) {
		customParser := func(s string) (*url.URL, error) {
			// Add https:// prefix if not present
			if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
				s = "https://" + s
			}
			return url.Parse(s)
		}

		hook := StringParserHookFunc(customParser)

		strValue := reflect.ValueOf("example.com")
		urlValue := reflect.ValueOf(&url.URL{})

		result, err := DecodeHookExec(hook, strValue, urlValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := &url.URL{Scheme: "https", Host: "example.com"}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("NonStringSource", func(t *testing.T) {
		hook := StringParserHookFunc(func(s string) (int32, error) {
			val, err := strconv.ParseInt(s, 10, 32)
			return int32(val), err
		})

		intValue := reflect.ValueOf(42)
		int32Value := reflect.ValueOf(int32(0))

		result, err := DecodeHookExec(hook, intValue, int32Value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return original data unchanged
		if result != 42 {
			t.Fatalf("expected %v, got %v", 42, result)
		}
	})

	t.Run("WrongTargetType", func(t *testing.T) {
		hook := StringParserHookFunc(func(s string) (int32, error) {
			val, err := strconv.ParseInt(s, 10, 32)
			return int32(val), err
		})

		strValue := reflect.ValueOf("42")
		int64Value := reflect.ValueOf(int64(0))

		result, err := DecodeHookExec(hook, strValue, int64Value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return original data unchanged
		if result != "42" {
			t.Fatalf("expected %v, got %v", "42", result)
		}
	})

	t.Run("ParseError", func(t *testing.T) {
		hook := StringParserHookFunc(func(s string) (int32, error) {
			val, err := strconv.ParseInt(s, 10, 32)
			return int32(val), err
		})

		strValue := reflect.ValueOf("not-a-number")
		int32Value := reflect.ValueOf(int32(0))

		_, err := DecodeHookExec(hook, strValue, int32Value)
		if err == nil {
			t.Fatal("expected error but got none")
		}
	})

	t.Run("IPNetSpecialCase", func(t *testing.T) {
		hook := StringParserHookFunc(func(s string) (*net.IPNet, error) {
			_, ipnet, err := net.ParseCIDR(s)
			return ipnet, err
		})

		strValue := reflect.ValueOf("192.168.1.0/24")
		ipnetValue := reflect.ValueOf(net.IPNet{})

		result, err := DecodeHookExec(hook, strValue, ipnetValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedIPNet := &net.IPNet{
			IP:   net.IPv4(192, 168, 1, 0),
			Mask: net.CIDRMask(24, 32),
		}

		resultIPNet, ok := result.(*net.IPNet)
		if !ok {
			t.Fatalf("expected *net.IPNet, got %T", result)
		}

		if !resultIPNet.IP.Equal(expectedIPNet.IP) || !reflect.DeepEqual(resultIPNet.Mask, expectedIPNet.Mask) {
			t.Fatalf("expected %v, got %v", expectedIPNet, resultIPNet)
		}
	})
}

func TestStringToHookFunc(t *testing.T) {
	t.Run("Int32", func(t *testing.T) {
		hook := StringToHookFunc[int32]()

		int32Value := reflect.ValueOf(int32(0))

		cases := []struct {
			input    string
			expected int32
			hasError bool
		}{
			{"42", 42, false},
			{"-42", -42, false},
			{"0", 0, false},
			{"0x2a", 42, false},
			{"052", 42, false},
			{"0b101010", 42, false},
			{"2147483647", 2147483647, false},
			{"-2147483648", -2147483648, false},
			{"2147483648", 0, true},  // overflow
			{"-2147483649", 0, true}, // underflow
			{"42.5", 0, true},        // float
			{"not-a-number", 0, true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, int32Value)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if result != tc.expected {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("Float64", func(t *testing.T) {
		hook := StringToHookFunc[float64]()

		float64Value := reflect.ValueOf(float64(0))

		cases := []struct {
			input    string
			expected float64
			hasError bool
		}{
			{"42.5", 42.5, false},
			{"-42.5", -42.5, false},
			{"0", 0, false},
			{"0.0", 0.0, false},
			{"3.14159", 3.14159, false},
			{"1e10", 1e10, false},
			{"1.5e-10", 1.5e-10, false},
			{"not-a-number", 0, true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, float64Value)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if result != tc.expected {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("Bool", func(t *testing.T) {
		hook := StringToHookFunc[bool]()

		boolValue := reflect.ValueOf(false)

		cases := []struct {
			input    string
			expected bool
			hasError bool
		}{
			{"true", true, false},
			{"false", false, false},
			{"1", true, false},
			{"0", false, false},
			{"t", true, false},
			{"f", false, false},
			{"T", true, false},
			{"F", false, false},
			{"TRUE", true, false},
			{"FALSE", false, false},
			{"True", true, false},
			{"False", false, false},
			{"yes", false, true},
			{"no", false, true},
			{"invalid", false, true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, boolValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if result != tc.expected {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("Duration", func(t *testing.T) {
		hook := StringToHookFunc[time.Duration]()

		durationValue := reflect.ValueOf(time.Duration(0))

		cases := []struct {
			input    string
			expected time.Duration
			hasError bool
		}{
			{"1h", time.Hour, false},
			{"30m", 30 * time.Minute, false},
			{"45s", 45 * time.Second, false},
			{"1h30m45s", time.Hour + 30*time.Minute + 45*time.Second, false},
			{"1000ms", time.Second, false},
			{"1000000us", time.Second, false},
			{"1000000000ns", time.Second, false},
			{"0", 0, false},
			{"invalid", 0, true},
			{"1", 0, true}, // missing unit
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, durationValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if result != tc.expected {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("URL", func(t *testing.T) {
		hook := StringToHookFunc[*url.URL]()

		urlValue := reflect.ValueOf(&url.URL{})

		cases := []struct {
			input    string
			expected *url.URL
			hasError bool
		}{
			{
				"https://example.com",
				&url.URL{Scheme: "https", Host: "example.com"},
				false,
			},
			{
				"http://example.com:8080/path?query=value",
				&url.URL{
					Scheme:   "http",
					Host:     "example.com:8080",
					Path:     "/path",
					RawQuery: "query=value",
				},
				false,
			},
			{
				"ftp://user:pass@example.com/file.txt",
				&url.URL{
					Scheme: "ftp",
					User:   url.UserPassword("user", "pass"),
					Host:   "example.com",
					Path:   "/file.txt",
				},
				false,
			},
			{
				"example.com", // relative URL
				&url.URL{Path: "example.com"},
				false,
			},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, urlValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("NetIP", func(t *testing.T) {
		hook := StringToHookFunc[net.IP]()

		ipValue := reflect.ValueOf(net.IP{})

		cases := []struct {
			input    string
			expected net.IP
			hasError bool
		}{
			{"192.168.1.1", net.IPv4(192, 168, 1, 1), false},
			{"::1", net.IPv6loopback, false},
			{"2001:db8::1", net.ParseIP("2001:db8::1"), false},
			{"invalid-ip", net.IP{}, true},
			{"", net.IP{}, true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, ipValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("NetIPNet", func(t *testing.T) {
		hook := StringToHookFunc[*net.IPNet]()

		ipnetValue := reflect.ValueOf(net.IPNet{})

		cases := []struct {
			input    string
			hasError bool
		}{
			{"192.168.1.0/24", false},
			{"10.0.0.0/8", false},
			{"2001:db8::/32", false},
			{"192.168.1.1", true},    // single IP, not CIDR
			{"192.168.1.0/33", true}, // invalid mask
			{"invalid", true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, ipnetValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			// Verify it's a valid IPNet
			if result == nil {
				t.Fatalf("case %d: expected non-nil result", i)
			}

			ipnet, ok := result.(*net.IPNet)
			if !ok {
				t.Fatalf("case %d: expected *net.IPNet, got %T", i, result)
			}

			if ipnet.IP == nil || ipnet.Mask == nil {
				t.Fatalf("case %d: invalid IPNet: %v", i, ipnet)
			}
		}
	})

	t.Run("NetipAddr", func(t *testing.T) {
		hook := StringToHookFunc[netip.Addr]()

		addrValue := reflect.ValueOf(netip.Addr{})

		cases := []struct {
			input    string
			hasError bool
		}{
			{"192.168.1.1", false},
			{"::1", false},
			{"2001:db8::1", false},
			{"invalid-ip", true},
			{"", true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, addrValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			// Verify it's a valid netip.Addr
			addr, ok := result.(netip.Addr)
			if !ok {
				t.Fatalf("case %d: expected netip.Addr, got %T", i, result)
			}

			if !addr.IsValid() {
				t.Fatalf("case %d: invalid netip.Addr: %v", i, addr)
			}
		}
	})

	t.Run("NetipAddrPort", func(t *testing.T) {
		hook := StringToHookFunc[netip.AddrPort]()

		addrPortValue := reflect.ValueOf(netip.AddrPort{})

		cases := []struct {
			input    string
			hasError bool
		}{
			{"192.168.1.1:8080", false},
			{"[::1]:8080", false},
			{"[2001:db8::1]:443", false},
			{"192.168.1.1", true},       // missing port
			{"192.168.1.1:99999", true}, // invalid port
			{"invalid", true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, addrPortValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			// Verify it's a valid netip.AddrPort
			addrPort, ok := result.(netip.AddrPort)
			if !ok {
				t.Fatalf("case %d: expected netip.AddrPort, got %T", i, result)
			}

			if !addrPort.IsValid() {
				t.Fatalf("case %d: invalid netip.AddrPort: %v", i, addrPort)
			}
		}
	})

	t.Run("NetipPrefix", func(t *testing.T) {
		hook := StringToHookFunc[netip.Prefix]()

		prefixValue := reflect.ValueOf(netip.Prefix{})

		cases := []struct {
			input    string
			hasError bool
		}{
			{"192.168.1.0/24", false},
			{"10.0.0.0/8", false},
			{"2001:db8::/32", false},
			{"192.168.1.1/32", false},
			{"192.168.1.0/33", true}, // invalid mask for IPv4
			{"2001:db8::/129", true}, // invalid mask for IPv6
			{"192.168.1.1", true},    // missing prefix length
			{"invalid", true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, prefixValue)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			// Verify it's a valid netip.Prefix
			prefix, ok := result.(netip.Prefix)
			if !ok {
				t.Fatalf("case %d: expected netip.Prefix, got %T", i, result)
			}

			if !prefix.IsValid() {
				t.Fatalf("case %d: invalid netip.Prefix: %v", i, prefix)
			}
		}
	})

	t.Run("Complex64", func(t *testing.T) {
		hook := StringToHookFunc[complex64]()

		complex64Value := reflect.ValueOf(complex64(0))

		cases := []struct {
			input    string
			expected complex64
			hasError bool
		}{
			{"1+2i", complex64(1 + 2i), false},
			{"3-4i", complex64(3 - 4i), false},
			{"5", complex64(5 + 0i), false},
			{"0", complex64(0 + 0i), false},
			{"-1", complex64(-1 + 0i), false},
			{"0+1i", complex64(0 + 1i), false},
			{"0-1i", complex64(0 - 1i), false},
			{"invalid", complex64(0), true},
			{"1+", complex64(0), true},
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			result, err := DecodeHookExec(hook, inputValue, complex64Value)

			if tc.hasError {
				if err == nil {
					t.Fatalf("case %d: expected error but got none", i)
				}
				continue
			}

			if err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}

			if result != tc.expected {
				t.Fatalf("case %d: expected %v, got %v", i, tc.expected, result)
			}
		}
	})

	t.Run("NonStringSource", func(t *testing.T) {
		hook := StringToHookFunc[int32]()

		intValue := reflect.ValueOf(42)
		int32Value := reflect.ValueOf(int32(0))

		result, err := DecodeHookExec(hook, intValue, int32Value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return original data unchanged
		if result != 42 {
			t.Fatalf("expected %v, got %v", 42, result)
		}
	})

	t.Run("WrongTargetType", func(t *testing.T) {
		hook := StringToHookFunc[int32]()

		strValue := reflect.ValueOf("42")
		int64Value := reflect.ValueOf(int64(0))

		result, err := DecodeHookExec(hook, strValue, int64Value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return original data unchanged
		if result != "42" {
			t.Fatalf("expected %v, got %v", "42", result)
		}
	})
}

func TestStringParserHookFuncWithTypeAlias(t *testing.T) {
	// Test with type alias to ensure tilde types work correctly
	type MyInt int32

	customParser := func(s string) (MyInt, error) {
		val, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0, err
		}
		return MyInt(val), nil
	}

	hook := StringParserHookFunc(customParser)

	strValue := reflect.ValueOf("42")
	myIntValue := reflect.ValueOf(MyInt(0))

	result, err := DecodeHookExec(hook, strValue, myIntValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := MyInt(42)
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestStringToHookFuncEdgeCases(t *testing.T) {
	t.Run("UintOverflow", func(t *testing.T) {
		hook := StringToHookFunc[uint8]()
		uintValue := reflect.ValueOf(uint8(0))

		cases := []struct {
			input    string
			hasError bool
		}{
			{"0", false},
			{"255", false},
			{"256", true}, // overflow
			{"-1", true},  // negative
		}

		for i, tc := range cases {
			inputValue := reflect.ValueOf(tc.input)
			_, err := DecodeHookExec(hook, inputValue, uintValue)

			if tc.hasError && err == nil {
				t.Fatalf("case %d: expected error but got none", i)
			}
			if !tc.hasError && err != nil {
				t.Fatalf("case %d: unexpected error: %v", i, err)
			}
		}
	})

	t.Run("EmptyStringHandling", func(t *testing.T) {
		t.Run("Int", func(t *testing.T) {
			hook := StringToHookFunc[int]()
			intValue := reflect.ValueOf(int(0))

			inputValue := reflect.ValueOf("")
			_, err := DecodeHookExec(hook, inputValue, intValue)
			if err == nil {
				t.Fatal("expected error for empty string")
			}
		})

		t.Run("Bool", func(t *testing.T) {
			hook := StringToHookFunc[bool]()
			boolValue := reflect.ValueOf(false)

			inputValue := reflect.ValueOf("")
			_, err := DecodeHookExec(hook, inputValue, boolValue)
			if err == nil {
				t.Fatal("expected error for empty string")
			}
		})

		t.Run("URL", func(t *testing.T) {
			hook := StringToHookFunc[*url.URL]()
			urlValue := reflect.ValueOf(&url.URL{})

			inputValue := reflect.ValueOf("")
			result, err := DecodeHookExec(hook, inputValue, urlValue)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Empty string should parse to empty URL
			expected := &url.URL{}
			if !reflect.DeepEqual(result, expected) {
				t.Fatalf("expected %v, got %v", expected, result)
			}
		})
	})

	t.Run("AllNumericTypes", func(t *testing.T) {
		// Test all supported numeric types work correctly
		testCases := []struct {
			name     string
			hookFunc DecodeHookFunc
			target   reflect.Value
			input    string
			expected interface{}
		}{
			{"int8", StringToHookFunc[int8](), reflect.ValueOf(int8(0)), "42", int8(42)},
			{"uint8", StringToHookFunc[uint8](), reflect.ValueOf(uint8(0)), "42", uint8(42)},
			{"int16", StringToHookFunc[int16](), reflect.ValueOf(int16(0)), "42", int16(42)},
			{"uint16", StringToHookFunc[uint16](), reflect.ValueOf(uint16(0)), "42", uint16(42)},
			{"int32", StringToHookFunc[int32](), reflect.ValueOf(int32(0)), "42", int32(42)},
			{"uint32", StringToHookFunc[uint32](), reflect.ValueOf(uint32(0)), "42", uint32(42)},
			{"int64", StringToHookFunc[int64](), reflect.ValueOf(int64(0)), "42", int64(42)},
			{"uint64", StringToHookFunc[uint64](), reflect.ValueOf(uint64(0)), "42", uint64(42)},
			{"int", StringToHookFunc[int](), reflect.ValueOf(int(0)), "42", int(42)},
			{"uint", StringToHookFunc[uint](), reflect.ValueOf(uint(0)), "42", uint(42)},
			{"float32", StringToHookFunc[float32](), reflect.ValueOf(float32(0)), "42.5", float32(42.5)},
			{"float64", StringToHookFunc[float64](), reflect.ValueOf(float64(0)), "42.5", float64(42.5)},
			{"complex64", StringToHookFunc[complex64](), reflect.ValueOf(complex64(0)), "1+2i", complex64(1 + 2i)},
			{"complex128", StringToHookFunc[complex128](), reflect.ValueOf(complex128(0)), "1+2i", complex128(1 + 2i)},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				inputValue := reflect.ValueOf(tc.input)
				result, err := DecodeHookExec(tc.hookFunc, inputValue, tc.target)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result != tc.expected {
					t.Fatalf("expected %v (%T), got %v (%T)", tc.expected, tc.expected, result, result)
				}
			})
		}
	})
}
