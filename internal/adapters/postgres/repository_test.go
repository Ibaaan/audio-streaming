//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
)

func TestWithinTxCommits(t *testing.T) {
	repo := newTestRepo(t)
	var track catalog.Track

	err := repo.WithinTx(context.Background(),
		func(tx *Repository) error {
			track = mustAddTrack(t, tx)
			return nil
		})

	require.NoError(t, err)
	_, err = repo.GetTrack(context.Background(), track.ID)
	assert.NoError(t, err, "track visible after commit")
}

func TestWithinTxRollsBackOnError(t *testing.T) {
	repo := newTestRepo(t)
	stepErr := errors.New("boom")
	var track catalog.Track

	err := repo.WithinTx(context.Background(),
		func(tx *Repository) error {
			track = mustAddTrack(t, tx)
			return stepErr
		})

	require.ErrorIs(t, err, stepErr)
	_, err = repo.GetTrack(context.Background(), track.ID)
	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
