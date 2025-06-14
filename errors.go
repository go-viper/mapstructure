package mapstructure

import (
	"fmt"
	"reflect"
)

// DecodeError is a generic error type that holds information about
// a decoding error together with the name of the field that caused the error.
type DecodeError struct {
	name string
	err  error
}

func newDecodeError(name string, err error) *DecodeError {
	return &DecodeError{
		name: name,
		err:  err,
	}
}

func (e *DecodeError) Name() string {
	return e.name
}

func (e *DecodeError) Unwrap() error {
	return e.err
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("'%s' %s", e.name, e.err)
}

// ParseError is an error type that indicates a value could not be parsed
// into the expected type.
type ParseError struct {
	Expected reflect.Value
	Value    interface{}
	Err      error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("cannot parse '%s' as '%s': %s",
		e.Value, e.Expected.Type(), e.Err)
}

// UnconvertibleTypeError is an error type that indicates a value could not be
// converted to the expected type.
type UnconvertibleTypeError struct {
	Expected reflect.Value
	Value    interface{}
}

func (e *UnconvertibleTypeError) Error() string {
	return fmt.Sprintf("expected type '%s', got unconvertible type '%s', value: '%v'",
		e.Expected.Type(), reflect.TypeOf(e.Value), e.Value)
}
