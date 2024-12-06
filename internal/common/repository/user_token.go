package repository

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"gorm.io/gorm"
)

type UserTokenRepository struct {
	Repo *core.Repo
}

func NewUserTokenRepository(db *gorm.DB) *UserTokenRepository {
	return &UserTokenRepository{Repo: core.NewRepo(db)}
}

func (r *UserTokenRepository) Create(values *models.UserToken, opts ...base.DbScope) error {
	return r.Repo.Create(values, opts...)
}
func (r *UserTokenRepository) Save(values *models.UserToken) error {
	return r.Repo.Save(values)
}

func (r *UserTokenRepository) Delete(opts ...base.DbScope) error {
	return r.Repo.Delete(&models.UserToken{}, opts...)
}

func (r *UserTokenRepository) Update(updateFunc func(*models.UserToken), opts ...base.DbScope) error {
	data := &models.UserToken{}
	updateFunc(data)
	return r.Repo.Update(data)
}

func (r *UserTokenRepository) FindOne(opts ...base.DbScope) (data *models.UserToken, err error) {
	err = r.Repo.FindOne(&data, opts...)
	return
}

func WithRevoked(b bool) base.DbScope {
	return base.WithQuery("revoked = ?", b)
}

func WithUserTokenIds(ids ...int) base.DbScope {
	return base.WithQuery("id in ?", ids)
}

func WithUserTokenUserId(id int) base.DbScope {
	return base.WithQuery("user_id = ?", id)
}

func WithUserToken(t string) base.DbScope {
	return base.WithQuery("token = ?", t)
}

func (r *UserTokenRepository) Find(opts ...base.DbScope) (data []models.UserToken, err error) {
	err = r.Repo.Find(&data, opts...)
	return
}
