package utils

import "context"

func Consume(ctx context.Context, fn func(ctx context.Context)) {
	for {
		fn(ctx)
	}
}