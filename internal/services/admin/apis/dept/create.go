package dept

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/config"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CreateRep struct {
	DeptId int `json:"dept_id" `
}

type CreateReq struct {
	ParentId *int   `json:"parentId" binding:"required"`          //上级部门
	DeptName string `json:"deptName"  binding:"required"`         //部门名称
	Sort     int    `json:"sort"`                                 //排序
	Leader   string `json:"leader"  binding:"required"`           //负责人
	Phone    string `json:"phone" `                               //手机
	Email    string `json:"email" `                               //邮箱
	Status   int    `json:"status"  binding:"required,oneof=1 2"` //状态
}

func (api *DeptApi) Create(c *gin.Context) {
	var createVals CreateReq
	err := c.ShouldBind(&createVals)
	if err != nil {
		c.Error(err)
		return
	}

	deptId, err := api.create(c, createVals)

	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(
		CreateRep{
			DeptId: deptId,
		},
	).SendGin(c)
}

func (api *DeptApi) create(c *gin.Context, createVals CreateReq) (int, error) {
	claims := core.ExtractClaims(c)

	sysDept := &models.SysDept{
		ParentId: createVals.ParentId,
		DeptName: createVals.DeptName,
		Sort:     createVals.Sort,
		Leader:   createVals.Leader,
		Phone:    createVals.Phone,
		Email:    createVals.Email,
		Status:   createVals.Status,
		ControlBy: coreModels.ControlBy{
			CreateBy: claims.UserId,
		},
	}

	err := api.deptRepo.Repo.WithTransaction(
		func(tx *gorm.DB) error {

			deptTsRepo := repository.NewDeptRepository(tx)

			if err := deptTsRepo.Create(sysDept); err != nil {
				return err
			}

			deptPath, err := api.makeDeptPath(sysDept)
			if err != nil {
				return err
			}
			return deptTsRepo.Update(func(sd *models.SysDept) {
				sd.DeptPath = deptPath
			}, repository.WithByDeptId(sysDept.DeptId))
		},
	)
	if err != nil {
		if repository.IsDuplicateError(config.DatabaseCfg.Driver, err) {
			return 0, core.NewApiBizErr(errors.WithStack(err)).
				SetBizCode(global.BizError).
				SetMsg("部门已经存在")
		}
		return 0, errors.WithStack(err)
	}
	return sysDept.DeptId, errors.WithStack(err)
}

func (api *DeptApi) makeDeptPath(sysDept *models.SysDept) (string, error) {
	deptPath := strconv.Itoa(sysDept.DeptId) + "/"
	if *sysDept.ParentId == 0 {
		return "/0/" + deptPath, nil
	}
	deptP, err := api.deptRepo.FindOne(repository.WithByDeptId(*sysDept.ParentId))
	if err != nil {
		return "", errors.WithStack(err)
	}
	return deptP.DeptPath + deptPath, nil

}

func (api *DeptApi) makeTopDeptPath(sysDept *models.SysDept) (string, error) {
	deptPath := strconv.Itoa(sysDept.DeptId) + "/"
	if *sysDept.ParentId == 0 {
		return "/0/" + deptPath, nil
	}
	deptP, err := api.deptRepo.FindOne(repository.WithByDeptId(*sysDept.ParentId))
	if err != nil {
		return "", errors.WithStack(err)
	}
	return deptP.DeptPath + deptPath, nil

}
