package apiServer

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) getServerMux(ctx context.Context) (*chi.Mux, error) {
	mux := s.wrapWithDefaultMiddlewares(chi.NewMux())

	mux.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/encode", encode.HttpHandlerFunc(s.logger(ctx), s.encodeFn))
		r.Post("/decode", decode.HttpHandlerFunc(s.logger(ctx), s.decodeFn))
	})
	return mux, nil
}

func (s *Server) wrapWithDefaultMiddlewares(mux *chi.Mux) *chi.Mux {
	mux.Use(middleware.Logger)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Mount("/debug", middleware.Profiler())

	return mux
}
