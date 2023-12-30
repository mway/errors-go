// Copyright (c) 2023 Matt Way
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
	"sync"
)

// An ErrorFunc is a function that returns an error.
type ErrorFunc = func() error

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

// Join combines all given errors into a single error. Any nil values are
// discarded.
func Join(errs ...error) error {
	return errors.Join(errs...)
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

// JoinFuncs evaluates fns serially, joining all non-nil return values and
// returning the resulting error. If fns is empty or if all fns return nil,
// nil is returned; if only one error is produced, it is returned verbatim.
// Otherwise, the resulting errors are joined with [Join].
func JoinFuncs(fns ...ErrorFunc) error {
	var total int
	for _, fn := range fns {
		if fn != nil {
			total++
		}
	}
	if total == 0 {
		return nil
	}

	var (
		tmp  [4]error
		errs = tmp[:0]
	)
	if cap(errs) < total {
		errs = make([]error, 0, total)
	}

	for _, fn := range fns {
		if fn != nil {
			if err := fn(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return Join(errs...)
	}
}

// AppendFunc evaluates fn and appends it to err. If either err or fn are nil,
// the other is returned. If fn returns a nil error, err is returned.
func AppendFunc(err error, fn ErrorFunc) error {
	switch {
	case fn == nil:
		return err
	case err == nil:
		return fn()
	default:
		if e := fn(); e != nil {
			return errors.Join(err, e)
		}
		return err
	}
}

// AppendFuncs evaluates fns serially, appending each return value to err. Nil
// errors are ignored. If err and fns produce no non-nil errors, nil is
// returned; if only one non-nil error is produced, it is returned verbatim.
// Otherwise, the resulting non-nil errors are joined with [Join].
func AppendFuncs(err error, fns ...ErrorFunc) error {
	if len(fns) == 0 {
		return err
	}

	var (
		tmp  [4]error
		errs = tmp[:0]
	)

	total := len(fns)
	if err != nil {
		total++
	}
	if cap(errs) < total {
		errs = make([]error, 0, total)
	}

	if err != nil {
		errs = append(errs, err)
	}

	for _, fn := range fns {
		if fn != nil {
			if e := fn(); e != nil {
				errs = append(errs, e)
			}
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return Join(errs...)
	}
}

// Lazy returns an error that will lazily evaluate fn; that is, fn will be
// called at most once, and not until the resulting error would be used.
func Lazy(fn ErrorFunc) error {
	return &lazyError{
		get: sync.OnceValue(fn),
	}
}

type lazyError struct {
	get ErrorFunc
}

func (e lazyError) As(target any) bool {
	return errors.As(e.get(), target)
}

func (e lazyError) Is(target error) bool {
	return errors.Is(e.get(), target)
}

func (e lazyError) Unwrap() error {
	return errors.Unwrap(e.get())
}

func (e lazyError) Error() string {
	return e.get().Error()
}
