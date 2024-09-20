package encode

import (
	"context"
	"sync"
	"time"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
)

type SaveEncodedURLJob = func(ctx context.Context) <-chan error

func NewSaveEncodedURLJob(
	logger *appLogger.AppLogger,
	store EncodedURLStore,
	batchSize int,
	concurrency int,
	tChan <-chan URLWasEncoded,
) SaveEncodedURLJob {
	retryPeriod := time.Duration(1+batchSize/40) * time.Millisecond
	return func(ctx context.Context) <-chan error {

		errChan := make(chan error, concurrency+1)

		process := func(ctx context.Context, batch []URLWasEncoded) error {
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
		for range concurrency {
			wg.Add(1)

			go func() {
				defer wg.Done()
				ticker := time.NewTicker(retryPeriod)

				var batch []URLWasEncoded
				for {
					select {
					case element, ok := <-tChan:
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
						logger.WarnContext(ctx, "NewSaveEncodedURLJob shut down gracefully")
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
