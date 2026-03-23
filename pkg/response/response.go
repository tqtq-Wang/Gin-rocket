package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func JSON(c *gin.Context, httpStatus int, message string, data any) {
	code := 0
	if httpStatus >= 400 {
		code = httpStatus
	}

	c.JSON(httpStatus, Body{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Success(c *gin.Context, httpStatus int, data any) {
	JSON(c, httpStatus, "ok", data)
}

func Fail(c *gin.Context, httpStatus int, message string) {
	JSON(c, httpStatus, message, nil)
}
