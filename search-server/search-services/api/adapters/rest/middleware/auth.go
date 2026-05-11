package middleware

import (
	"context"
	"net/http"
	"strings"

	"yadro.com/course/api/core"
)

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (core.UserPermissions, error)
}

func Auth(next http.HandlerFunc, verifier TokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := tokenFromRequest(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userPerms, err := verifier.Verify(r.Context(), token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := core.WithUser(r.Context(), userPerms)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func AdminAuth(next http.HandlerFunc, verifier TokenVerifier) http.HandlerFunc {
	return Auth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := core.UserFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}, verifier)
}

func tokenFromRequest(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}

	const prefix = "Token "
	if !strings.HasPrefix(auth, prefix) {
		return "", false
	}

	token := strings.TrimPrefix(auth, prefix)
	if token == "" {
		return "", false
	}

	return token, true
}
