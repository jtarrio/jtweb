package genericsql

import (
	"context"
	"database/sql"
	"fmt"
)

type sqlErrorWrapper struct {
	parent error
}

func sqlError(e error) error {
	if e == nil {
		return nil
	}
	return &sqlErrorWrapper{parent: e}
}

func (e *sqlErrorWrapper) Error() string {
	return e.parent.Error()
}

func IsSqlError(e error) bool {
	_, ok := e.(*sqlErrorWrapper)
	return ok
}

func doInReadTx[R any](e *GenericSqlEngine, op func(tx *sql.Tx) (R, error)) (R, error) {
	return doInTx(e, sql.LevelReadCommitted, op)
}

func doInWriteTx[R any](e *GenericSqlEngine, op func(tx *sql.Tx) (R, error)) (R, error) {
	retries_left := 1
	for {
		ret, err := doInTx(e, sql.LevelSerializable, op)
		if !IsSqlError(err) {
			return ret, err
		}
		if retries_left <= 0 {
			var zero R
			return zero, fmt.Errorf("maximum retries exceeded: %s", err)
		}
		retries_left--
	}
}

func doInWriteTxNoReturn(e *GenericSqlEngine, op func(tx *sql.Tx) error) error {
	_, err := doInWriteTx(e, func(tx *sql.Tx) (bool, error) {
		return false, op(tx)
	})
	return err
}

func doInTx[R any](e *GenericSqlEngine, level sql.IsolationLevel, op func(tx *sql.Tx) (R, error)) (R, error) {
	var zero R
	tx, err := e.db.BeginTx(context.TODO(), &sql.TxOptions{Isolation: (sql.LevelReadCommitted)})
	if err != nil {
		return zero, err
	}
	ret, err := op(tx)
	if err != nil {
		return zero, sqlError(tx.Rollback())
	}
	return ret, sqlError(tx.Commit())
}
