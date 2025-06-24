package authfx

import (
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
