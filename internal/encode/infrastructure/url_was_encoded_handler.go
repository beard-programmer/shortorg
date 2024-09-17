package infrastructure

import (
	"context"
	"sync"
	"time"

	"github.com/beard-programmer/shortorg/internal/encode"
	"go.uber.org/zap"
)

type URLWasEncodedHandlerFn = func(ctx context.Context) <-chan error

type BatchSave interface {
	SaveMany(context.Context, []encode.UrlWasEncoded) error
}

func NewUrlWasEncodedHandler(
	logger *zap.Logger,
	store BatchSave,
	batchSize int,
	concurrency int,
	TChan <-chan encode.UrlWasEncoded,
) URLWasEncodedHandlerFn {
	retryPeriod := time.Duration(1+batchSize/40) * time.Millisecond
	return func(ctx context.Context) <-chan error {

		errChan := make(chan error, concurrency+1)

		process := func(ctx context.Context, batch []encode.UrlWasEncoded) error {
			err := store.SaveMany(ctx, batch)
			if err != nil {
				select {
				case errChan <- err:
				default:
					logger.Error("Error channel full, error discarded:", zap.Error(err))
				}
			}
			return nil
		}

		wg := sync.WaitGroup{}
		for i := 0; i < concurrency; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				ticker := time.NewTicker(retryPeriod)

				var batch []encode.UrlWasEncoded
				for {
					select {
					case element, ok := <-TChan:
						if !ok {
							if 0 < len(batch) {
								err := process(ctx, batch)
								if err != nil {
									errChan <- err
									return
								}
							}
							logger.Info("Input channel closed, worker shutting down")
							return
						}

						batch = append(batch, element)

						if batchSize <= len(batch) {
							err := process(ctx, batch)
							if err != nil {
								errChan <- err
								return
							}
							batch = nil
						}
						ticker.Reset(retryPeriod)
					case <-ticker.C:
						if 0 < len(batch) {
							err := process(ctx, batch)
							if err != nil {
								errChan <- err
								return
							}
							batch = nil
						}
						ticker.Reset(retryPeriod)
					case <-ctx.Done():
						if 0 < len(batch) {
							logger.Info("Context canceled, processing remaining batch before shutdown")
							err := process(ctx, batch)
							if err != nil {
								errChan <- err
								return
							}
						}
						logger.Info("NewUrlWasEncodedHandler shut down gracefully")
						return
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		return errChan
	}
}
