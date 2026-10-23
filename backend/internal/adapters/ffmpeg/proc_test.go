package ffmpeg

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests run sh in place of ffmpeg to check process handling.

func TestProcReadsStdout(t *testing.T) {
	p, err := start(context.Background(), "sh", []string{"-c", "printf abc"})
	require.NoError(t, err)

	got, err := io.ReadAll(p)

	require.NoError(t, err)
	assert.Equal(t, "abc", string(got))
}

func TestProcReportsFailure(t *testing.T) {
	p, err := start(context.Background(), "sh", []string{"-c", "printf abc; echo boom >&2; exit 3"})
	require.NoError(t, err)

	_, err = io.ReadAll(p)

	assert.ErrorContains(t, err, "boom")
}

func TestProcCloseStopsProcess(t *testing.T) {
	p, err := start(context.Background(), "sh", []string{"-c", "exec yes"})
	require.NoError(t, err)
	_, err = p.Read(make([]byte, 16))
	require.NoError(t, err)

	err = p.Close()

	assert.NoError(t, err)
}

func TestProcCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p, err := start(ctx, "sh", []string{"-c", "exec yes"})
	require.NoError(t, err)
	cancel()

	_, err = io.ReadAll(p)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestCappedDropsOverflow(t *testing.T) {
	c := &capped{max: 4}

	n, err := c.Write([]byte("abcdef"))

	require.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "abcd", c.buf.String())
}
