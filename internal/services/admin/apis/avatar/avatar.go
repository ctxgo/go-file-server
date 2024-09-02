package avatar

import "go-file-server/internal/common/repository"

type AvatarAPI struct {
	avatarRepo *repository.AvatarRepository
}

func NewAvatarAPI(
	avatarRepo *repository.AvatarRepository,

) *AvatarAPI {
	return &AvatarAPI{
		avatarRepo: avatarRepo,
	}
}
