package authfx

import (
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"go.uber.org/fx"
)

func Module(supabaseJWTKey []byte) fx.Option {
	return fx.Module("auth",
		fx.Provide(func() (auth.JWTVerifier, error) {
			return &auth.JWTVerifierImpl{
				JWTKey: supabaseJWTKey,
			}, nil
		}),
	)
}

func ModuleAuth0(auth0Config *config.Auth0Config) fx.Option {
	return fx.Module("auth",
		fx.Provide(func() (auth.JWTVerifier, error) {
			return auth.NewJWTVerifierAuth0Impl(auth0Config)
		}),
	)
}
