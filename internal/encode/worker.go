package encode

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type UrlSaveWorker struct {
	eventChan chan UrlWasEncoded
	logger    *zap.SugaredLogger
	provider  SaveEncodedUrlProvider
}

type SaveEncodedUrlProvider interface {
	SaveEncodedURL(context.Context, []EncodedUrl) error
}

func (_ *UrlSaveWorker) New(provider SaveEncodedUrlProvider, logger *zap.SugaredLogger) *UrlSaveWorker {
	return &UrlSaveWorker{
		eventChan: make(chan UrlWasEncoded, 100),
		logger:    logger,
		provider:  provider,
	}
}

func (w *UrlSaveWorker) GetEventChan() chan<- UrlWasEncoded {
	return w.eventChan
}

func (w *UrlSaveWorker) Stop() {
	close(w.eventChan)
}

func (w *UrlSaveWorker) Start(ctx context.Context) {
	go func() {
		var batch []EncodedUrl
		batchTicker := time.NewTicker(1 * time.Second) // Ticker for periodic batch processing
		defer batchTicker.Stop()                       // Stop the ticker when the goroutine exits

		batchLimit := 10 // Define your batch limit

		for {
			select {
			case <-ctx.Done():
				// When context is canceled, process the remaining batch (if any) before exiting
				if len(batch) > 0 {
					w.logger.Info("Context canceled, processing remaining batch before shutdown")
					w.processBatch(ctx, batch)
				}
				w.logger.Info("Worker shut down gracefully")
				return

			case event := <-w.eventChan:
				// Add new events to the batch
				batch = append(batch, EncodedUrl{event.URL, &event.Token.TokenIdentifier})

				// If the batch limit is reached, process the batch
				if len(batch) >= batchLimit {
					w.processBatch(ctx, batch)
					batch = nil // Reset batch after processing
				}

			case <-batchTicker.C:
				// Periodically process the batch even if the batch limit is not reached
				if len(batch) > 0 {
					w.processBatch(ctx, batch)
					batch = nil // Reset batch after processing
				}
			}
		}
	}()
}

func (w *UrlSaveWorker) processBatch(ctx context.Context, batch []EncodedUrl) {
	batchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := w.provider.SaveEncodedURL(batchCtx, batch)
	if err != nil {
		w.logger.Errorf("Failed to save encoded URL: %v", err)
	}
}
