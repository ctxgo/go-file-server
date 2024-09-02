package opera

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	Ids []int `json:"ids" binding:"required,min=1"`
}

type DeleteRep struct {
	DeleteReq
}

func (api *OperaAPI) Delete(c *gin.Context) {
	var deleteReq DeleteReq
	err := c.ShouldBind(&deleteReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.operarepo.Delete(repository.WithOperaIds(deleteReq.Ids...))
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(DeleteRep{
		DeleteReq: deleteReq,
	}).SendGin(c)
}
