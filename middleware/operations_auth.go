package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"
)

// BearerToken protects an operational handler when a token is configured.
// An empty token leaves the handler open for explicitly selected local modes.
func BearerToken(next http.Handler, token string) http.Handler {
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		authorization := r.Header.Get("Authorization")
		provided := ""
		if strings.HasPrefix(authorization, prefix) {
			provided = strings.TrimPrefix(authorization, prefix)
		}
		providedDigest := sha256.Sum256([]byte(provided))
		tokenDigest := sha256.Sum256([]byte(token))
		if subtle.ConstantTimeCompare(providedDigest[:], tokenDigest[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
