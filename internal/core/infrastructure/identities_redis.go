package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/redis/go-redis/v9"
)

type IdentitiesRedisError struct {
	Err error
}

func (e IdentitiesRedisError) Error() string {
	return fmt.Sprintf("IdentitiesRedis error: %v", e.Err)
}

type IdentitiesRedis struct {
	Redis *redis.Client
}

func (p *IdentitiesRedis) Issue(ctx context.Context) (*core.TokenKey, error) {
	cmd := p.Redis.IncrBy(ctx, "token_identifier", 2)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), context.Canceled) {
			return nil, cmd.Err()
		}

		return nil, IdentitiesRedisError{Err: cmd.Err()}
	}

	value := cmd.Val()
	return core.TokenKey{}.New(value)
}
