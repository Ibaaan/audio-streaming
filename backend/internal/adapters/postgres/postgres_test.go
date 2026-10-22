package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTx records Commit/Rollback calls. Embedding pgx.Tx satisfies the rest
// of the interface; calling any other method panics on the nil embed.
type fakeTx struct {
	pgx.Tx
	commitErr  error
	committed  bool
	rolledBack bool
	rbCtxErr   error
}

func (f *fakeTx) Commit(ctx context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeTx) Rollback(ctx context.Context) error {
	f.rolledBack = true
	f.rbCtxErr = ctx.Err()
	return nil
}

type fakeBeginner struct {
	tx  *fakeTx
	err error
}

func (b *fakeBeginner) Begin(ctx context.Context) (pgx.Tx, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.tx, nil
}

func TestWithTransactionCommitsWhenAllStepsSucceed(t *testing.T) {
	tx := &fakeTx{}
	var calls []int
	number := 0
	ok := func(DBTX) error {
		calls = append(calls, number)
		number += 1
		return nil
	}

	err := WithTransaction(
		context.Background(),
		&fakeBeginner{tx: tx},
		func(dbtx DBTX) error {
			if err := ok(dbtx); err != nil {
				return err
			}

			if err := ok(dbtx); err != nil {
				return err
			}

			if err := ok(dbtx); err != nil {
				return err
			}
			return nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2}, calls)
	assert.True(t, tx.committed, "committed")
}

func TestWithTransactionStopsAndRollsBackOnStepError(t *testing.T) {
	tx := &fakeTx{}
	stepErr := errors.New("boom")
	var calls []int
	number := 0
	ok := func(DBTX) error {
		calls = append(calls, number)
		number += 1
		return nil
	}
	fail := func(DBTX) error {
		calls = append(calls, number)
		number += 1
		return stepErr
	}

	test := func(dbtx DBTX) error {
		if err := ok(dbtx); err != nil {
			return err
		}

		if err := fail(dbtx); err != nil {
			return err
		}

		if err := ok(dbtx); err != nil {
			return err
		}
		return nil
	}

	err := WithTransaction(context.Background(), &fakeBeginner{tx: tx},
		test)

	require.ErrorIs(t, err, stepErr)
	assert.Equal(t, []int{0, 1}, calls, "steps after failure skipped")
	assert.False(t, tx.committed, "committed")
	assert.True(t, tx.rolledBack, "rolled back")
}

func TestWithTransactionReturnsBeginError(t *testing.T) {
	beginErr := errors.New("no conn")
	called := false
	step := func(DBTX) error { called = true; return nil }

	err := WithTransaction(context.Background(),
		&fakeBeginner{err: beginErr}, step)

	require.ErrorIs(t, err, beginErr)
	assert.ErrorContains(t, err, "begin tx")
	assert.False(t, called, "step called")
}

func TestWithTransactionReturnsCommitError(t *testing.T) {
	commitErr := errors.New("serialization failure")
	tx := &fakeTx{commitErr: commitErr}
	called := false
	step := func(DBTX) error { called = true; return nil }

	err := WithTransaction(context.Background(), &fakeBeginner{tx: tx},
		step)

	require.ErrorIs(t, err, commitErr)
	assert.ErrorContains(t, err, "commit tx")
	assert.True(t, called, "step called")
	assert.True(t, tx.committed, "commit attempted")
	assert.True(t, tx.rolledBack)
}

func TestWithTransactionRollsBackWithLiveCtxAfterCancel(t *testing.T) {
	tx := &fakeTx{}
	ctx, cancel := context.WithCancel(context.Background())
	cancelThenFail := func(DBTX) error {
		cancel()
		return ctx.Err()
	}

	err := WithTransaction(ctx, &fakeBeginner{tx: tx}, cancelThenFail)

	require.ErrorIs(t, err, context.Canceled)
	assert.True(t, tx.rolledBack, "rolled back")
	assert.NoError(t, tx.rbCtxErr, "rollback ctx should not be cancelled")
}
