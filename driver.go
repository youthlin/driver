// https://github.com/golang/go/wiki/SQLInterface
// 打开数据库连接
//
//	db, err := sql.Open(driver, dataSourceName)
//
// 执行 SQL 语句
//
//	result, err :=db.ExecContext(ctx, "INSERT INTO users (name, age) VALUES ($1, $2)", "gopher", 27)
//
// 查询内容
//
//	// 查询多行
//	rows, err := db.QueryContext(ctx, "SELECT name FROM users WHERE age = $1", age)
//	// 查询单条记录
//	err := db.QueryRowContext(ctx, "SELECT age FROM users WHERE name = $1", name).Scan(&age)
//	// 预编译 SQL 语句
//	stmt, err := db.PrepareContext(ctx, "SELECT name FROM users WHERE age = $1")
//	rows, err := stmt.Query(age)
// 事务
// 	tx, err :=db.BeginTx(ctx, nil)
// 	tx.Commit()
// 	tx.Rollback()

package driver

import (
	"database/sql"
	"database/sql/driver"
)

// Register register a driver with the given driverName and add a hook to it.
// 使用给定的驱动名称注册一个数据库驱动，并在其上添加 hook 方法。
func Register(driverName string, driver driver.Driver, hook Hook) {
	sql.Register(driverName, Wrap(driver, hook))
}

// Wrap return a wrapped Driver.
// 返回包装后的 Driver.
func Wrap(d driver.Driver, hook Hook) driver.Driver {
	return &myDriver{driver: d, hook: safeHook(hook)}
}

var _ driver.Driver = (*myDriver)(nil)

// myDriver implements driver.Driver.
// 自定义驱动结构体。
type myDriver struct {
	driver driver.Driver
	hook   Hook
}

// Open implements the driver.Driver interface.
// 必须实现 driver.Driver 的 Open 方法。
func (my *myDriver) Open(name string) (driver.Conn, error) {
	conn, err := my.driver.Open(name)
	return &myConn{Conn: conn, hook: my.hook}, err
}
