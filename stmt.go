package driver

import (
	"context"
	"database/sql/driver"
)

type myStmt struct {
	driver.Stmt
	query string
	hook  Hook
	ctx   context.Context
}

func (my *myStmt) Exec(args []driver.Value) (driver.Result, error) {
	my.hook.Before(my.ctx, MethodExec, my.query, args)
	// nolint
	// SA1019: my.Stmt.Exec has been deprecated since Go 1.8: Drivers should implement StmtExecContext instead (or additionally). (staticcheck)
	result, err := my.Stmt.Exec(args)
	my.hook.After(my.ctx, MethodExec, my.query, args, result, err)
	return result, err
}

func (my *myStmt) Query(args []driver.Value) (driver.Rows, error) {
	my.hook.Before(my.ctx, MethodQuery, my.query, args)
	rows, err := my.Stmt.Query(args) // nolint
	my.hook.After(my.ctx, MethodQuery, my.query, args, rows, err)
	return rows, err
}

var _ driver.Stmt = (*myStmt)(nil)
var _ driver.StmtExecContext = (*myStmt)(nil)

func (my *myStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if siCtx, ok := my.Stmt.(driver.StmtExecContext); ok {
		my.hook.Before(ctx, MethodExec, my.query, args)
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

func (my *myStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if siCtx, ok := my.Stmt.(driver.StmtQueryContext); ok {
		my.hook.Before(ctx, MethodQuery, my.query, args)
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
