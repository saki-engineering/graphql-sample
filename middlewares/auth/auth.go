package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
)

type userNameKey struct{}

const (
	tokenPrefix = "UT"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		if token == "" {
			next.ServeHTTP(w, req)
			return
		}

		userName, err := validateToken(token)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"reason": "invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), userNameKey{}, userName)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func GetUserName(ctx context.Context) (string, bool) {
	switch v := ctx.Value(userNameKey{}).(type) {
	case string:
		return v, true
	default:
		return "", false
	}
}

func validateToken(token string) (string, error) {
	tElems := strings.SplitN(token, "_", 2)
	if len(tElems) < 2 {
		return "", errors.New("invalid token")
	}

	tType, tUserName := tElems[0], tElems[1]
	if tType != tokenPrefix {
		return "", errors.New("invalid token")
	}
	return tUserName, nil
}
