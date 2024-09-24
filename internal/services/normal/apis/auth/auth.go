package auth

import "go-file-server/internal/common/repository"

type Authenticator struct {
	deptRepo     *repository.DeptRepository
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	loginLogRepo *repository.LoginLogRepository
}

func NewAuthenticatorApi(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	loginLogRepo *repository.LoginLogRepository,
	deptRepo *repository.DeptRepository,
) *Authenticator {
	return &Authenticator{
		deptRepo:     deptRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		loginLogRepo: loginLogRepo,
	}
}
