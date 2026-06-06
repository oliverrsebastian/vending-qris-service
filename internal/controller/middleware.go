package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	zax "github.com/yuseferi/zax/v2"
	"go.uber.org/zap"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "RequestID"
)

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = fmt.Sprintf("%v", strings.ReplaceAll(uuid.NewString(), "-", ""))
			r.Header.Set(RequestIDHeader, requestID)
		}

		ctx = zax.Set(ctx, []zap.Field{zap.String(RequestIDKey, requestID)})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func Authenticate(authKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.WriteHeader(http.StatusUnauthorized)
			}

			if auth != authKey {
				w.WriteHeader(http.StatusUnauthorized)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
