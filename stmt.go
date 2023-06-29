package driver

import (
	"context"
	"database/sql/driver"
)

// myStmt 使用 hook 包装驱动返回的 Stmt.
type myStmt struct {
	driver.Stmt
	query string
	hook  Hook
	ctx   context.Context
}

// Exec implements the driver.Stmt interface.
// 在执行前后调用 hook 方法。
func (my *myStmt) Exec(args []driver.Value) (driver.Result, error) {
	my.ctx = my.hook.Before(my.ctx, MethodExec, my.query, args)
	// nolint
	// SA1019: my.Stmt.Exec has been deprecated since Go 1.8: Drivers should implement StmtExecContext instead (or additionally). (staticcheck)
	result, err := my.Stmt.Exec(args)
	my.hook.After(my.ctx, MethodExec, my.query, args, result, err)
	return result, err
}

// Query implements the driver.Stmt interface.
// 在执行前后调用 hook 方法。
func (my *myStmt) Query(args []driver.Value) (driver.Rows, error) {
	my.ctx = my.hook.Before(my.ctx, MethodQuery, my.query, args)
	rows, err := my.Stmt.Query(args) // nolint
	my.hook.After(my.ctx, MethodQuery, my.query, args, rows, err)
	return rows, err
}

var _ driver.Stmt = (*myStmt)(nil)
var _ driver.StmtExecContext = (*myStmt)(nil)

// ExecContext implements the driver.StmtQueryContext interface.
// 如果驱动实现了 driver.StmtQueryContext 则直接通过 ExecContext 执行，并在执行前后调用 hook 方法。
// 否则走到 Exec 方法。
func (my *myStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if siCtx, ok := my.Stmt.(driver.StmtExecContext); ok {
		ctx = my.hook.Before(ctx, MethodExec, my.query, args)
		result, err := siCtx.ExecContext(ctx, args)
		my.hook.After(ctx, MethodExec, my.query, args, result, err)
		return result, err
	}

	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	my.ctx = ctx
	return my.Exec(dargs)
}

var _ driver.StmtQueryContext = (*myStmt)(nil)

// QueryContext implements the driver.StmtQueryContext interface.
// 如果驱动实现了 driver.StmtQueryContext 则直接通过 QueryContext 查询，并在查询前后调用 hook 方法。
// 否则走到 Query 方法。
func (my *myStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if siCtx, ok := my.Stmt.(driver.StmtQueryContext); ok {
		ctx = my.hook.Before(ctx, MethodQuery, my.query, args)
		rows, err := siCtx.QueryContext(ctx, args)
		my.hook.After(ctx, MethodQuery, my.query, args, rows, err)
		return rows, err
	}

	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	my.ctx = ctx
	return my.Query(dargs)
}
