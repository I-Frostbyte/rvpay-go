package observability

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewRequestID(t *testing.T) {
	a := NewRequestID()
	b := NewRequestID()
	if a == "" {
		t.Fatal("NewRequestID returned empty string")
	}
	if a == b {
		t.Fatalf("NewRequestID returned duplicate IDs %q, %q", a, b)
	}
}

func TestRequestIDContext(t *testing.T) {
	ctx := WithRequestID(context.Background(), "abc")
	if got := RequestIDFromContext(ctx); got != "abc" {
		t.Fatalf("RequestIDFromContext = %q, want %q", got, "abc")
	}

	ctx2, id := GetOrCreate(context.Background())
	if id == "" {
		t.Fatal("GetOrCreate generated an empty request ID")
	}
	if got := RequestIDFromContext(ctx2); got != id {
		t.Fatalf("GetOrCreate context id = %q, want %q", got, id)
	}

	// Existing ID must be preserved.
	ctx3, id3 := GetOrCreate(WithRequestID(context.Background(), "keep"))
	if id3 != "keep" {
		t.Fatalf("GetOrCreate id = %q, want %q", id3, "keep")
	}
	if got := RequestIDFromContext(ctx3); got != "keep" {
		t.Fatalf("GetOrCreate context id = %q, want %q", got, "keep")
	}
}

func TestGetOrCreateWithValue(t *testing.T) {
	ctx, id := GetOrCreateWithValue(context.Background(), "from-header")
	if id != "from-header" {
		t.Fatalf("GetOrCreateWithValue id = %q, want %q", id, "from-header")
	}
	if got := RequestIDFromContext(ctx); got != "from-header" {
		t.Fatalf("context id = %q, want %q", got, "from-header")
	}
}

func TestUnaryServerInterceptor_PropagatesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	interceptor := UnaryServerInterceptor(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/clientsgrpc.ClientsService/GetClient"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		if got := RequestIDFromContext(ctx); got != "req-123" {
			t.Errorf("handler request_id = %q, want %q", got, "req-123")
		}
		return "resp", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(Header, "req-123"))
	resp, err := interceptor(ctx, struct{}{}, info, handler)
	if err != nil {
		t.Fatalf("interceptor error: %v", err)
	}
	if resp != "resp" {
		t.Fatalf("resp = %v, want %v", resp, "resp")
	}

	out := buf.String()
	if !strings.Contains(out, `"request_id":"req-123"`) {
		t.Errorf("log missing request_id: %s", out)
	}
	if !strings.Contains(out, `"rpc":"/clientsgrpc.ClientsService/GetClient"`) {
		t.Errorf("log missing rpc: %s", out)
	}
	if !strings.Contains(out, `"grpc_code":"OK"`) {
		t.Errorf("log missing grpc_code OK: %s", out)
	}
	if !strings.Contains(out, `"duration_ms":`) {
		t.Errorf("log missing duration_ms: %s", out)
	}
}

func TestUnaryServerInterceptor_GeneratesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	interceptor := UnaryServerInterceptor(logger)
	var seen string
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		seen = RequestIDFromContext(ctx)
		return "resp", nil
	}

	_, err := interceptor(context.Background(), struct{}{}, &grpc.UnaryServerInfo{FullMethod: "/x/X/Y"}, handler)
	if err != nil {
		t.Fatalf("interceptor error: %v", err)
	}
	if seen == "" {
		t.Fatal("handler did not receive a generated request ID")
	}
}

func TestUnaryServerInterceptor_ErrorClassification(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	interceptor := UnaryServerInterceptor(logger)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "client not found")
	}

	_, err := interceptor(context.Background(), struct{}{}, &grpc.UnaryServerInfo{FullMethod: "/x/X/Y"}, handler)
	if err == nil {
		t.Fatal("expected error from handler")
	}

	out := buf.String()
	if !strings.Contains(out, `"grpc_code":"NotFound"`) {
		t.Errorf("log missing grpc_code NotFound: %s", out)
	}
	if !strings.Contains(out, "client not found") {
		t.Errorf("log missing error detail: %s", out)
	}
	if !strings.Contains(out, `"level":"warn"`) {
		t.Errorf("failed RPC should be logged at warn: %s", out)
	}
}

func TestAccessLog_PropagatesAndEchoesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(Header); got != "http-1" {
			t.Errorf("handler request_id = %q, want %q", got, "http-1")
		}
		if got := RequestIDFromContext(r.Context()); got != "http-1" {
			t.Errorf("context request_id = %q, want %q", got, "http-1")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/public/clients", nil)
	req.Header.Set(Header, "http-1")

	AccessLog(logger)(inner).ServeHTTP(rec, req)

	if got := rec.Header().Get(Header); got != "http-1" {
		t.Errorf("response header = %q, want %q", got, "http-1")
	}

	out := buf.String()
	if !strings.Contains(out, `"method":"GET"`) {
		t.Errorf("log missing method: %s", out)
	}
	if !strings.Contains(out, `"path":"/v1/public/clients"`) {
		t.Errorf("log missing path: %s", out)
	}
	if !strings.Contains(out, `"status":200`) {
		t.Errorf("log missing status: %s", out)
	}
	if !strings.Contains(out, `"request_id":"http-1"`) {
		t.Errorf("log missing request_id: %s", out)
	}
}

func TestAccessLog_GeneratesRequestIDWhenAbsent(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/public/platforms", nil)

	AccessLog(logger)(inner).ServeHTTP(rec, req)

	if got := rec.Header().Get(Header); got == "" {
		t.Error("response missing generated request ID")
	}
}

func TestAccessLog_HealthzAtDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	AccessLog(logger)(inner).ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, `"level":"debug"`) {
		t.Errorf("healthz access log should be debug: %s", out)
	}
}

func TestAccessLog_NoPayloadLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/public/clients", strings.NewReader(`{"secret":"sensitive-body"}`))
	req.Header.Set("Authorization", "Bearer token-abc")

	AccessLog(logger)(inner).ServeHTTP(rec, req)

	out := buf.String()
	if strings.Contains(out, "sensitive-body") || strings.Contains(out, "token-abc") {
		t.Errorf("access log leaked request body or authorization header: %s", out)
	}
	if !strings.Contains(out, `"status":400`) {
		t.Errorf("log missing status 400: %s", out)
	}
}