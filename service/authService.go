package service

import (
	"fmt"

	"github.com/codacube/banking-auth/domain"
	"github.com/codacube/banking-auth/dto"
	"github.com/codacube/banking-lib/errs"
	"github.com/codacube/banking-lib/logger"
	"github.com/dgrijalva/jwt-go"
)

type AuthService interface {
	Login(dto.LoginRequest) (*dto.LoginResponse, *errs.AppError)
	Verify(urlParams map[string]string) *errs.AppError
	Refresh(request dto.RefreshTokenRequest) (*dto.LoginResponse, *errs.AppError)
}

type DefaultAuthService struct {
	repo            domain.AuthRepository
	rolePermissions domain.RolePermissions
}

func (s DefaultAuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, *errs.AppError) {
	var appErr *errs.AppError
	var login *domain.Login

	if login, appErr = s.repo.FindBy(req.Username, req.Password); appErr != nil {
		return nil, appErr
	}

	claims := login.ClaimsForAccessToken()
	authToken := domain.NewAuthToken(claims)

	var accessToken, refreshToken string
	if accessToken, appErr = authToken.NewAccessToken(); appErr != nil {
		return nil, appErr
	}
	if refreshToken, appErr = s.repo.GenerateAndSaveRefreshTokenToStore(authToken); appErr != nil {
		return nil, appErr
	}

	return &dto.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s DefaultAuthService) Verify(urlParams map[string]string) *errs.AppError {
	// Convert the string token to JWT struct
	jwtToken, err := jwtTokenFromString(urlParams["token"])
	if err != nil {
		return errs.NewAuthorisationError(err.Error())
	}

	/*
		Checking the validity of the token, this verifies the expiry time and
		the signature of the token
	*/
	if !jwtToken.Valid {
		return errs.NewAuthorisationError("invalid token")
	}

	// Cast the token claims to jwt.MapClaims
	claims := jwtToken.Claims.(*domain.AccessTokenClaims)
	/* If Role is user then check if the account_id and customer_id coming
	   in the URL belongs to the same token
	*/
	if claims.IsUserRole() && !claims.IsRequestVerifiedWithTokenClaims(urlParams) {
		return errs.NewAuthorisationError("request not verified with the token claims")
	}

	// Verify if the role is authorised to use the route
	if !s.rolePermissions.IsAuthorisedFor(claims.Role, urlParams["routeName"]) {
		return errs.NewAuthorisationError(fmt.Sprintf("%s role is not authorised", claims.Role))
	}

	// Everythink is ok
	return nil
}

func (s DefaultAuthService) Refresh(request dto.RefreshTokenRequest) (*dto.LoginResponse, *errs.AppError) {
	if vErr := request.IsAccessTokenValid(); vErr != nil {
		if vErr.Errors == jwt.ValidationErrorExpired {
			// continue with the refresh token functionality
			var appErr *errs.AppError
			if appErr = s.repo.RefreshTokenExists(request.RefreshToken); appErr != nil {
				return nil, appErr
			}
			// generate a access token from refresh token.
			var accessToken string
			if accessToken, appErr = domain.NewAccessTokenFromRefreshToken(request.RefreshToken); appErr != nil {
				return nil, appErr
			}
			return &dto.LoginResponse{AccessToken: accessToken}, nil
		}
		return nil, errs.NewAuthenticationError("invalid token")
	}
	return nil, errs.NewAuthenticationError("cannot generate a new access token until the current one expires")
}

func jwtTokenFromString(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(domain.HMAC_SAMPLE_SECRET), nil
	})
	if err != nil {
		logger.Error("Error while parsing token: " + err.Error())
		return nil, err
	}
	return token, nil
}

func NewLoginService(repo domain.AuthRepository, permissions domain.RolePermissions) DefaultAuthService {
	return DefaultAuthService{repo, permissions}
}
