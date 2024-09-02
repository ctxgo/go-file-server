package avatar

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (api *AvatarAPI) Create(c *gin.Context) {
	file, err := c.FormFile("upload[]")
	if err != nil {
		core.ErrRep().SetMsg("Invalid file upload").SendGin(c)
		return
	}
	// Open uploaded file
	f, err := file.Open()
	if err != nil {
		core.ErrRep().SetMsg("Cannot open file").SendGin(c)
		return
	}
	defer f.Close()
	// Read file content
	data, err := io.ReadAll(f)
	if err != nil {
		core.ErrRep().SetMsg("Cannot read file").SendGin(c)
		return
	}
	err = api.create(c, data)
	if err != nil {
		c.Error(err)
	}
	core.OKRep(nil).SendGin(c)
}

func (api *AvatarAPI) create(c *gin.Context, data []byte) error {
	claims := core.ExtractClaims(c)

	avatarData := &models.Avatar{
		UserID: claims.UserId,
		Data:   data,
	}

	err := api.avatarRepo.Create(avatarData, core.WithClauses("user_id")("data"))

	return errors.WithStack(err)

}
