package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Manager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewManager(secret string, ttlHours int) *Manager {
	return &Manager{
		secret: []byte(secret),
		ttl:    time.Duration(ttlHours) * time.Hour,
		issuer: "banking-service",
	}
}

func (m *Manager) Issue(userID string, role string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (m *Manager) Parse(tokenValue string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenValue, &Claims{}, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
