package async

import (
	"context"
	"errors"
	"sync"
	"time"
)

var _ context.Context = (*joinedCtx)(nil)

type joinedCtx struct {
	fallback, primary context.Context

	done func() <-chan struct{}
}

func joinCtx(fallback, primary context.Context) context.Context {
	return joinedCtx{
		fallback: fallback,
		primary:  primary,
		done: sync.OnceValue(func() <-chan struct{} {
			primaryCh := primary.Done()
			fallbackCh := fallback.Done()

			if primaryCh == nil && fallbackCh == nil {
				return nil
			}

			ch := make(chan struct{})

			go func() {
				defer close(ch)

				select {
				case <-primaryCh:
				case <-fallbackCh:
				}
			}()

			return ch
		}),
	}
}

// Deadline implements [context.Context].
func (j joinedCtx) Deadline() (deadline time.Time, ok bool) {
	deadline, ok = j.primary.Deadline()
	if !ok {
		return j.fallback.Deadline()
	}

	fallback, ok := j.fallback.Deadline()
	if !ok {
		return deadline, true
	}

	if deadline.Before(fallback) {
		return deadline, true
	}

	return fallback, true
}

// Done implements [context.Context].
func (j joinedCtx) Done() <-chan struct{} {
	return j.done()
}

// Err implements [context.Context].
func (j joinedCtx) Err() error {
	return errors.Join(j.primary.Err(), j.fallback.Err())
}

// Value implements [context.Context].
func (j joinedCtx) Value(key any) any {
	if v := j.primary.Value(key); v != nil {
		return v
	}

	return j.fallback.Value(key)
}
