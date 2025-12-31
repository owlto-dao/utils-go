package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/owlto-dao/utils-go/errors"
)

type Response struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondJSON(c *gin.Context, code int64, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func RespondError(c *gin.Context, err error) {
	if err == nil {
		RespondOK(c, nil)
		return
	}
	if bizErr, ok := err.(*errors.BizError); ok {
		RespondJSON(c, bizErr.Code, bizErr.Msg, nil)
	} else {
		RespondJSON(c, 1001, err.Error(), nil)
	}
}

func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
