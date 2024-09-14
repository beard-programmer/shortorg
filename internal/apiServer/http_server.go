package apiServer

import (
	"context"
	"net/http"

	"github.com/beard-programmer/shortorg/internal/apiServer/middleware"
	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
)

func (s *Server) getServerMux(ctx context.Context) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	mux.Handle("/encode", encode.HttpHandler(s.logger(ctx), s.encodeFn))
	mux.Handle("/decode", decode.HttpHandler(s.logger(ctx), s.decodeFn))
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

	return mux, nil
}

func (s *Server) wrapWithDefaultMiddlewares(h http.Handler) http.Handler {
	h = middleware.LoggingMiddleware(s._logger, h)
	return h
}
