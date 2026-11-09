package dto

import (
	"errors"

	"github.com/codacube/go-microservice-banking/internal/auth/domain"
	"github.com/golang-jwt/jwt/v4"
)

type RefreshTokenRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r RefreshTokenRequest) IsAccessTokenValid() *jwt.ValidationError {

	// 1. Invalid token
	// 2. Valid token but expired
	_, err := jwt.Parse(r.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(domain.HMAC_SAMPLE_SECRET), nil
	})
	if err != nil {
		var vErr *jwt.ValidationError
		if errors.As(err, &vErr) {
			return vErr
		}
	}

	return nil
}
