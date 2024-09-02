package global

import (
	"net/http"
)

type HttpCode int //通用http类型code

// http code
const (
	HttpSuccess         HttpCode = http.StatusOK                  //成功
	UnauthorizedError   HttpCode = http.StatusUnauthorized        //缺少有效认证，通常用于认证失败
	InternalServerError HttpCode = http.StatusInternalServerError //服务器遇到错误，不能完成请求
	BadRequestError     HttpCode = http.StatusBadRequest          //请求格式错误或不能被处理
	StatusNotFound      HttpCode = http.StatusNotFound            //请求的资源未找到
	ForbiddenError      HttpCode = http.StatusForbidden           //请求被禁止
	ConflictError       HttpCode = http.StatusConflict            // 数据或状态冲突
)

type BizCode int // 业务code

// 业务code
const (
	BizSuccess           BizCode = iota + 1000 // 1000, 成功
	BizError                                   // 1001, 失败
	BizBadRequest                              // 1002, 请求格式错误或不能被处理
	BizAccessDenied                            // 1003, 无操作权限
	BizUnauthorizedErr                         // 1004, 无效的登录凭证
	BizTokenExpiredErr                         // 1005, token过期
	BizDataInvalid                             // 1006, 数据不符合要求
	BizOperationFailed                         // 1007, 操作未能成功执行
	BizRateLimitExceeded                       // 1008, 超出了频率限制
	BizNotFound                                // 1009, 资源未找到
)

var MessageMap map[BizCode]string = map[BizCode]string{
	BizSuccess:           "成功",
	BizError:             "失败",
	BizBadRequest:        "请求格式错误或不能被处理",
	BizUnauthorizedErr:   "无效的登录凭证",
	BizTokenExpiredErr:   "无效的登录凭证",
	BizAccessDenied:      "无操作权限",
	BizDataInvalid:       "数据不符合要求",
	BizOperationFailed:   "操作未能成功执行",
	BizRateLimitExceeded: "超出了频率限制",
	BizNotFound:          "资源未找到",
}
