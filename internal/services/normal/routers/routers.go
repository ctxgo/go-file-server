package routers

import (
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/normal/apis/auth"
	"go-file-server/internal/services/normal/apis/root"

	"go.uber.org/fx"
)

var repos = fx.Options(
	fx.Provide(
		repository.NewLoginLogRepository,
		repository.NewUserRepository,
		repository.NewRoleRepository,
	),
)

var apis = fx.Options(
	fx.Provide(
		root.NewRootHandler,
		auth.NewAuthenticatorApi,
		auth.NewCaptchaAPI,
	),
)

var Routers = fx.Options(
	repos,
	apis,
	fx.Invoke(
		RegisterAuthRoutes,
		RegisterRootRoutes,
	),
)
