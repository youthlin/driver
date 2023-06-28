# driver

```bash
go get -u github.com/youthlin/driver
```

```go
import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/youthlin/driver"
	sqlite3 "github.com/youthlin/go-sqlcipher"
)

// ...

sql.Register("test", driver.Wrap(&sqlite3.SQLiteDriver{}, driver.NewHook(
	/*before*/ func(ctx context.Context, method driver.Method, query string, args any) {},
	/*after*/  func(ctx context.Context, method driver.Method, query string, args, result any, err error) {
		log.Printf("cost=%v, method=%v, query=%v, args=%+v, resutl=%v, err=%+v\n",
			driver.Cost(ctx), method, query, args, result, err)
	},
)))

db, err := sql.Open("test", ":memory:")

// ...

```
