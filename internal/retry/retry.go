package retry

import (
	"context"
	"time"
)

func Do(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		err := fn()

		if err == nil {
			return nil
		}

		if i == attempts-1 {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(baseDelay):
			baseDelay *= 2
		}
	}

	return nil
}
