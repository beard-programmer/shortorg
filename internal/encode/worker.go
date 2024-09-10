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

func (w *UrlSaveWorker) Start(ctx context.Context, errChan chan<- error) {
	go func() {
		var batch []EncodedUrl
		batchTicker := time.NewTicker(1 * time.Second)
		defer batchTicker.Stop()

		batchLimit := 100

		for {
			select {
			case <-ctx.Done():
				if len(batch) > 0 {
					w.logger.Info("Context canceled, processing remaining batch before shutdown")
					w.processBatch(ctx, batch, errChan)
				}
				w.logger.Info("Worker shut down gracefully")
				return

			case event := <-w.eventChan:
				batch = append(batch, EncodedUrl{event.URL, &event.Token.TokenIdentifier})

				if len(batch) >= batchLimit {
					w.processBatch(ctx, batch, errChan)
					batch = nil
				}

			case <-batchTicker.C:
				if len(batch) > 0 {
					w.processBatch(ctx, batch, errChan)
					batch = nil
				}
			}
		}
	}()
}

func (w *UrlSaveWorker) processBatch(ctx context.Context, batch []EncodedUrl, errChan chan<- error) {
	batchCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := w.provider.SaveEncodedURL(batchCtx, batch)
	if err != nil {
		w.logger.Errorf("Failed to save encoded URL: %v", err)
		errChan <- err
	}
}
