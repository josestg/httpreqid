package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/josestg/httpreqid"
)

func main() {
	log := slog.New(
		// wrap JSONHandler with request id log handler.
		httpreqid.LogHandler(
			slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}),
		),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// logging with context automatically adds the request ID to the log request,
		// so there's no need to add it manually.
		log.InfoContext(ctx, "ping requested")

		// we can still retrieve the request ID from the context if needed.
		rid := httpreqid.FromContext(ctx)

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "PONG! request id %q", rid)
	})

	requestIDGenerator := generator()
	srv := http.Server{
		Addr: ":8080",
		// wrap the http handler with the request id http handler.
		Handler: httpreqid.Handler(mux, requestIDGenerator),
	}

	log.Info("server is listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("listen and server failed", "error", err)
		os.Exit(1)
	}
}

func generator() httpreqid.GeneratorFunc {
	return func(_ context.Context) string {
		buf := make([]byte, 16)
		_, err := io.ReadFull(rand.Reader, buf)
		if err != nil {
			panic(err)
		}
		return hex.EncodeToString(buf)
	}
}
