package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/owlto-dao/utils-go/errors"
	"github.com/owlto-dao/utils-go/response"
	"github.com/owlto-dao/utils-go/util"

	"strings"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserAddressKey      = "user_address"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.RespondError(c, errors.UnauthorizedErr)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)
		if token == "" {
			response.RespondError(c, errors.UnauthorizedErr)
			c.Abort()
			return
		}

		valid, err := util.ValidateToken(token)
		if err != nil || !valid {
			response.RespondError(c, errors.UnauthorizedErr)
			c.Abort()
			return
		}

		address, err := util.GetAddressFromToken(token)
		if err != nil {
			response.RespondError(c, errors.UnauthorizedErr)
			c.Abort()
			return
		}

		c.Set(UserAddressKey, address)

		c.Next()
	}
}

func GetUserAddress(c *gin.Context) (string, bool) {
	address, exists := c.Get(UserAddressKey)
	if !exists {
		return "", false
	}

	addressStr, ok := address.(string)
	return addressStr, ok
}
