package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type AvatarRepository struct {
	Repo *core.Repo
}

func NewAvatarRepository(db *gorm.DB) *AvatarRepository {
	return &AvatarRepository{Repo: core.NewRepo(db)}
}

func (r *AvatarRepository) Create(values *models.Avatar, opts ...base.DbScope) error {
	return r.Repo.Create(values, opts...)
}

func (r *AvatarRepository) Save(values *models.Avatar) error {
	return r.Repo.Save(values)
}

func (r *AvatarRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.Avatar{}, opts...)
}

func (r *AvatarRepository) Update(updateFunc func(*models.Avatar), opts ...base.DbScope) error {
	avatar := &models.Avatar{}
	updateFunc(avatar)
	return r.Repo.Update(avatar)
}

func (r *AvatarRepository) FindOne(opts ...base.DbScope) (avatar *models.Avatar, err error) {
	err = r.Repo.FindOne(&avatar, opts...)
	return
}
