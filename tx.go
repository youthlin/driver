package driver

import (
	"context"
	"database/sql/driver"
)

// myTx 使用 hook 包装驱动返回的 Tx.
type myTx struct {
	tx   driver.Tx
	ctx  context.Context
	hook Hook
}

var _ driver.Tx = (*myTx)(nil)

// Commit implements the driver.Tx interface.
// 在提交前后调用 hook 方法。
func (my *myTx) Commit() error {
	my.ctx = my.hook.Before(my.ctx, MethodCommit, string(MethodCommit), nil)
	err := my.tx.Commit()
	my.hook.After(my.ctx, MethodCommit, string(MethodCommit), nil, nil, err)
	return err
}

// Rollback implements the driver.Tx interface.
// 在回滚前后调用 hook 方法。
func (my *myTx) Rollback() error {
	my.ctx = my.hook.Before(my.ctx, MethodRollback, string(MethodRollback), nil)
	err := my.tx.Rollback()
	my.hook.After(my.ctx, MethodRollback, string(MethodRollback), nil, nil, err)
	return err
}
