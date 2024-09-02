package login

import "go-file-server/internal/common/repository"

type LoginAPI struct {
	loginrepo *repository.LoginLogRepository
}

func NewLogApi(
	loginrepo *repository.LoginLogRepository,

) *LoginAPI {
	return &LoginAPI{
		loginrepo: loginrepo,
	}
}
