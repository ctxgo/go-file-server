package routers

import (
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/apis/avatar"
	"go-file-server/internal/services/admin/apis/dept"
	"go-file-server/internal/services/admin/apis/fs"
	"go-file-server/internal/services/admin/apis/log/login"
	"go-file-server/internal/services/admin/apis/log/opera"
	"go-file-server/internal/services/admin/apis/menu"
	"go-file-server/internal/services/admin/apis/role"
	"go-file-server/internal/services/admin/apis/system"
	"go-file-server/internal/services/admin/apis/user"

	"go.uber.org/fx"
)

var repos = fx.Options(
	fx.Provide(
		repository.NewLoginLogRepository,
		repository.NewUserRepository,
		repository.NewUserTokenRepository,
		repository.NewRoleRepository,
		repository.NewOperaLogRepository,
		repository.NewDeptRepository,
		repository.NewMenuRepository,
		repository.NewAvatarRepository,
		repository.NewFsRepository,
	),
)

var apis = fx.Options(
	fx.Provide(
		dept.NewDeptApi,
		login.NewLogApi,
		opera.NewOperaAPI,
		role.NewRoleApi,
		user.NewUserAPI,
		avatar.NewAvatarAPI,
		menu.NewRoleApi,
		fs.NewFsApi,
		system.NewSystemApi,
	),
)

var Routers = fx.Options(
	repos,
	apis,
	fx.Invoke(
		RegisterDeptRoutes,
		RegisterUserRoutes,
		RegisterRoleRoutes,
		RegisterLogRoutes,
		RegisterAvatarRoutes,
		RegisterMenuRoutes,
		RegisterFsRoutes,
		RegisterSystemRoutes,
	),
)
