package infrastructure

import (
	"context"
	"sync"
	"time"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/encode"
)

type URLWasEncodedHandlerFn = func(ctx context.Context) <-chan error

type BatchSave interface {
	SaveMany(context.Context, []encode.URLWasEncoded) error
}

func NewUrlWasEncodedHandler(
	logger *appLogger.AppLogger,
	store BatchSave,
	batchSize int,
	concurrency int,
	TChan <-chan encode.URLWasEncoded,
) URLWasEncodedHandlerFn {
	retryPeriod := time.Duration(1+batchSize/40) * time.Millisecond
	return func(ctx context.Context) <-chan error {

		errChan := make(chan error, concurrency+1)

		process := func(ctx context.Context, batch []encode.URLWasEncoded) error {
			err := store.SaveMany(ctx, batch)
			if err != nil {
				select {
				case errChan <- err:
				default:
					logger.ErrorContext(ctx, "Error channel full, error discarded:", err)
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

				var batch []encode.URLWasEncoded
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
							logger.WarnContext(ctx, "Input channel closed, worker shutting down")
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
							logger.WarnContext(ctx, "Context canceled, processing remaining batch before shutdown")
							err := process(ctx, batch)
							if err != nil {
								errChan <- err
								return
							}
						}
						logger.WarnContext(ctx, "NewUrlWasEncodedHandler shut down gracefully")
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
