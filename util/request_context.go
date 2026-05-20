package util

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const RequestIDHeader = "X-Request-ID"

type requestContextKey string

const requestIDContextKey requestContextKey = "request_id"

// NewRequestID creates a compact trace identifier for request-scoped logs.
func NewRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err == nil {
		return hex.EncodeToString(bytes)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ContextWithRequestID stores the request ID in a context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

// RequestIDFromContext returns the request ID stored in a context.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
