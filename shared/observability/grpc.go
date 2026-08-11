package observability

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary interceptor that ensures every
// request has a request ID (read from the X-Request-ID metadata key when
// present, otherwise generated), attaches it to the context and response
// metadata, and logs the RPC lifecycle: service, method, gRPC status,
// and duration.
//
// The interceptor is intended to be chained after the existing
// grpc_recovery interceptor. It never logs request payloads.
func UnaryServerInterceptor(logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		md, _ := metadata.FromIncomingContext(ctx)
		reqID := ""
		if vals := md.Get(Header); len(vals) > 0 {
			reqID = vals[0]
		}

		ctx, reqID = GetOrCreateWithValue(ctx, reqID)

		resp, err := handler(ctx, req)

		event := logger.Info()
		event.Str("request_id", reqID)
		event.Str("rpc", info.FullMethod)
		event.Int64("duration_ms", time.Since(start).Milliseconds())
		if err != nil {
			code := status.Code(err)
			event = logger.Warn().Str("request_id", reqID).Str("rpc", info.FullMethod)
			event.Int64("duration_ms", time.Since(start).Milliseconds())
			event.Str("grpc_code", code.String())
			event.Err(err)
		} else {
			event.Str("grpc_code", "OK")
		}
		event.Msg("grpc request")

		// Echo the request ID back to the client in response metadata.
		if err := grpc.SetHeader(ctx, metadata.Pairs(Header, reqID)); err != nil {
			// Header may already be sent; the ID is best-effort.
			logger.Debug().Err(err).Msg("failed to set request id response header")
		}

		return resp, err
	}
}

// GetOrCreateWithValue returns a context carrying the provided id when
// non-empty, otherwise a new request ID, along with the effective ID.
func GetOrCreateWithValue(ctx context.Context, id string) (context.Context, string) {
	if id != "" {
		return WithRequestID(ctx, id), id
	}
	return GetOrCreate(ctx)
}