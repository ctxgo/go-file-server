package role

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/zlog"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type DeleteReq struct {
	Ids []int `json:"ids" binding:"required,min=1"`
}

type DeleteRep struct {
	Data []DeleteResult `json:"data"`
}

type DeleteResult struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}

func (api *RoleApi) Delete(c *gin.Context) {
	var deleteReq DeleteReq
	err := c.ShouldBind(&deleteReq)
	if err != nil {
		c.Error(err)
		return
	}
	results := api.delete(deleteReq)

	core.OKRep(
		DeleteRep{
			Data: results,
		}).SendGin(c)
}

func (api *RoleApi) delete(deleteReq DeleteReq) []DeleteResult {
	results := make([]DeleteResult, len(deleteReq.Ids))
	resultChan := make(chan DeleteResult, len(deleteReq.Ids)) // 创建结果管道
	defer close(resultChan)
	for _, id := range deleteReq.Ids {
		go func(id int) {
			resultChan <- api.deleteRole(id) // 将结果发送到管道
		}(id)
	}

	for i := 0; i < len(deleteReq.Ids); i++ {
		results[i] = <-resultChan
	}

	return results

}

func (api *RoleApi) deleteRole(id int) (result DeleteResult) {
	var err error
	var data *models.SysRole
	result.Id = id
	defer func() {
		result.Msg = "删除成功"
		if err != nil {
			zlog.SugLog.Error(err)
			result.Msg = "删除失败"
		}
	}()

	data, err = api.roleRepo.FindOne(
		repository.WithPreloadSysDept(),
		repository.WithRoleId(id),
	)
	if err != nil {
		return
	}

	err = api.roleRepo.CascadeDelete(data)
	if err != nil {
		return
	}
	err = api.deletePolicies(data.RoleKey)
	return
}

func (api *RoleApi) deletePolicies(roleKey string) error {
	_, err := api.casbinEnforcer.RemoveFilteredPolicy(0, roleKey)
	if err != nil {
		return errors.WithStack(err)
	}
	err = api.casbinEnforcer.LoadPolicy()
	return errors.WithStack(err)
}
