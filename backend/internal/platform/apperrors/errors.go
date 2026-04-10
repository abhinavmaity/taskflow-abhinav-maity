package apperrors

import (
	"errors"
	"fmt"
)

type Kind string

const (
	KindValidation   Kind = "validation"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindInternal     Kind = "internal"
)

type Error struct {
	Kind   Kind
	Msg    string
	Fields map[string]string
	Err    error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Kind)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewValidation(fields map[string]string) *Error {
	return &Error{
		Kind:   KindValidation,
		Msg:    "validation failed",
		Fields: fields,
	}
}

func NewUnauthorized() *Error {
	return &Error{
		Kind: KindUnauthorized,
		Msg:  "unauthorized",
	}
}

func NewForbidden() *Error {
	return &Error{
		Kind: KindForbidden,
		Msg:  "forbidden",
	}
}

func NewNotFound() *Error {
	return &Error{
		Kind: KindNotFound,
		Msg:  "not found",
	}
}

func WrapInternal(err error) *Error {
	return &Error{
		Kind: KindInternal,
		Msg:  "internal server error",
		Err:  err,
	}
}

func As(err error) (*Error, bool) {
	var appErr *Error
	ok := errors.As(err, &appErr)
	return appErr, ok
}

func Wrapf(kind Kind, err error, format string, args ...any) *Error {
	return &Error{
		Kind: kind,
		Msg:  fmt.Sprintf(format, args...),
		Err:  err,
	}
}
