package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	supabaseauth "github.com/supabase-community/auth-go"
)

type AuthClient = supabaseauth.Client

type AuthInfo struct {
	UserID       string
	Email        string
	Phone        string
	AccesssToken string
	IssuedAt     time.Time
	ExpiresAt    time.Time
	SessionID    string
	IsAnonymous  bool
	AppMetadata  map[string]any
	UserMetadata map[string]any
}

type JWTVerifier interface {
	Verify(accesssToken string) (*AuthInfo, error)
}

type JWTVerifierImpl struct {
	JWTKey []byte
}

func (a *JWTVerifierImpl) Verify(accesssToken string) (*AuthInfo, error) {
	token, err := jwt.ParseWithClaims(
		accesssToken,
		&JWTClaims{},
		func(token *jwt.Token) (any, error) {
			return a.JWTKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Convert timestamps
	issuedAt := time.Unix(claims.Iat, 0)
	expiresAt := time.Unix(claims.Exp, 0)

	return &AuthInfo{
		UserID:       claims.Sub,
		Email:        claims.Email,
		Phone:        claims.Phone,
		AccesssToken: accesssToken,
		IssuedAt:     issuedAt,
		ExpiresAt:    expiresAt,
		SessionID:    claims.SessionID,
		IsAnonymous:  claims.IsAnonymous,
		AppMetadata:  claims.AppMetadata,
		UserMetadata: claims.UserMetadata,
	}, nil
}

// JWTClaims represents the Supabase JWT token claims structure
type JWTClaims struct {
	Exp          int64          `json:"exp"`           // Expiration time
	Iat          int64          `json:"iat"`           // Issued at time
	Sub          string         `json:"sub"`           // Subject (User ID)
	Aud          string         `json:"aud"`           // Audience
	Iss          string         `json:"iss"`           // Issuer
	Email        string         `json:"email"`         // User email
	Phone        string         `json:"phone"`         // User phone
	AppMetadata  map[string]any `json:"app_metadata"`  // Application metadata
	UserMetadata map[string]any `json:"user_metadata"` // User metadata
	Role         string         `json:"role"`          // User role
	Aal          string         `json:"aal"`           // Authentication Assurance Level
	Amr          []struct {
		Method    string `json:"method"`
		Timestamp int64  `json:"timestamp"`
	} `json:"amr"` // Authentication Methods References
	SessionID   string `json:"session_id"`   // Session identifier
	IsAnonymous bool   `json:"is_anonymous"` // Whether user is anonymous
}

// Implement jwt.Claims interface methods
func (c JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.Exp == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Exp, 0)), nil
}

func (c JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.Iat == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Iat, 0)), nil
}

func (c JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c JWTClaims) GetIssuer() (string, error) {
	return c.Iss, nil
}

func (c JWTClaims) GetSubject() (string, error) {
	return c.Sub, nil
}

func (c JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	if c.Aud == "" {
		return nil, nil
	}
	return jwt.ClaimStrings{c.Aud}, nil
}
