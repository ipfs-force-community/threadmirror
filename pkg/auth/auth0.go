package auth

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type JWTVerifierAuth0Impl struct {
	auth0Domain   string
	auth0Audience string
	validator     *validator.Validator
}

func NewJWTVerifierAuth0Impl(auth0Domain string, auth0Audience string) (*JWTVerifierAuth0Impl, error) {
	issuerURL, err := url.Parse("https://" + auth0Domain + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %w", err)
	}
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{auth0Audience},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}
	return &JWTVerifierAuth0Impl{
		auth0Domain:   auth0Domain,
		auth0Audience: auth0Audience,
		validator:     jwtValidator,
	}, nil
}

func (a *JWTVerifierAuth0Impl) Verify(accesssToken string) (*AuthInfo, error) {
	token, err := a.validator.ValidateToken(context.Background(), accesssToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	// todo: parse token
	fmt.Println(token)
	return &AuthInfo{}, nil
}
