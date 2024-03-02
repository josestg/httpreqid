package httpreqid

import (
	"context"
	"log/slog"
	"net/http"
)

var (
	_contextKey     = &contextKey{}
	_defaultHeaders = []string{
		"X-Request-ID",
		"X-Correlation-ID",
		"X-Trace-ID",
		"Request-ID",
		"Correlation-ID",
		"Trace-ID",
	}
)

type contextKey struct{}

// FromContext retrieves the request ID from the provided context. If not found, it returns an empty string.
func FromContext(ctx context.Context) string {
	rid, _ := ctx.Value(_contextKey).(string)
	return rid
}

// DefaultHeaders returns the common header keys used to propagate the request ID.
func DefaultHeaders() []string { return _defaultHeaders }

// Generator defines the interface for generating unique request IDs.
type Generator interface {
	// Generate generates a unique request id for each call.
	Generate(ctx context.Context) string
}

// GeneratorFunc is an adapter used to make an ordinary function with the same signature implement the Generator.
type GeneratorFunc func(ctx context.Context) string

func (f GeneratorFunc) Generate(ctx context.Context) string { return f(ctx) }

// Handler decorates the provided http.Handler with middleware that appends a request ID to the response header and
// the request context. If no header names are provided, it defaults to DefaultHeaders(). Otherwise, it uses the given
// header names. The first found header name in the request header is used to propagate the request ID. If none are
// found, it uses the first header name.
func Handler(h http.Handler, g Generator, headerNames ...string) http.Handler {
	if len(headerNames) == 0 {
		headerNames = DefaultHeaders()
	}
	return middleware(g, headerNames)(h)
}

func middleware(g Generator, headerNames []string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			name, value, found := getHeader(r.Header, headerNames)
			if !found {
				name, value = headerNames[0], g.Generate(ctx)
			}

			// propagate the request id to the response header.
			w.Header().Set(name, value)

			// set the request id to the context.
			ctx = context.WithValue(ctx, _contextKey, value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getHeader(h http.Header, keys []string) (name, value string, found bool) {
	for _, k := range keys {
		if v := h.Get(k); v != "" {
			return k, v, true
		}
	}
	return "", "", false
}

type slogHandler struct {
	attr string
	slog.Handler
}

// LogHandler decorates the provided slog.Handler with request ID logging,
// appending the request ID to the log record if it's available in the context.
// By default, it uses "request_id" as the record key. To customize the key, use LogHandlerWithKey.
func LogHandler(h slog.Handler) slog.Handler {
	return LogHandlerWithKey(h, "request_id")
}

// LogHandlerWithKey decorates the provided slog.Handler with request ID logging,
// using the specified key to append the request ID to the log record if available.
func LogHandlerWithKey(h slog.Handler, key string) slog.Handler {
	return &slogHandler{Handler: h, attr: key}
}

// Handle adds request id to the Record if the provided context has request id.
func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	if rid := FromContext(ctx); len(rid) > 0 {
		r.AddAttrs(slog.String(h.attr, rid))
	}
	return h.Handler.Handle(ctx, r)
}
