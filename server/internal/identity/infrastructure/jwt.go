package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct{ secret []byte }

func NewJWTManager(secret []byte) JWTManager {
	return JWTManager{secret: append([]byte(nil), secret...)}
}

func (m JWTManager) IssueAccess(userID int64, now time.Time) (string, time.Time, error) {
	expires := now.Add(15 * time.Minute)
	claims := jwt.RegisteredClaims{Subject: strconv.FormatInt(userID, 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expires)}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return raw, expires, err
}

func (m JWTManager) ParseAccess(raw string) (int64, error) {
	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid access token")
	}
	return strconv.ParseInt(token.Claims.(*jwt.RegisteredClaims).Subject, 10, 64)
}

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
