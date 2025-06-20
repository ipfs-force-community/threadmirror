package testsuit

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"github.com/stretchr/testify/mock"
	supabase_auth "github.com/supabase-community/auth-go"
	"github.com/supabase-community/auth-go/types"
	"gorm.io/datatypes"
)

// MockAuthClient is a mock implementation of auth.Client for testing
// It only implements the methods we actually use in our auth handlers
type MockAuthClient struct {
	mock.Mock
}

// Implement only the methods we use in our auth handlers
func (m *MockAuthClient) Signup(req types.SignupRequest) (*types.SignupResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.SignupResponse), args.Error(1)
}

func (m *MockAuthClient) SignInWithEmailPassword(
	email, password string,
) (*types.TokenResponse, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) SignInWithPhonePassword(
	phone, password string,
) (*types.TokenResponse, error) {
	args := m.Called(phone, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) Authorize(req types.AuthorizeRequest) (*types.AuthorizeResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.AuthorizeResponse), args.Error(1)
}

func (m *MockAuthClient) VerifyForUser(
	req types.VerifyForUserRequest,
) (*types.VerifyForUserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.VerifyForUserResponse), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(refreshToken string) (*types.TokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) WithToken(token string) supabase_auth.Client {
	args := m.Called(token)
	return args.Get(0).(supabase_auth.Client)
}

func (m *MockAuthClient) Logout() error {
	args := m.Called()
	return args.Error(0)
}

// Stub implementations for all other required interface methods
func (m *MockAuthClient) WithCustomAuthURL(url string) supabase_auth.Client  { return m }
func (m *MockAuthClient) WithClient(client http.Client) supabase_auth.Client { return m }

func (m *MockAuthClient) AdminAudit(
	req types.AdminAuditRequest,
) (*types.AdminAuditResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminGenerateLink(
	req types.AdminGenerateLinkRequest,
) (*types.AdminGenerateLinkResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminListSSOProviders() (*types.AdminListSSOProvidersResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminCreateSSOProvider(
	req types.AdminCreateSSOProviderRequest,
) (*types.AdminCreateSSOProviderResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminGetSSOProvider(
	req types.AdminGetSSOProviderRequest,
) (*types.AdminGetSSOProviderResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminUpdateSSOProvider(
	req types.AdminUpdateSSOProviderRequest,
) (*types.AdminUpdateSSOProviderResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminDeleteSSOProvider(
	req types.AdminDeleteSSOProviderRequest,
) (*types.AdminDeleteSSOProviderResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminCreateUser(
	req types.AdminCreateUserRequest,
) (*types.AdminCreateUserResponse, error) {
	return nil, nil
}
func (m *MockAuthClient) AdminListUsers() (*types.AdminListUsersResponse, error) { return nil, nil }

func (m *MockAuthClient) AdminGetUser(
	req types.AdminGetUserRequest,
) (*types.AdminGetUserResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminUpdateUser(
	req types.AdminUpdateUserRequest,
) (*types.AdminUpdateUserResponse, error) {
	return nil, nil
}
func (m *MockAuthClient) AdminDeleteUser(req types.AdminDeleteUserRequest) error { return nil }

func (m *MockAuthClient) AdminListUserFactors(
	req types.AdminListUserFactorsRequest,
) (*types.AdminListUserFactorsResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminUpdateUserFactor(
	req types.AdminUpdateUserFactorRequest,
) (*types.AdminUpdateUserFactorResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) AdminDeleteUserFactor(req types.AdminDeleteUserFactorRequest) error {
	return nil
}

func (m *MockAuthClient) EnrollFactor(
	req types.EnrollFactorRequest,
) (*types.EnrollFactorResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) ChallengeFactor(
	req types.ChallengeFactorRequest,
) (*types.ChallengeFactorResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) VerifyFactor(
	req types.VerifyFactorRequest,
) (*types.VerifyFactorResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) UnenrollFactor(
	req types.UnenrollFactorRequest,
) (*types.UnenrollFactorResponse, error) {
	return nil, nil
}
func (m *MockAuthClient) HealthCheck() (*types.HealthCheckResponse, error) { return nil, nil }
func (m *MockAuthClient) Invite(req types.InviteRequest) (*types.InviteResponse, error) {
	return nil, nil
}
func (m *MockAuthClient) Magiclink(req types.MagiclinkRequest) error { return nil }
func (m *MockAuthClient) OTP(req types.OTPRequest) error             { return nil }
func (m *MockAuthClient) Reauthenticate() error                      { return nil }
func (m *MockAuthClient) Recover(req types.RecoverRequest) error     { return nil }

func (m *MockAuthClient) GetSettings() (*types.SettingsResponse, error) { return nil, nil }

func (m *MockAuthClient) Token(
	req types.TokenRequest,
) (*types.TokenResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) GetUser() (*types.UserResponse, error) { return nil, nil }

func (m *MockAuthClient) UpdateUser(
	req types.UpdateUserRequest,
) (*types.UpdateUserResponse, error) {
	return nil, nil
}

func (m *MockAuthClient) Verify(req types.VerifyRequest) (*types.VerifyResponse, error) {
	return nil, nil
}
func (m *MockAuthClient) SAMLMetadata() ([]byte, error)                        { return nil, nil }
func (m *MockAuthClient) SAMLACS(req *http.Request) (*http.Response, error)    { return nil, nil }
func (m *MockAuthClient) SSO(req types.SSORequest) (*types.SSOResponse, error) { return nil, nil }

// SetTestAuthInfo sets authentication info in gin context for testing purposes
func SetTestAuthInfo(c *gin.Context, userID datatypes.UUID) {
	SetTestAuthInfoWithEmail(c, userID, "test@example.com")
}

// SetTestAuthInfoWithEmail sets authentication info in gin context with custom email for testing purposes
func SetTestAuthInfoWithEmail(c *gin.Context, userID datatypes.UUID, email string) {
	authInfo := &auth.AuthInfo{
		UserID:    userID,
		Email:     email,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	auth.SetAuthInfo(c, authInfo)
}
