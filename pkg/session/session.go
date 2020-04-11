package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

type Session struct {
	ID      string
	UserID  string
	Expires time.Time
}

const IDLength int = 32

var (
	ErrNoAuth = errors.New("No session found in context")
)

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes), err
}

func NewSession(userID string) (*Session, error) {
	id, err := generateRandomString(IDLength)
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:      id,
		UserID:  userID,
		Expires: time.Now().Add(time.Hour),
	}, nil
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value("session").(*Session)
	if !ok || sess == nil {
		return nil, ErrNoAuth
	}
	return sess, nil
}
