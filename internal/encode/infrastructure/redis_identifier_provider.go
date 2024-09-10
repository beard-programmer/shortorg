package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/common"
	"github.com/redis/go-redis/v9"
)

type RedisIdentityProviderError struct {
	Err error
}

func (e RedisIdentityProviderError) Error() string {
	return fmt.Sprintf("RedisIdentityProviderError error: %v", e.Err)
}

type RedisIdentifierProvider struct {
	Redis *redis.Client
}

func (p *RedisIdentifierProvider) ProduceTokenIdentifier(ctx context.Context) (*common.IntegerBase58Exp5To6, error) {
	cmd := p.Redis.IncrBy(ctx, "token_identifier", 2)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), context.Canceled) {
			return nil, cmd.Err()
		}

		return nil, RedisIdentityProviderError{Err: cmd.Err()}
	}

	value := cmd.Val()
	return new(common.IntegerBase58Exp5To6).FromInt(value)
}
