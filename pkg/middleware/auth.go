package middleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"net/http"
	"redditclone/pkg/auth"
	"redditclone/pkg/session"
)

func Auth(sm *session.SessionsManager, logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inToken := r.Header.Get("Authorization")
		logger.Infow("AUTHORIZATION MIDDLEWARE",
			"token", inToken,
		)

		if len(inToken) < len("Bearer ") {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}

		tokenString := inToken[len("Bearer "):]

		token, err := auth.GetToken(tokenString)
		if err != nil || !token.Valid {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}

		payload, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}

		userData, ok := payload["user"].(map[string]interface{})
		if !ok {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}
		userID, ok := userData["id"].(string)
		if !ok {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}
		userLogin, ok := userData["username"].(string)
		if !ok {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}
		sessionID, ok := payload["session_id"].(string)
		if !ok {
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}

		logger.Infow("AUTHORIZATION PASSED",
			"user_id", userID,
			"user_login", userLogin,
			"session_id", sessionID,
		)

		ctx := r.Context()

		sess, err := sm.Check(sessionID)
		if err != nil {
			logger.Infow("AUTHORIZATION. Can't check session",
				"token_string", tokenString,
			)
			http.Error(w, `Unauthorized`, http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, "session", sess)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
