package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"

	"github.com/ipfs-force-community/threadmirror/internal/config"
)

type JWTVerifierAuth0Impl struct {
	validator *validator.Validator
}

func NewJWTVerifierAuth0Impl(auth0Config *config.Auth0Config) (*JWTVerifierAuth0Impl, error) {
	issuerURL, err := url.Parse("https://" + auth0Config.Domain + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %w", err)
	}
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{auth0Config.Audience},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}
	return &JWTVerifierAuth0Impl{
		validator: jwtValidator,
	}, nil
}

func (a *JWTVerifierAuth0Impl) Verify(accesssToken string) (*AuthInfo, error) {
	token, err := a.validator.ValidateToken(context.Background(), accesssToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	claims, ok := token.(*validator.ValidatedClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast to validated claims")
	}
	userIdStr := strings.ReplaceAll(claims.RegisteredClaims.Subject, "twitter|", "")
	return &AuthInfo{
		UserID:       userIdStr,
		IssuedAt:     time.Unix(claims.RegisteredClaims.IssuedAt, 0),
		ExpiresAt:    time.Unix(claims.RegisteredClaims.Expiry, 0),
		AccesssToken: accesssToken,
	}, nil
}
