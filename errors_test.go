package mapstructure

import (
	"errors"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestDecodeError(t *testing.T) {
	innerErr := errors.New("underlying error")
	err := newDecodeError("FieldA", innerErr)

	if err.Name() != "FieldA" {
		t.Fatalf("expected Name() to be 'FieldA', got %q", err.Name())
	}
	if err.Unwrap() != innerErr {
		t.Fatalf("expected Unwrap() to return inner error, got %v", err.Unwrap())
	}
	if err.Error() != "'FieldA' underlying error" {
		t.Fatalf("expected Error() to be \"'FieldA' underlying error\", got %q", err.Error())
	}

	var mErr Error
	if !errors.As(err, &mErr) {
		t.Fatal("expected DecodeError to implement Error interface")
	}
}

func TestParseError(t *testing.T) {
	val := reflect.ValueOf(123)
	err := &ParseError{
		Expected: val,
		Value:    "abc",
		Err:      errors.New("invalid syntax"),
	}

	if err.Error() != "cannot parse value as 'int': invalid syntax" {
		t.Fatalf("unexpected Error() string: %q", err.Error())
	}

	var mErr Error
	if !errors.As(err, &mErr) {
		t.Fatal("expected ParseError to implement Error interface")
	}
}

func TestUnconvertibleTypeError(t *testing.T) {
	val := reflect.ValueOf("string")
	err := &UnconvertibleTypeError{
		Expected: val,
		Value:    123,
	}

	if err.Error() != "expected type 'string', got unconvertible type 'int'" {
		t.Fatalf("unexpected Error() string: %q", err.Error())
	}

	var mErr Error
	if !errors.As(err, &mErr) {
		t.Fatal("expected UnconvertibleTypeError to implement Error interface")
	}
}

func TestWrapStrconvNumError(t *testing.T) {
	if wrapStrconvNumError(nil) != nil {
		t.Fatal("expected nil")
	}

	genericErr := errors.New("generic")
	if wrapStrconvNumError(genericErr) != genericErr {
		t.Fatal("expected generic error unchanged")
	}

	numErr := &strconv.NumError{
		Func: "ParseInt",
		Num:  "abc",
		Err:  strconv.ErrSyntax,
	}
	wrapped := wrapStrconvNumError(numErr)
	if wrapped.Error() != "strconv.ParseInt: invalid syntax" {
		t.Fatalf("unexpected wrapped error: %q", wrapped.Error())
	}
	if errors.Unwrap(wrapped) != numErr {
		t.Fatal("expected unwrapped error to match numErr")
	}
}

func TestWrapUrlError(t *testing.T) {
	if wrapUrlError(nil) != nil {
		t.Fatal("expected nil")
	}

	urlErr := &url.Error{
		Op:  "parse",
		URL: ":foo",
		Err: errors.New("missing protocol scheme"),
	}
	wrapped := wrapUrlError(urlErr)
	if wrapped.Error() != "missing protocol scheme" {
		t.Fatalf("unexpected wrapped error: %q", wrapped.Error())
	}
	if errors.Unwrap(wrapped) != urlErr {
		t.Fatal("expected unwrapped error to match urlErr")
	}
}

func TestWrapNetParseError(t *testing.T) {
	if wrapNetParseError(nil) != nil {
		t.Fatal("expected nil")
	}

	netErr := &net.ParseError{
		Type: "IP address",
		Text: "invalid",
	}
	wrapped := wrapNetParseError(netErr)
	if wrapped.Error() != "invalid IP address" {
		t.Fatalf("unexpected wrapped error: %q", wrapped.Error())
	}
	if errors.Unwrap(wrapped) != netErr {
		t.Fatal("expected unwrapped error to match netErr")
	}
}

func TestWrapTimeParseError(t *testing.T) {
	if wrapTimeParseError(nil) != nil {
		t.Fatal("expected nil")
	}

	timeErr := &time.ParseError{
		Layout:     time.RFC3339,
		LayoutElem: "2006",
		Value:      "abc",
	}
	wrapped := wrapTimeParseError(timeErr)
	if wrapped.Error() != `parsing time as "2006-01-02T15:04:05Z07:00": cannot parse as "2006"` {
		t.Fatalf("unexpected wrapped error: %q", wrapped.Error())
	}
	if errors.Unwrap(wrapped) != timeErr {
		t.Fatal("expected unwrapped error to match timeErr")
	}

	timeErrWithMessage := &time.ParseError{
		Message: ": extra text: foo",
	}
	wrappedMsg := wrapTimeParseError(timeErrWithMessage)
	if wrappedMsg.Error() != "parsing time : extra text: foo" {
		t.Fatalf("unexpected wrapped error: %q", wrappedMsg.Error())
	}
}

func TestWrapNetIPParseErrors(t *testing.T) {
	if wrapNetIPParseAddrError(nil) != nil {
		t.Fatal("expected nil")
	}
	if wrapNetIPParseAddrPortError(nil) != nil {
		t.Fatal("expected nil")
	}
	if wrapNetIPParsePrefixError(nil) != nil {
		t.Fatal("expected nil")
	}

	addrErr := errors.New("ParseAddr: invalid IP address")
	if wrapNetIPParseAddrError(addrErr).Error() != "ParseAddr: invalid IP address" {
		t.Fatalf("unexpected addr error: %q", wrapNetIPParseAddrError(addrErr).Error())
	}

	portErr := errors.New("invalid port \"foo\"")
	if wrapNetIPParseAddrPortError(portErr).Error() != "invalid port" {
		t.Fatalf("unexpected port error: %q", wrapNetIPParseAddrPortError(portErr).Error())
	}

	ipPortErr := errors.New("invalid ip:port \"foo\"")
	if wrapNetIPParseAddrPortError(ipPortErr).Error() != "invalid ip:port" {
		t.Fatalf("unexpected ip:port error: %q", wrapNetIPParseAddrPortError(ipPortErr).Error())
	}

	prefixErr := errors.New("netip.ParsePrefix: invalid CIDR prefix")
	if wrapNetIPParsePrefixError(prefixErr).Error() != "netip.ParsePrefix: invalid CIDR prefix" {
		t.Fatalf("unexpected prefix error: %q", wrapNetIPParsePrefixError(prefixErr).Error())
	}
}

func TestWrapTimeParseDurationAndLocationError(t *testing.T) {
	if wrapTimeParseDurationError(nil) != nil {
		t.Fatal("expected nil")
	}
	if wrapTimeParseLocationError(nil) != nil {
		t.Fatal("expected nil")
	}

	durErr := errors.New("time: unknown unit \"x\" in sequence")
	if wrapTimeParseDurationError(durErr).Error() != "time: unknown unit" {
		t.Fatalf("unexpected duration error: %q", wrapTimeParseDurationError(durErr).Error())
	}

	locErr := errors.New("unknown time zone Foo/Bar")
	wrappedLoc := wrapTimeParseLocationError(locErr)
	if !errors.Is(wrappedLoc, locErr) {
		t.Fatal("expected wrapped location error to unwrap to locErr")
	}
}
