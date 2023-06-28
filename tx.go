package driver

import (
	"context"
	"database/sql/driver"
)

type myTx struct {
	tx   driver.Tx
	ctx  context.Context
	hook Hook
}

var _ driver.Tx = (*myTx)(nil)

func (my *myTx) Commit() error {
	my.hook.Before(my.ctx, MethodCommit, string(MethodCommit), nil)
	err := my.tx.Commit()
	my.hook.After(my.ctx, MethodCommit, string(MethodCommit), nil, nil, err)
	return err
}

func (my *myTx) Rollback() error {
	my.hook.Before(my.ctx, MethodRollback, string(MethodRollback), nil)
	err := my.tx.Rollback()
	my.hook.After(my.ctx, MethodRollback, string(MethodRollback), nil, nil, err)
	return err
}
