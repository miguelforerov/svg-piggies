package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
)

func TestRunInBobTransactionCommitsSuccessfulCallback(t *testing.T) {
	t.Parallel()

	state := &transactionState{}
	transactor := fakeTransactor{transaction: fakeTransaction{state: state}}
	called := false
	err := RunInBobTransaction(
		context.Background(),
		transactor,
		func(_ context.Context, transaction bob.Transaction) error {
			called = true
			if transaction == nil {
				t.Fatal("transaction is nil")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("RunInBobTransaction() error = %v", err)
	}
	if !called || !state.committed {
		t.Fatalf("called = %v, committed = %v", called, state.committed)
	}
}

func TestRunInBobTransactionRollsBackFailedCallback(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("operation failed")
	state := &transactionState{}
	transactor := fakeTransactor{transaction: fakeTransaction{state: state}}
	err := RunInBobTransaction(
		context.Background(),
		transactor,
		func(context.Context, bob.Transaction) error { return wantErr },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunInBobTransaction() error = %v, want %v", err, wantErr)
	}
	if state.committed || !state.rolledBack {
		t.Fatalf("committed = %v, rolledBack = %v", state.committed, state.rolledBack)
	}
}

type transactionState struct {
	committed  bool
	rolledBack bool
}

type fakeTransaction struct {
	state *transactionState
}

func (t fakeTransaction) Commit(context.Context) error {
	t.state.committed = true
	return nil
}

func (t fakeTransaction) Rollback(context.Context) error {
	if !t.state.committed {
		t.state.rolledBack = true
	}
	return nil
}

func (fakeTransaction) ExecContext(
	context.Context,
	string,
	...any,
) (sql.Result, error) {
	return nil, nil
}

func (fakeTransaction) QueryContext(
	context.Context,
	string,
	...any,
) (scan.Rows, error) {
	return nil, nil
}

type fakeTransactor struct {
	transaction fakeTransaction
}

func (t fakeTransactor) Begin(context.Context) (fakeTransaction, error) {
	return t.transaction, nil
}

func (fakeTransactor) ExecContext(
	context.Context,
	string,
	...any,
) (sql.Result, error) {
	return nil, nil
}

func (fakeTransactor) QueryContext(
	context.Context,
	string,
	...any,
) (scan.Rows, error) {
	return nil, nil
}
