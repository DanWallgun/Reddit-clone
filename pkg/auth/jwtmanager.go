package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"redditclone/pkg/user"
)

var (
	JWTSecretKey = []byte("MyC00lKeyYouCan'tHackIt")
)

func GenerateToken(u *user.User, sessionID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.User{
			Login: u.Login,
			ID:    u.ID,
		},
		"session_id": sessionID,
	})
	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GetToken(tokenString string) (*jwt.Token, error) {
	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return JWTSecretKey, nil
	}
	token, err := jwt.Parse(tokenString, hashSecretGetter)
	return token, err
}
