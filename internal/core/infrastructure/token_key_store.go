package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TokenKeysPostgresError struct {
	Err error
}

func (e TokenKeysPostgresError) Error() string {
	return fmt.Sprintf("TokenKeyStore error: %v", e.Err)
}

type TokenKeyStore struct {
	postgresClient *sqlx.DB
	logger         *zap.Logger
	bufferChan     chan core.TokenKey
	errChan        chan error
}

func NewTokenKeyStore(ctx context.Context, postgresClient *sqlx.DB, logger *zap.Logger, bufferSize int) (*TokenKeyStore, error) {
	if postgresClient == nil {
		return nil, errors.New("postgresClient is nil")
	}
	bufferChan := make(chan core.TokenKey, bufferSize)
	errChan := make(chan error, 1)
	store := TokenKeyStore{
		postgresClient, logger, bufferChan, errChan,
	}
	go store.bufferRefillInfiniteLoop(ctx)
	return &store, nil
}

func (s *TokenKeyStore) Issue(ctx context.Context) (*core.TokenKey, error) {
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	for {
		select {
		case ti := <-s.bufferChan:
			return &ti, nil
		case err := <-s.errChan:
			return nil, fmt.Errorf("issue: %w", err)
		case <-ctx.Done():
			return nil, fmt.Errorf("issue: %w", ctx.Err())
		}
	}
}

func (s *TokenKeyStore) bufferRefillInfiniteLoop(ctx context.Context) {
	const targetRps = 100000
	refillFrequency := time.Duration(1+cap(s.bufferChan)*1000/targetRps) * time.Millisecond
	ticker := time.NewTicker(refillFrequency)

	for {
		select {
		case <-ticker.C:
			freeCapacity := cap(s.bufferChan) - len(s.bufferChan)
			if freeCapacity != 0 {
				batch, err := s.issueBatch(ctx, freeCapacity)
				if err != nil {
					s.errChan <- fmt.Errorf("bufferRefillInfiniteLoop error: %w", err)
					return
				}
				for _, ti := range batch {
					s.bufferChan <- *ti
				}
				ticker.Reset(refillFrequency)
			}
		case <-ctx.Done():
			s.errChan <- ctx.Err()
			return
		}
	}
}

func (p *TokenKeyStore) issueBatch(ctx context.Context, batchSize int) ([]*core.TokenKey, error) {
	var uniqueIDs []int64

	err := p.postgresClient.SelectContext(ctx, &uniqueIDs, `SELECT nextval('token_identifier') FROM generate_series(1, $1)`, batchSize)
	if err != nil {
		return nil, TokenKeysPostgresError{Err: err}
	}

	if len(uniqueIDs) != batchSize {
		return nil, TokenKeysPostgresError{Err: fmt.Errorf("IssueBatch error: incorrect number of unique ids from postgresClient: %d", len(uniqueIDs))}
	}

	tokens := make([]*core.TokenKey, 0, batchSize)
	for _, id := range uniqueIDs {
		token, err := core.TokenKey{}.New(id)
		if err != nil {
			return nil, TokenKeysPostgresError{Err: fmt.Errorf("IssueBatch error: failed to convert id %d: %w", id, err)}
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
