package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 响应结构。
// 业务状态码 code 与 HTTP 状态码解耦：无论成功失败，HTTP 统一返回 200，
// 由 code 表达业务结果，便于前端统一拦截。
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 常用业务状态码
const (
	CodeSuccess      = 200
	CodeError        = 500
	CodeInvalidParam = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	// CodeInsufficientBalance 下单时可用余额不足以支付应付金额
	CodeInsufficientBalance = 402
)

// Success 成功并返回数据
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: "success", Data: data})
}

// SuccessWithMsg 成功并返回自定义消息
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: msg, Data: data})
}

// Page 返回分页数据
func Page(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: gin.H{"list": list, "total": total},
	})
}

// Fail 业务失败（HTTP 仍为 200，由 code 区分）
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: nil})
}

// FailWithData 业务失败并返回附加数据
func FailWithData(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: data})
}

// Error 系统级错误，使用对应 HTTP 状态码
func Error(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{Code: httpCode, Msg: msg, Data: nil})
}
