package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
	"github.com/jmoiron/sqlx"
)

var errLinkKeyStore = errors.New("errLinkKeyStore")

type LinkKeyStore struct {
	postgresClient *sqlx.DB
	logger         *logger.AppLogger
	bufferChan     chan core.LinkKey
	errChan        chan error
}

func NewLinkKeyStore(
	ctx context.Context,
	logger *logger.AppLogger,
	postgresClient *sqlx.DB,
	config tokenStoreConfig,
) (*LinkKeyStore, error) {
	if postgresClient == nil {
		return nil, fmt.Errorf("%w: NewLinkKeyStore: postgresClient is not provided", errLinkKeyStore)
	}
	bufferChan := make(chan core.LinkKey, config.BufferSize)
	errChan := make(chan error, 1)
	store := LinkKeyStore{
		postgresClient, logger, bufferChan, errChan,
	}
	go store.bufferRefillInfiniteLoop(ctx)
	return &store, nil
}

const issueTimeout = 50 * time.Millisecond

func (s *LinkKeyStore) Issue(ctx context.Context) (*core.LinkKey, error) {
	ctx, cancel := context.WithTimeout(ctx, issueTimeout)
	defer cancel()

	for {
		select {
		case ti := <-s.bufferChan:
			return &ti, nil
		case err := <-s.errChan:
			return nil, fmt.Errorf("%w: issue: %s", errLinkKeyStore, err)
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: issue: %s", errLinkKeyStore, ctx.Err())
		}
	}
}

func (s *LinkKeyStore) bufferRefillInfiniteLoop(ctx context.Context) {
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
					s.errChan <- fmt.Errorf("%w: bufferRefillInfiniteLoop : %s", errLinkKeyStore, err)
					return
				}
				for _, tokenKey := range batch {
					s.bufferChan <- *tokenKey
				}
				ticker.Reset(refillFrequency)
			}
		case <-ctx.Done():
			s.errChan <- ctx.Err()
			return
		}
	}
}

func (s *LinkKeyStore) issueBatch(ctx context.Context, batchSize int) ([]*core.LinkKey, error) {
	var uniqueIDs []int64

	err := s.postgresClient.SelectContext(
		ctx,
		&uniqueIDs,
		`SELECT nextval('token_identifier') FROM generate_series(1, $1)`,
		batchSize,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: issueBatch: %s", errLinkKeyStore, err)
	}

	if len(uniqueIDs) != batchSize {
		return nil, fmt.Errorf(
			"%w: issueBatch: incorrect number of unique ids from postgresClient : %s",
			errLinkKeyStore,
			err,
		)
	}

	tokens := make([]*core.LinkKey, 0, batchSize)
	for _, id := range uniqueIDs {
		token, newLinkKeyErr := core.NewLinkKey(id)
		if newLinkKeyErr != nil {
			return nil,
				fmt.Errorf("%w: issueBatch: failed to convert id: %s", errLinkKeyStore, newLinkKeyErr)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
