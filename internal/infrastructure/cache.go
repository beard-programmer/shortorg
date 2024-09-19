package infrastructure

import (
	"context"
	"fmt"

	"github.com/dgraph-io/ristretto"
	ekoCache "github.com/eko/gocache/lib/v4/cache"
	ristrettoStore "github.com/eko/gocache/store/ristretto/v4"
)

type CacheInMemory[T any] struct {
	cacheManager ekoCache.Cache[T]
}

func NewCache[T any](cfg cacheConfig) (Cache[T], error) {
	if !cfg.UseCache {
		c := CacheMock[T]{}
		return &c, nil
	}

	ristrettoCache, err := ristretto.NewCache(
		&ristretto.Config{
			NumCounters: cfg.MaxNumberOfElements,
			MaxCost:     cfg.MaxMbSize,
			BufferItems: 64,
		},
	)
	if err != nil {
		return nil, err
	}
	rStore := ristrettoStore.NewRistretto(ristrettoCache)
	cacheManager := ekoCache.New[T](rStore)
	return &CacheInMemory[T]{*cacheManager}, nil
}

func (c *CacheInMemory[T]) Get(ctx context.Context, key any) (T, error) {
	return c.cacheManager.Get(ctx, key)
}

func (c *CacheInMemory[T]) Set(ctx context.Context, key any, value T) error {
	return c.cacheManager.Set(ctx, key, value)
}

type CacheMock[T any] struct{}

func (m *CacheMock[T]) Get(_ context.Context, key any) (T, error) {
	return *new(T), fmt.Errorf("not found in MOCK cache %v", key)
}

func (m *CacheMock[T]) Set(_ context.Context, _ any, _ T) error {
	return nil
}
