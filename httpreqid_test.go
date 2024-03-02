package httpreqid

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFromContext(t *testing.T) {
	tests := []struct {
		ctx context.Context
		rid string
	}{
		{context.Background(), ""},
		{context.WithValue(context.Background(), _contextKey, "foo"), "foo"},
		{context.WithValue(context.WithValue(context.Background(), _contextKey, "foo"), _contextKey, "bar"), "bar"},
	}

	for _, tt := range tests {
		if rid := FromContext(tt.ctx); rid != tt.rid {
			t.Errorf("FromContext(); got %q expect %q", rid, tt.rid)
		}
	}
}

func TestDefaultHeaders(t *testing.T) {
	if eq := reflect.DeepEqual(DefaultHeaders(), _defaultHeaders); !eq {
		t.Error("expect DefaultHeaders() returns the global _defaultHeaders")
	}
}

func IdentityGenerator(rid string) GeneratorFunc {
	return func(_ context.Context) string { return rid }
}

func TestHandler(t *testing.T) {
	g := IdentityGenerator("foo")

	t.Run("request doesn't have a request id", func(t *testing.T) {
		var visited bool
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			visited = true
			if rid := FromContext(r.Context()); rid != "foo" {
				t.Errorf("unxpected request id in context; got %q", rid)
			}
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		Handler(h, g).ServeHTTP(res, req)
		if !visited {
			t.Fatal("handler must be visited")
		}

		rid := res.Header().Get(DefaultHeaders()[0])
		if rid != "foo" {
			t.Errorf("unxpected request id in response; got %q", rid)
		}
	})

	t.Run("request already have a request id", func(t *testing.T) {
		var visited bool
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			visited = true
			if rid := FromContext(r.Context()); rid != "bar" {
				t.Errorf("unxpected request id in context; got %q", rid)
			}
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Correlation-ID", "bar")

		res := httptest.NewRecorder()
		Handler(h, g).ServeHTTP(res, req)
		if !visited {
			t.Fatal("handler must be visited")
		}

		rid := res.Header().Get("X-Correlation-ID")
		if rid != "bar" {
			t.Errorf("unxpected request id in response; got %q", rid)
		}
	})
}

type slogHandlerObserver struct {
	slog.Handler
	OnHandle func(ctx context.Context, rec slog.Record) error
}

func (obs *slogHandlerObserver) Handle(ctx context.Context, rec slog.Record) error {
	return obs.OnHandle(ctx, rec)
}

func TestSlogHandler_Handle(t *testing.T) {
	ridAttr := slog.String("request_id", "foo")

	t.Run("context doesn't have a request id", func(t *testing.T) {
		var visited bool
		h := &slogHandlerObserver{
			Handler: slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}),
			OnHandle: func(ctx context.Context, rec slog.Record) error {
				rec.Attrs(func(attr slog.Attr) bool {
					if attr.Equal(ridAttr) {
						t.Fail()
					}
					return true
				})
				visited = true
				return nil
			},
		}

		l := slog.New(LogHandler(h))
		l.InfoContext(context.Background(), "a message")
		if !visited {
			t.Fatal("handler must be visited")
		}
	})

	t.Run("context have a request id", func(t *testing.T) {
		var visited, found bool
		h := &slogHandlerObserver{
			Handler: slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}),
			OnHandle: func(ctx context.Context, rec slog.Record) error {
				rec.Attrs(func(attr slog.Attr) bool {
					found = attr.Equal(ridAttr)
					return !found
				})
				visited = true
				return nil
			},
		}

		l := slog.New(LogHandler(h))

		ctx := context.WithValue(context.Background(), _contextKey, "foo")
		l.InfoContext(ctx, "a message")
		if !visited {
			t.Fatal("handler must be visited")
		}

		if !found {
			t.Errorf("expecting record contains a request id")
		}
	})
}
