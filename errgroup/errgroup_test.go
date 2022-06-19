package errgroup_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/errors/errgroup"
	"go.uber.org/multierr"
)

var (
	errA = errors.New("a")
	errB = errors.New("b")
	errC = errors.New("c")
)

func TestErrGroup(t *testing.T) {
	err := errgroup.All(
		func() error {
			time.Sleep(100 * time.Millisecond)
			return errA
		},
		func() error {
			time.Sleep(50 * time.Millisecond)
			return errB
		},
		func() error {
			return errC
		},
	)

	require.Error(t, err)
	require.ErrorContains(t, err, errA.Error())
	require.ErrorContains(t, err, errB.Error())
	require.ErrorContains(t, err, errC.Error())
}

func TestErrGroupInline(t *testing.T) {
	var (
		expectErr = multierr.Combine(errA, errB, errC)
		err       = errgroup.AllInline(
			func() error {
				time.Sleep(100 * time.Millisecond)
				return errA
			},
			func() error {
				time.Sleep(50 * time.Millisecond)
				return errB
			},
			func() error {
				return errC
			},
		)
	)

	require.EqualError(t, err, expectErr.Error())
}

func TestErrGroupFirst(t *testing.T) {
	var (
		expectErr = errB
		wait      = make(chan struct{})
		err       = errgroup.First(
			func() error {
				<-wait
				time.Sleep(100 * time.Millisecond)
				return errA
			},
			func() error {
				defer close(wait)
				return errB
			},
			func() error {
				<-wait
				time.Sleep(100 * time.Millisecond)
				return errC
			},
		)
	)

	require.EqualError(t, err, expectErr.Error())
}

func TestErrGroupFirstInline(t *testing.T) {
	var (
		expectErr = errA
		err       = errgroup.FirstInline(
			func() error {
				return errA
			},
			func() error {
				return errB
			},
			func() error {
				return errC
			},
		)
	)

	require.EqualError(t, err, expectErr.Error())
}

func TestErrGroupIgnoredErrors(t *testing.T) {
	var (
		expectErr = errC
		g         = errgroup.New(
			errgroup.WithInline(),
			errgroup.WithIgnoredErrors(io.EOF),
		)
	)

	g.Add(
		func() error {
			return nil
		},
		func() error {
			return io.EOF
		},
		func() error {
			return errC
		},
	)

	require.EqualError(t, g.Wait(), expectErr.Error())
}

func TestErrGroupNoErrors(t *testing.T) {
	err := errgroup.All(
		func() error { return nil },
		func() error { return nil },
		func() error { return nil },
	)

	require.NoError(t, err)
}

func TestWithoutContext(t *testing.T) {
	var (
		err = errors.New("foo")
		fnA = func() error {
			return err
		}
		fnB = errgroup.WithoutContext(func(context.Context) error {
			return err
		})
	)

	require.ErrorIs(t, fnB(), fnA())
}
