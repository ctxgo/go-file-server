package routers

import (
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/normal/apis/auth"
	"go-file-server/internal/services/normal/apis/captcha"
	"go-file-server/internal/services/normal/apis/config"
	"go-file-server/internal/services/normal/apis/root"

	"go.uber.org/fx"
)

var repos = fx.Options(
	fx.Provide(
		repository.NewLoginLogRepository,
		repository.NewUserRepository,
		repository.NewRoleRepository,
		repository.NewDeptRepository,
	),
)

var apis = fx.Options(
	fx.Provide(
		root.NewRootHandler,
		config.NewConfigHandler,
		auth.NewAuthenticatorApi,
		captcha.NewCaptchaAPI,
	),
)

var Routers = fx.Options(
	repos,
	apis,
	fx.Invoke(
		RegisterAuthRoutes,
		RegisterCaptchaRoutes,
		RegisterRootRoutes,
		RegisterConfigRoutes,
	),
)
