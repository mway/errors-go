// Copyright (c) 2022 Matt Way
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE THE SOFTWARE.

// Package errors provides error generators and helpers.
package errors

import (
	"errors"
	"fmt"

	"go.uber.org/multierr"
)

// As is a proxy for the standard library's errors.As.
//
// As finds the first error in err's chain that matches target, and if one is
// found, sets target to that error value and returns true. Otherwise, it
// returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained
// by repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the
// value pointed to by target, or if the error has a method As(interface{})
// bool such that As(target) returns true. In the latter case, the As method is
// responsible for setting target.
//
// An error type might provide an As method so it can be treated as if it were
// a different error type.
//
// As panics if target is not a non-nil pointer to either a type that
// implements error, or to any interface type.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is is a proxy for the standard library's errors.Is.
//
// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained
// by repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See syscall.Errno.Is for an
// example in the standard library. An Is method should only shallowly compare
// err and the target and not call Unwrap on either.
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// New is a proxy for the standard library's errors.New.
//
// New returns an error that formats as the given text. Each call to New
// returns a distinct error value even if the text is identical.
func New(text string) error {
	return errors.New(text)
}

// Newf is a proxy for the standard library's fmt.Errorf.
//
// Newf formats according to a format specifier and returns the string as a
// value that satisfies error.
//
// If the format specifier includes a %w verb with an error operand, the
// returned error will implement an Unwrap method returning the operand. It is
// invalid to include more than one %w verb or to supply it with an operand
// that does not implement the error interface. The %w verb is otherwise a
// synonym for %v.
func Newf(text string, args ...any) error {
	return fmt.Errorf(text, args...)
}

// Unwrap is a proxy for the standard library's errors.Unwrap.
//
// Unwrap returns the result of calling the Unwrap method on err, if err's type
// contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Wrap returns a new error that wraps base, using msg as its error message.
// Wrap produces an error of the format "msg: base" in order to provide the
// consistent and coherent layering of errors.
//
// If base is nil, Wrap returns a nil error. If msg is an empty string, base
// is returned verbatim.
func Wrap(base error, msg string) error {
	switch {
	case base == nil:
		return nil
	case len(msg) == 0:
		return base
	default:
		return fmt.Errorf("%s: %w", msg, base)
	}
}

// Wrapf returns a new error that wraps base, using msg and args to format its
// error message. Wrap produces an error of the format "msg: base", where msg
// includes the interpolation of all sprintf placeholders and variables, in
// order to provide the consistent and coherent layering of errors.
//
// Wrapf supports wrapping errors with the %w verb.
//
// If base is nil, Wrapf returns a nil error. If msg is an empty string and
// args is empty, base is returned verbatim.
func Wrapf(base error, msg string, args ...any) error {
	switch {
	case base == nil:
		return nil
	case len(msg) == 0 && len(args) == 0:
		return base
	default:
		tmp := make([]any, len(args)+1)
		copy(tmp, args)
		tmp[len(tmp)-1] = base

		return fmt.Errorf(msg+": %w", tmp...)
	}
}

// Append returns a combined error with right appended to left. If either is
// nil, the other is returned verbatim.
func Append(left error, right error) error {
	return multierr.Append(left, right)
}

// Combine returns a combined error with each successive error appended to the
// previous errors. If an error is nil, it is omitted. If all errors are nil,
// or if errs is empty, nil is returned.
func Combine(errs ...error) error {
	return multierr.Combine(errs...)
}

// WrapAppend is syntactic sugar for Wrap(Append(left, right), msg).
func WrapAppend(left error, right error, msg string) error {
	return Wrap(Append(left, right), msg)
}

// WrapfAppend is syntactic sugar for Wrapf(Append(left, right), msg, args...).
func WrapfAppend(left error, right error, msg string, args ...any) error {
	return Wrapf(Append(left, right), msg, args...)
}

// WrapCombine is syntactic sugar for Wrap(Combine(errs...), msg).
func WrapCombine(msg string, errs ...error) error {
	return Wrap(Combine(errs...), msg)
}
