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

// Package errgroup provides error synchronization and combination tools.
package errgroup

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/multierr"
)

type (
	// An ErrFunc is a function that returns an error.
	ErrFunc = func() error
	// A ContextErrFunc is a function that accepts a context.Context and
	// returns an error.
	ContextErrFunc = func(context.Context) error
)

// Group is functionally similar to the standard library's x/sync/errgroup
// package, but offers more options to customize its behavior based on a given
// workflow. See the Options documentation for more information.
//
// Groups cannot be reused. A zero-value Group is valid and ready to use.
type Group struct {
	err     error
	mu      sync.Mutex
	wg      sync.WaitGroup
	options Options
}

// New creates a new Group with the given options.
func New(opts ...Option) *Group {
	options := DefaultOptions().With(opts...)
	return &Group{
		options: options,
	}
}

// Add executes the provided functions and stores returned errors for retrieval
// with Wait(). If the Group was configured using the WithInline() option, the
// given functions are executed immediately and serially in the calling
// goroutine; otherwise, the given functions are executed in parallel.
func (g *Group) Add(fns ...ErrFunc) {
	if g.options.Inline {
		for _, f := range fns {
			g.appendError(f())
		}
		return
	}

	for _, f := range fns {
		f := f
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			g.appendError(f())
		}()
	}
}

// Wait blocks until all functions passed to Add have been executed and
// returns an error if any were encountered.
//
// The error return depends upon whether the Group was configured using the
// WithFirstOnly() option. If WithFirstOnly was not used, the returned error
// is an error containing a chain of all non-nil errors returned by the
// executed functions; if WithFirstOnly was used, the returned error is the
// first non-nil error returned verbatim by the first function to finish
// executing.
func (g *Group) Wait() error {
	g.wg.Wait()

	g.mu.Lock()
	defer g.mu.Unlock()

	return g.err
}

func (g *Group) appendError(err error) {
	if err == nil {
		return
	}

	for _, ignored := range g.options.IgnoredErrors {
		if errors.Is(err, ignored) {
			return
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.err != nil && g.options.FirstOnly {
		return
	}

	g.err = multierr.Append(g.err, err)
}

// WithoutContext wraps a ContextErrFunc in an ErrFunc, providing a background
// context to the given ContextErrFunc.
func WithoutContext(fn ContextErrFunc) ErrFunc {
	return func() error {
		return fn(context.Background())
	}
}

// All executes all of the given functions in parallel, and collects and
// combines all of their returned errors.
func All(fns ...ErrFunc) error {
	return do(fns)
}

// AllInline executes all of the given functions serially in the calling goroutine,
// and collects and combines all of their returned errors.
func AllInline(fns ...ErrFunc) error {
	return do(fns, WithInline())
}

// First executes all of the given functions in parallel, and returns the first
// error returned by them.
func First(fns ...ErrFunc) error {
	return do(fns, WithFirstOnly())
}

// FirstInline executes all of the given functions serially in the calling
// goroutine, and returns the first error returned by them.
func FirstInline(fns ...ErrFunc) error {
	return do(fns, WithFirstOnly(), WithInline())
}

func do(fns []ErrFunc, opts ...Option) error {
	g := Group{
		options: DefaultOptions(),
	}

	for _, opt := range opts {
		opt.apply(&g.options)
	}

	g.Add(fns...)
	return g.Wait()
}
