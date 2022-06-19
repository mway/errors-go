package errgroup_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mway.dev/errors/errgroup"
)

func TestOptionsWith(t *testing.T) {
	var (
		base     = errgroup.DefaultOptions()
		previous = base.With(
			errgroup.WithFirstOnly(),
			errgroup.WithIgnoredErrors(context.Canceled),
		)
		updated = previous.With(
			errgroup.DefaultOptions().With(
				errgroup.WithInline(),
				errgroup.WithIgnoredErrors(io.EOF),
			),
		)
	)

	require.True(t, previous.FirstOnly)
	require.False(t, previous.Inline)
	require.Len(t, previous.IgnoredErrors, 1)

	require.False(t, updated.FirstOnly)
	require.True(t, updated.Inline)
	require.Len(t, updated.IgnoredErrors, 2)
}
