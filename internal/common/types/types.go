package types

import (
	"go-file-server/internal/common/global"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/pathtool"

	"github.com/casbin/casbin/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SvcCtx struct {
	Router         gin.IRouter
	Db             *gorm.DB
	FsIndexer      *pathtool.FileIndexer
	CasbinEnforcer *casbin.CachedEnforcer
	Cache          cache.AdapterCache
}

func (ctx *SvcCtx) Clone() *SvcCtx {
	newCtx := *ctx
	return &newCtx
}

type JwtClaims struct {
	UserId          int    `json:"user_id"`
	Username        string `json:"user_name"`
	RoleId          int    `json:"role_id"`
	RoleKey         string `json:"role_key"`
	IsPersonalToken bool   `json:"is_personal_token"`
	RoleName        string `json:"role_name"`
	DataScope       string `json:"data_scope"`
	jwt.StandardClaims
}

type Pagination struct {
	PageIndex int `form:"pageIndex"`
	PageSize  int `form:"pageSize"`
}

type Page struct {
	Count     int64 `json:"count"`
	PageIndex int   `json:"pageIndex"`
	PageSize  int   `json:"pageSize"`
}

func NewPage(Count int64, PageIndex, PageSize int) Page {
	return Page{
		Count:     Count,
		PageIndex: PageIndex,
		PageSize:  PageSize,
	}
}

type MinRep struct {
	HttpCode global.HttpCode `json:"-"`
	BizCode  global.BizCode  `json:"code"`
	Message  string          `json:"msg"`
}

func (r *MinRep) SetMsg(s string) {
	r.Message = s
}

func (r *MinRep) GetMsg() string {
	if r.Message == "" {
		return global.MessageMap[r.GetBizCode()]
	}
	return r.Message
}

func (r *MinRep) SetHttpCode(c global.HttpCode) {
	r.HttpCode = c
}

func (r *MinRep) GetHttpCode() global.HttpCode {
	return r.HttpCode
}

func (r *MinRep) SetBizCode(c global.BizCode) {
	r.BizCode = c
}

func (r *MinRep) GetBizCode() global.BizCode {
	return r.BizCode
}

type HttpRep interface {
	SetMsg(s string) HttpRep
	GetMsg() string
	SetHttpCode(c global.HttpCode) HttpRep
	GetHttpCode() global.HttpCode
	SetBizCode(c global.BizCode) HttpRep
	GetBizCode() global.BizCode
	SendGin(c *gin.Context)
}

type ErrRep struct {
	*MinRep
}

func (r *ErrRep) SetMsg(s string) HttpRep {
	r.MinRep.SetMsg(s)
	return r
}

func (r *ErrRep) SetHttpCode(c global.HttpCode) HttpRep {
	r.MinRep.SetHttpCode(c)
	return r
}

func (r *ErrRep) SetBizCode(c global.BizCode) HttpRep {
	r.MinRep.SetBizCode(c)
	return r
}

func (r *ErrRep) SendGin(c *gin.Context) {
	r.SetMsg(r.GetMsg())
	c.AbortWithStatusJSON(int(r.GetHttpCode()), r)
}

type Rep struct {
	*MinRep
	Data interface{} `json:"data,omitempty"`
}

func (r *Rep) SetMsg(s string) HttpRep {
	r.MinRep.SetMsg(s)
	return r
}

func (r *Rep) SetHttpCode(c global.HttpCode) HttpRep {
	r.MinRep.SetHttpCode(c)
	return r
}

func (r *Rep) SetBizCode(c global.BizCode) HttpRep {
	r.MinRep.SetBizCode(c)
	return r
}

func (r *Rep) SendGin(c *gin.Context) {
	r.SetMsg(r.GetMsg())
	c.AbortWithStatusJSON(int(r.GetHttpCode()), r)

}
