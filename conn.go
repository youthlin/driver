package driver

import (
	"context"
	"database/sql/driver"
	"errors"
)

var _ driver.Conn = (*myConn)(nil)

type myConn struct {
	// 嵌入接口 使得我们不需要实现 driver.Conn 的所有方法
	// 而只需要重写我们需要关心的 Prepare 方法
	driver.Conn
	hook Hook
}

func (my *myConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := my.Conn.Prepare(query)
	return &myStmt{Stmt: stmt, ctx: context.Background()}, err
}

// db.ExecContext:
// 如果 Conn 实现了 driver.ExecerContext 则通过 ExecContext 执行
// 否则如果 Conn 实现了 driver.Execer 则通过 Exec 执行
// 否则通过 PrepareContext / Prepare 预编译再执行
var (
	_ driver.ExecerContext = (*myConn)(nil)
	// nolint
	// SA1019: driver.Execer has been deprecated since Go 1.8: Drivers should implement ExecerContext instead. (staticcheck)
	_ driver.Execer = (*myConn)(nil)
	// 带 ctx 的 Prepare
	_ driver.ConnPrepareContext = (*myConn)(nil)
)

// ExecContext implements the driver.ExecerContext interface.
func (my *myConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if execerCtx, ok := my.Conn.(driver.ExecerContext); ok {
		my.hook.Before(ctx, MethodExec, query, args)
		result, err := execerCtx.ExecContext(ctx, query, args)
		my.hook.After(ctx, MethodExec, query, args, result, err)
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
	return my.Exec(query, dargs)
}

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

// Exec implements the driver.Execer interface.
func (my *myConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := my.Conn.(driver.Execer); ok { // nolint
		ctx := context.Background()
		my.hook.Before(ctx, MethodExec, query, args)
		result, err := execer.Exec(query, args)
		my.hook.After(ctx, MethodExec, query, args, result, err)
		return result, err
	}
	return nil, driver.ErrSkip
}

// PrepareContext implements the driver.ConnPrepareContext interface.
func (my *myConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if ciCtx, ok := my.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := ciCtx.PrepareContext(ctx, query)
		return &myStmt{Stmt: stmt, query: query, hook: my.hook, ctx: ctx}, err
	}
	stmt, err := my.Conn.Prepare(query)
	return &myStmt{Stmt: stmt, query: query, hook: my.hook, ctx: ctx}, err
}

// db.QueryContext:
var (
	_ driver.QueryerContext = (*myConn)(nil)
	// nolint
	_ driver.Queryer = (*myConn)(nil)
)

func (my *myConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if queryerCtx, ok := my.Conn.(driver.QueryerContext); ok {
		my.hook.Before(ctx, MethodQuery, query, args)
		rows, err := queryerCtx.QueryContext(ctx, query, args)
		my.hook.After(ctx, MethodQuery, query, args, rows, err)
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
	return my.Query(query, dargs)
}

func (my *myConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := my.Conn.(driver.Queryer); ok { // nolint
		ctx := context.Background()
		my.hook.Before(ctx, MethodQuery, query, args)
		rows, err := queryer.Query(query, args)
		my.hook.After(ctx, MethodQuery, query, args, rows, err)
		return rows, err
	}
	return nil, driver.ErrSkip
}

var _ driver.ConnBeginTx = (*myConn)(nil)

func (my *myConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if ciCtx, ok := my.Conn.(driver.ConnBeginTx); ok {
		my.hook.Before(ctx, MethodBegin, string(MethodBegin), opts)
		tx, err := ciCtx.BeginTx(ctx, opts)
		my.hook.After(ctx, MethodBegin, string(MethodBegin), opts, tx, err)
		return &myTx{tx: tx, hook: my.hook, ctx: ctx}, err
	}

	if ctx.Done() == nil {
		tx, err := my.Conn.Begin() // nolint
		return &myTx{tx: tx, hook: my.hook, ctx: ctx}, err
	}

	tx, err := my.Conn.Begin() // nolint
	tx = &myTx{tx: tx, hook: my.hook, ctx: ctx}
	if err == nil {
		select {
		default:
		case <-ctx.Done():
			_ = tx.Rollback()
			return nil, ctx.Err()
		}
	}
	return tx, err
}