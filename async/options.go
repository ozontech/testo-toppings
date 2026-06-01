package async

import (
	"context"

	"github.com/ozontech/testo/testoplugin"
)

type option func(p *PluginAsync)

// WithLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
// A limit of zero will prevent any new goroutines from being added.
//
// Any subsequent call to the [Run] method will block until it can add an active
// goroutine without exceeding the configured limit.
func WithLimit(n int) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginAsync) {
			if n < 0 {
				p.sem = nil

				return
			}

			p.sem = make(chan struct{}, n)
		}),
		Propagate: true,
	}
}

func withOnFailNow(f func()) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginAsync) {
			p.onFailNow = f
		}),
	}
}

func withContext(ctx context.Context) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginAsync) {
			p.parentCtx = ctx
		}),
	}
}
