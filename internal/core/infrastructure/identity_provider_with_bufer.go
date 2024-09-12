package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/beard-programmer/shortorg/internal/core"
	"go.uber.org/zap"
)

type IdentityProviderWithBufferError struct {
	Err error
}

func (e IdentityProviderWithBufferError) Error() string {
	return fmt.Sprintf("IdentityProviderWithBuffer error: %v", e.Err)
}

type IdentityProviderBulk interface {
	GenerateMany(ctx context.Context, batchSize int) ([]*core.TokenKey, error)
}

type ProviderBulk interface {
	ProvideBulk(ctx context.Context, batchSize int) ([]*core.TokenKey, error)
}

type IdentityProviderWithBuffer struct {
	provider     IdentityProviderBulk
	logger       *zap.SugaredLogger
	identityChan chan core.TokenKey
}

func NewIdentityProviderWithBuffer(ctx context.Context, provider IdentityProviderBulk, logger *zap.SugaredLogger, bufferSize int) (*IdentityProviderWithBuffer, <-chan error) {
	producer := IdentityProviderWithBuffer{
		provider:     provider,
		logger:       logger,
		identityChan: make(chan core.TokenKey, bufferSize),
	}

	errChan := make(chan error, 1)

	producer.refillInfiniteLoop(ctx, errChan)

	return &producer, errChan
}

func (s *IdentityProviderWithBuffer) Issue(ctx context.Context) (*core.TokenKey, error) {
	for {
		select {
		case ti := <-s.identityChan:
			return &ti, nil
		case <-time.After(10 * time.Millisecond):
			s.logger.Warnln("10 milliseconds waiting for identity!")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (s *IdentityProviderWithBuffer) refillInfiniteLoop(ctx context.Context, errChan chan<- error) {
	duration := s.refillDurationForTargetRps(100000)
	go func() {
		ticker := time.NewTicker(duration)

		for {
			select {
			case <-ticker.C:
				freeCapacity := cap(s.identityChan) - len(s.identityChan)
				if freeCapacity != 0 {
					batch, err := s.provider.GenerateMany(ctx, freeCapacity)
					if err != nil {
						errChan <- IdentityProviderWithBufferError{fmt.Errorf("refillInfiniteLoop error: %w", err)}
						return
					}
					for _, ti := range batch {
						s.identityChan <- *ti
					}
					ticker.Reset(duration)
				}
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}
	}()
}

func (s *IdentityProviderWithBuffer) refillDurationForTargetRps(targetRps int) time.Duration {
	return time.Duration(1+cap(s.identityChan)*1000/targetRps) * time.Millisecond
}
