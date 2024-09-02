package core

import (
	"errors"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/types"
	"go-file-server/pkgs/zlog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type HttpErr interface {
	Error() string
	GetHttpCode() global.HttpCode
	GetBizCode() global.BizCode
}

type ApiErr struct {
	types.HttpRep
	rawErr error
}

func NewApiErr(e error) *ApiErr {
	r := ErrRep().
		SetBizCode(global.BizError).
		SetHttpCode(global.InternalServerError)
	return &ApiErr{
		HttpRep: r,
		rawErr:  e,
	}
}

func NewApiBizErr(e error) *ApiErr {
	return NewApiErr(e).SetHttpCode(global.HttpSuccess)
}

func (e *ApiErr) SetBizCode(c global.BizCode) *ApiErr {
	e.HttpRep.SetBizCode(c)
	return e
}

func (e *ApiErr) SetHttpCode(c global.HttpCode) *ApiErr {
	e.HttpRep.SetHttpCode(c)
	return e
}

func (e *ApiErr) SetMsg(s string) *ApiErr {
	e.HttpRep.SetMsg(s)
	return e
}

func (e *ApiErr) Error() string {
	return e.HttpRep.GetMsg()
}

func (e *ApiErr) GetRawErr() error {
	return e.rawErr
}

type ValidationErrors validator.ValidationErrors

func (e ValidationErrors) SetGinErr(c *gin.Context) {
	c.Error(e)
}

func (e ValidationErrors) GetHttpCode() global.HttpCode {
	return global.BadRequestError
}

func (e ValidationErrors) GetBizCode() global.BizCode {
	return global.BizBadRequest
}

func (e ValidationErrors) Error() string {
	var errMsgs []string
	for _, valErr := range e {
		fieldPath := strings.Split(valErr.Namespace(), ".")
		fieldName := fieldPath[len(fieldPath)-1]
		errMsgs = append(errMsgs, fieldName+" is required")
	}
	return strings.Join(errMsgs, ", ")
}

func NewMinRep() *types.MinRep {
	return &types.MinRep{}
}

func NewErrRep() *types.ErrRep {
	return &types.ErrRep{
		MinRep: NewMinRep(),
	}
}

func NewRep(data any) *types.Rep {
	return &types.Rep{
		MinRep: NewMinRep(),
		Data:   data,
	}
}

func SendHttpErrRep(c *gin.Context, e HttpErr) {
	ErrRep().
		SetHttpCode(e.GetHttpCode()).
		SetBizCode(e.GetBizCode()).
		SetMsg(e.Error()).
		SendGin(c)
}

func ErrRep() types.HttpRep {
	return NewErrRep().
		SetHttpCode(global.InternalServerError).
		SetBizCode(global.BizError)
}

func ErrBizRep() types.HttpRep {
	return ErrRep().
		SetHttpCode(global.HttpSuccess)
}

func OKRep(data any) types.HttpRep {
	return NewRep(data).
		SetHttpCode(global.HttpSuccess).
		SetBizCode(global.BizSuccess)
}

func HandlingErr(c *gin.Context, err error) {

	switch e := err.(type) {
	case validator.ValidationErrors:
		SendHttpErrRep(c, ValidationErrors(e))
	case *ApiErr:
		zlog.SugLog.Errorf("%s, err: %+v", e.GetMsg(), e.GetRawErr())
		SendHttpErrRep(c, e)
	case *SseErr:
		zlog.SugLog.Errorf("%s, err: %+v", e.GetMsg(), e.GetRawErr())
		OnceStream(c, "error", e.GetMsg())
	default:
		zlog.SugLog.Errorf("%s, err: %+v", "服务异常", e)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ErrBizRep().SetBizCode(global.BizNotFound).SendGin(c)
			return
		}
		SendHttpErrRep(c, NewApiErr(e))
	}

}
