package driver

import (
	"context"
	"time"
)

type Method string

const (
	MethodExec     Method = "Exec"
	MethodQuery    Method = "Query"
	MethodBegin    Method = "Begin"
	MethodCommit   Method = "Commit"
	MethodRollback Method = "Rollback"
)

// Hook
type Hook interface {
	Before(ctx context.Context, method Method, query string, args any)
	After(ctx context.Context, method Method, query string, args any, result any, err error)
}

func NewHook(before BeforeFn, after AfterFn) Hook {
	return h{BeforeFn: before, AfterFn: after}
}

type BeforeFn func(ctx context.Context, method Method, query string, args any)
type AfterFn func(ctx context.Context, method Method, query string, args any, result any, err error)

type h struct {
	BeforeFn
	AfterFn
}

func (h h) Before(ctx context.Context, method Method, query string, args any) {
	if h.BeforeFn != nil {
		h.BeforeFn(ctx, method, query, args)
	}
}

func (h h) After(ctx context.Context, method Method, query string, args any, result any, err error) {
	if h.AfterFn != nil {
		h.AfterFn(ctx, method, query, args, result, err)
	}
}

func safeHook(hook Hook) Hook {
	return &myHook{hook: hook}
}

type myHook struct {
	hook Hook
}

func safeFn(fn func()) {
	defer func() {
		_ = recover()
	}()
	fn()
}

var startAt int64

func Cost(ctx context.Context) time.Duration {
	v := ctx.Value(&startAt)
	if start, ok := v.(int64); ok {
		return time.Since(time.Unix(0, start))
	}
	return 0
}

func (my *myHook) Before(ctx context.Context, method Method, query string, args any) {
	safeFn(func() {
		my.hook.Before(context.WithValue(ctx, &startAt, time.Now().UnixNano()), method, query, args)
	})
}

func (my *myHook) After(ctx context.Context, method Method, query string, args any, result any, err error) {
	safeFn(func() {
		my.hook.After(ctx, method, query, args, result, err)
	})
}
