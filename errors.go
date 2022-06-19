package errors

import (
	"errors"
	"fmt"
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
// Note that msg must not contain either the formatting verb %w nor its escaped
// version, %%w.
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
		return fmt.Errorf("%s: %w", fmt.Sprintf(msg, args...), base)
	}
}
