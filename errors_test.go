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

package errors_test

import (
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
