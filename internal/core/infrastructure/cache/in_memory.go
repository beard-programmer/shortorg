package cache

import (
	"context"
	"fmt"

	"github.com/dgraph-io/ristretto"
	ekoCache "github.com/eko/gocache/lib/v4/cache"
	ristrettoStore "github.com/eko/gocache/store/ristretto/v4"
)

type InMemory[T any] struct {
	cacheManager ekoCache.Cache[T]
}

func NewInMemory[T any](cgx Config) (*InMemory[T], error) {
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cgx.MaxNumberOfElements,
		MaxCost:     cgx.MaxMbSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	rStore := ristrettoStore.NewRistretto(ristrettoCache)
	cacheManager := ekoCache.New[T](rStore)
	return &InMemory[T]{*cacheManager}, nil
}

func (c *InMemory[T]) Get(ctx context.Context, key any) (T, error) {
	return c.cacheManager.Get(ctx, key)
}

func (c *InMemory[T]) Set(ctx context.Context, key any, value T) error {
	return c.cacheManager.Set(ctx, key, value)
}

type MockCache[T any] struct{}

func (m *MockCache[T]) Get(_ context.Context, key any) (T, error) {
	return *new(T), fmt.Errorf("not found in MOCK cache %v", key)
}

func (m *MockCache[T]) Set(_ context.Context, _ any, _ T) error {
	return nil
}
