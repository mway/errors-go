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

package errors_test

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/errors"
)

func TestAs(t *testing.T) {
	chain := errors.Wrap(testError("testError"), "oops")

	var dstA error
	require.True(t, errors.As(chain, &dstA))
	require.Equal(t, chain.Error(), dstA.Error())

	var dstB testError
	require.True(t, errors.As(chain, &dstB))
	require.Equal(t, "testError", dstB.Error())

	var dstC *net.DNSError
	require.False(t, errors.As(chain, &dstC))
}

func TestIs(t *testing.T) {
	var (
		errs  = newChain(3)
		cases = []struct {
			giveChain  error
			giveTarget error
			want       bool
		}{
			{
				giveChain:  errs[2],
				giveTarget: nil,
				want:       false,
			},
			{
				giveChain:  errs[0],
				giveTarget: errs[0],
				want:       true,
			},
			{
				giveChain:  errs[1],
				giveTarget: errs[0],
				want:       true,
			},
			{
				giveChain:  errs[2],
				giveTarget: errs[0],
				want:       true,
			},
			{
				giveChain:  errs[1],
				giveTarget: errs[1],
				want:       true,
			},
			{
				giveChain:  errs[2],
				giveTarget: errs[1],
				want:       true,
			},
			{
				giveChain:  errs[2],
				giveTarget: errs[2],
				want:       true,
			},
			{
				giveChain:  errs[2],
				giveTarget: errors.Wrap(errs[1], "error3"),
				want:       false,
			},
		}
	)

	for _, tt := range cases {
		t.Run(tt.giveChain.Error(), func(t *testing.T) {
			require.Equal(t, tt.want, errors.Is(tt.giveChain, tt.giveTarget))
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		give string
		want string
	}{
		{
			give: "",
			want: "",
		},
		{
			give: "foo",
			want: "foo",
		},
		{
			give: "foo\x00bar",
			want: "foo\x00bar",
		},
	}

	for _, tt := range cases {
		t.Run(tt.give, func(t *testing.T) {
			actual := errors.New(tt.give)
			require.NotNil(t, actual)
			require.Equal(t, tt.want, actual.Error())
		})
	}
}

func TestNewf(t *testing.T) {
	cases := []struct {
		giveMessage string
		giveArgs    []any
		want        string
	}{
		{
			giveMessage: "",
			giveArgs:    nil,
			want:        "",
		},
		{
			giveMessage: "hello",
			giveArgs:    nil,
			want:        "hello",
		},
		{
			giveMessage: "hello %s",
			giveArgs:    []any{"world"},
			want:        "hello world",
		},
		{
			giveMessage: "expired after %v",
			giveArgs:    []any{time.Second},
			want:        "expired after 1s",
		},
		{
			giveMessage: "upstream said %q",
			giveArgs:    []any{"something"},
			want:        "upstream said \"something\"",
		},
	}

	for _, tt := range cases {
		t.Run(tt.giveMessage, func(t *testing.T) {
			actual := errors.Newf(tt.giveMessage, tt.giveArgs...)
			require.NotNil(t, actual)
			require.Equal(t, tt.want, actual.Error())
		})
	}
}

func TestJoin(t *testing.T) {
	var (
		errA = errors.New("foo")
		errB = errors.New("bar")
		err  = errors.Join(errA, errB)
	)

	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}

func TestUnwrap(t *testing.T) {
	var (
		errs = newChain(128)
		err  = errs[len(errs)-1]
	)

	for i := 0; i < len(errs)-1; i++ {
		tmp := errors.Unwrap(err)
		require.NotNil(t, tmp)
		require.True(t, errors.Is(err, tmp))
		err = tmp
	}

	require.NotNil(t, err)
	require.Nil(t, errors.Unwrap(err))
}

func TestWrap(t *testing.T) {
	cases := []struct {
		giveErr     error
		giveMessage string
		wantNil     bool
		wantMessage string
	}{
		{
			giveErr:     errors.New("world"),
			giveMessage: "hello",
			wantNil:     false,
			wantMessage: "hello: world",
		},
		{
			giveErr:     nil,
			giveMessage: "hello",
			wantNil:     true,
			wantMessage: "",
		},
		{
			giveErr:     errors.New("hello world"),
			giveMessage: "",
			wantNil:     false,
			wantMessage: "hello world",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := errors.Wrap(tt.giveErr, tt.giveMessage)
			if tt.wantNil {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.Equal(t, tt.wantMessage, err.Error())
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	cases := []struct {
		giveErr     error
		giveMessage string
		giveArgs    []any
		wantNil     bool
		wantMessage string
	}{
		{
			giveErr:     errors.New("world"),
			giveMessage: "hello %s %s",
			giveArgs:    []any{"to", "the"},
			wantNil:     false,
			wantMessage: "hello to the: world",
		},
		{
			giveErr:     nil,
			giveMessage: "hello %s",
			giveArgs:    []any{"world"},
			wantNil:     true,
			wantMessage: "",
		},
		{
			giveErr:     errors.New("world"),
			giveMessage: "hello",
			giveArgs:    nil,
			wantNil:     false,
			wantMessage: "hello: world",
		},
		{
			giveErr:     errors.New("hello"),
			giveMessage: "",
			giveArgs:    nil,
			wantNil:     false,
			wantMessage: "hello",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := errors.Wrapf(tt.giveErr, tt.giveMessage, tt.giveArgs...)
			if tt.wantNil {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.Equal(t, tt.wantMessage, err.Error())
			}
		})
	}
}

func TestJoinFuncs(t *testing.T) {
	var (
		errA    = errors.New("a")
		errB    = errors.New("b")
		errC    = errors.New("c")
		errFunc = func(err error) func() error {
			return func() error { return err }
		}
	)

	cases := map[string]struct {
		give []errors.ErrorFunc
		want []error
	}{
		"nominal": {
			give: []errors.ErrorFunc{
				errFunc(errA),
				errFunc(errB),
				errFunc(errC),
			},
			want: []error{
				errA,
				errB,
				errC,
			},
		},
		"no errors": {
			give: []errors.ErrorFunc{
				errFunc(nil),
				errFunc(nil),
				errFunc(nil),
			},
			want: nil,
		},
		"single error": {
			give: []errors.ErrorFunc{
				errFunc(errA),
				errFunc(nil),
				errFunc(nil),
			},
			want: []error{
				errA,
			},
		},
		"nils": {
			give: []errors.ErrorFunc{
				nil,
				nil,
				nil,
			},
			want: nil,
		},
		"interspersed": {
			give: []errors.ErrorFunc{
				nil,
				errFunc(errA),
				nil,
				errFunc(errB),
				nil,
				errFunc(nil),
				errFunc(errC),
				errFunc(nil),
			},
			want: []error{
				errA,
				errB,
				errC,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			haveErr := errors.JoinFuncs(tt.give...)
			for _, wantErr := range tt.want {
				require.ErrorIs(t, haveErr, wantErr)
			}

			if len(tt.want) == 0 {
				require.NoError(t, haveErr)
			}
		})
	}
}

func TestAppendFunc(t *testing.T) {
	cases := map[string]struct {
		lower     error
		upper     error
		wantLower bool
		wantUpper bool
	}{
		"non-nil lower and upper": {
			lower:     errors.New("a"),
			upper:     errors.New("b"),
			wantLower: true,
			wantUpper: true,
		},
		"nil lower": {
			lower:     nil,
			upper:     errors.New("b"),
			wantLower: false,
			wantUpper: true,
		},
		"nil upper": {
			lower:     errors.New("a"),
			upper:     nil,
			wantLower: true,
			wantUpper: false,
		},
		"nil lower and upper": {
			lower:     nil,
			upper:     nil,
			wantLower: true,
			wantUpper: true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			haveErr := errors.AppendFunc(tt.lower, func() error {
				return tt.upper
			})
			require.Equal(t, tt.wantLower, errors.Is(haveErr, tt.lower))
			require.Equal(t, tt.wantUpper, errors.Is(haveErr, tt.upper))
		})
	}

	t.Run("nil func", func(t *testing.T) {
		var (
			wantErr = errors.New("lower")
			haveErr = errors.AppendFunc(wantErr, nil)
		)

		require.ErrorIs(t, haveErr, wantErr)
		require.Equal(t, fmt.Sprintf("%p", wantErr), fmt.Sprintf("%p", haveErr))
	})
}

func TestAppendFuncs(t *testing.T) {
	cases := map[string]struct {
		lower     error
		upper     []error
		wantLower bool
		wantUpper []bool
	}{
		"non-nil lower and uppers": {
			lower: errors.New("a"),
			upper: []error{
				errors.New("b"),
				errors.New("c"),
				errors.New("d"),
			},
			wantLower: true,
			wantUpper: []bool{
				true, // b
				true, // c
				true, // d
			},
		},
		"non-nil lower and no uppers": {
			lower:     errors.New("a"),
			upper:     nil,
			wantLower: true,
			wantUpper: nil,
		},
		"non-nil lower and nil uppers": {
			lower: errors.New("a"),
			upper: []error{
				nil, // b
				nil, // c
				nil, // d
			},
			wantLower: true,
			wantUpper: []bool{
				false, // b
				false, // c
				false, // d
			},
		},
		"non-nil lower and mixed uppers": {
			lower: errors.New("a"),
			upper: []error{
				errors.New("b"),
				nil, // c
				errors.New("d"),
			},
			wantLower: true,
			wantUpper: []bool{
				true,  // b
				false, // c
				true,  // d
			},
		},
		"nil lower and no uppers": {
			lower:     nil,
			upper:     nil,
			wantLower: true,
			wantUpper: nil,
		},
		"nil lower and nil uppers": {
			lower: nil, // a
			upper: []error{
				nil, // b
				nil, // c
				nil, // d
			},
			wantLower: true, // a
			wantUpper: []bool{
				true, // b
				true, // c
				true, // d
			},
		},
		"nil lower and non-nil uppers": {
			lower: nil,
			upper: []error{
				errors.New("b"),
				errors.New("c"),
				errors.New("d"),
				errors.New("e"),
				errors.New("f"),
			},
			wantLower: false,
			wantUpper: []bool{
				true, // b
				true, // c
				true, // d
				true, // e
				true, // f
			},
		},
	}

	for name, tt := range cases {
		require.Equal(t, len(tt.upper), len(tt.wantUpper)) // meta

		t.Run(name, func(t *testing.T) {
			var fns []errors.ErrorFunc
			for _, upper := range tt.upper {
				upper := upper
				fns = append(fns, func() error { return upper })
			}

			haveErr := errors.AppendFuncs(tt.lower, fns...)
			require.Equal(t, tt.wantLower, errors.Is(haveErr, tt.lower))

			for i, want := range tt.wantUpper {
				require.Equal(t, want, errors.Is(haveErr, tt.upper[i]))
			}
		})
	}
}

func TestLazy_Is(t *testing.T) {
	var (
		errA    = errors.New("explicit error")
		errB    = errors.New("lazy error")
		errFunc = errors.Lazy(func() error {
			return errB
		})
		haveErr = errors.Join(errA, errFunc)
	)

	require.ErrorIs(t, haveErr, errA)
	require.ErrorIs(t, haveErr, errB)
}

func TestLazy_As(t *testing.T) {
	err := errors.Lazy(func() error {
		return testError(t.Name())
	})

	var dst testError
	require.True(t, errors.As(err, &dst))
	require.Equal(t, t.Name(), string(dst))
}

func TestLazy_Unwrap(t *testing.T) {
	err := errors.Lazy(func() error {
		return fmt.Errorf("wrapped: %w", testError(t.Name()))
	})

	unwrapped := errors.Unwrap(err)
	require.Error(t, unwrapped)
	require.ErrorContains(t, unwrapped, t.Name())
	require.NotContains(t, unwrapped.Error(), "wrapped")

	var dst testError
	require.True(t, errors.As(err, &dst))
	require.Equal(t, t.Name(), string(dst))
}

func TestLazy_Error(t *testing.T) {
	err := errors.Lazy(func() error {
		return testError(t.Name())
	})
	require.Equal(t, t.Name(), err.Error())
}

func newChain(size int) []error {
	var (
		errs []error
		err  error
	)

	for i := 1; i <= size; i++ {
		if err == nil {
			err = errors.Newf("error%d", i)
		} else {
			err = errors.Wrapf(err, "error%d", i)
		}

		errs = append(errs, err)
	}

	return errs
}

type testError string

func (e testError) Error() string {
	return string(e)
}

func (e testError) IsTest() bool {
	return true
}
