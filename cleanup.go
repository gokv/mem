package mem

import (
	"context"
	"time"
)

func (s *Store) Cleanup(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, e := range s.m {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !e.validAt(now) {
			delete(s.m, k)
		}
	}
}

func start(fn func(context.Context), timeout, interval time.Duration) (stop func()) {
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fnCtx, fnCancel := context.WithTimeout(ctx, timeout)
				fn(fnCtx)

				fnCancel()
				time.Sleep(interval)
			}
		}
	}()
	return stop
}
