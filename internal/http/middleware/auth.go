package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gamidoc/backend/internal/http/response"
	"github.com/gamidoc/backend/internal/token"
)

type authUserIDKey struct{}

func RequireAuth(manager *token.Manager, blacklists ...*token.Blacklist) func(http.Handler) http.Handler {
	var blacklist *token.Blacklist
	if len(blacklists) > 0 {
		blacklist = blacklists[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if manager == nil {
				response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
				return
			}

			header := r.Header.Get("Authorization")
			if header == "" {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing bearer token", nil)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid bearer token", nil)
				return
			}

			claims, err := manager.Parse(parts[1])
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid bearer token", nil)
				return
			}

			// Check jti blacklist
			if blacklist != nil && claims.ID != "" {
				revoked, err := blacklist.IsBlacklisted(r.Context(), claims.ID)
				if err == nil && revoked {
					response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token has been revoked", nil)
					return
				}
			}

			ctx := context.WithValue(r.Context(), authUserIDKey{}, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAuthUserID(ctx context.Context) string {
	value, ok := ctx.Value(authUserIDKey{}).(string)
	if !ok {
		return ""
	}
	return value
}
