package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/config"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type CreateRep struct {
	UserId int `json:"userId" comment:"用户ID"`
}

type CreateReq struct {
	Username string `json:"username" comment:"用户名" vd:"len($)>0"`
	Password string `json:"password" comment:"密码"`
	NickName string `json:"nickName" comment:"昵称" vd:"len($)>0"`
	Phone    string `json:"phone" comment:"手机号" vd:"len($)>0"`
	RoleId   int    `json:"roleId" comment:"角色ID"`
	Avatar   string `json:"avatar" comment:"头像"`
	Sex      string `json:"sex" comment:"性别"`
	Email    string `json:"email" comment:"邮箱" vd:"len($)>0,email"`
	DeptId   int    `json:"deptId" comment:"部门" vd:"$>0"`
	PostId   int    `json:"postId" comment:"岗位"`
	Remark   string `json:"remark" comment:"备注"`
	Status   string `json:"status" comment:"状态" vd:"len($)>0" default:"1"`
}

func (api *UserAPI) Create(c *gin.Context) {

	var createReq CreateReq
	err := c.ShouldBind(&createReq)
	if err != nil {
		c.Error(err)
		return
	}
	id, err := api.create(c, createReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(CreateRep{
		UserId: id,
	}).SendGin(c)

}

func (api *UserAPI) create(c *gin.Context, createReq CreateReq) (int, error) {
	if createReq.Password == "" {
		createReq.Password = "123456"
	}
	data := &models.SysUser{
		Username: createReq.Username,
		Password: createReq.Password,
		NickName: createReq.NickName,
		Phone:    createReq.Phone,
		RoleId:   createReq.RoleId,
		Sex:      createReq.Sex,
		Email:    createReq.Email,
		Status:   createReq.Status,
		DeptId:   createReq.DeptId,
		ControlBy: coreModels.ControlBy{
			CreateBy: core.GetUserId(c),
		},
	}

	err := api.userRepo.Repo.Create(data)

	if err == nil {
		return 0, errors.WithStack(err)
	}

	if repository.IsDuplicateError(config.DatabaseCfg.Driver, err) {
		return 0, core.
			NewApiBizErr(errors.WithStack(err)).
			SetMsg("用户已存在")
	}
	return data.UserId, errors.Errorf(global.ErrServerNotOK)
}
