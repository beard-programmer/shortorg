package infrastructure

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

func ProcessChan[T any](
	logger *zap.SugaredLogger, processBatch func(context.Context, []T) error, // Dependencies
) func(ctx context.Context, batchSize int, concurrency int, ticketDuration time.Duration, TChan <-chan T) <-chan error {
	return func(ctx context.Context, batchSize int, concurrency int, tickerDuration time.Duration, TChan <-chan T) <-chan error {

		errChan := make(chan error, concurrency+1)

		process := func(ctx context.Context, batch []T) {
			err := processBatch(ctx, batch)
			if err != nil {
				select {
				case errChan <- err:
				default:
					logger.Errorln("Error channel full, error discarded:", "error", err)
				}
			}
		}

		wg := sync.WaitGroup{}
		for i := 0; i < concurrency; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				ticker := time.NewTicker(tickerDuration)

				var batch []T
				processBatchCtx := context.Background()
				for {
					select {
					case element, ok := <-TChan:
						if !ok {
							if 0 < len(batch) {
								process(processBatchCtx, batch)
							}
							logger.Infoln("Input channel closed, worker shutting down")
							return
						}

						batch = append(batch, element)

						if batchSize <= len(batch) {
							process(processBatchCtx, batch)
							batch = nil
						}
						ticker.Reset(tickerDuration)
					case <-ticker.C:
						if 0 < len(batch) {
							process(processBatchCtx, batch)
							batch = nil
						}
						ticker.Reset(tickerDuration)
					case <-ctx.Done():
						if 0 < len(batch) {
							logger.Infow("Context canceled, processing remaining batch before shutdown")
							process(processBatchCtx, batch)
						}
						logger.Infow("Worker shut down gracefully")
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
