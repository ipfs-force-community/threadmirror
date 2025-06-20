package servicefx

import (
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(fx.Annotate(sqlrepo.NewUserRepo, fx.As(new(service.UserRepoInterface)))),
	fx.Provide(service.NewUserService),
)
