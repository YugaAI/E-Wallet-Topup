package cmd

import (
	"ewallet-topup/helpers"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (d *Dependency) MiddlewareValidateToken(c *gin.Context) {

	var (
		log = helpers.Logger
	)
	auth := c.GetHeader("Authorization")
	if auth == "" {
		log.Warn("Authorization header is empty")
		helpers.SendResponseHTTP(c, http.StatusUnauthorized, "unauthorized", nil)
		c.Abort()
		return
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		log.Warn("invalid authorization header format")
		helpers.SendResponseHTTP(c, http.StatusUnauthorized, "unauthorized", nil)
		c.Abort()
		return
	}

	token := strings.TrimPrefix(auth, prefix)

	tokenData, err := d.External.ValidateToken(c.Request.Context(), token)
	if err != nil {
		log.Error(err)
		helpers.SendResponseHTTP(c, http.StatusUnauthorized, "unauthorized", nil)
		c.Abort()
		return
	}

	tokenData.Token = auth

	c.Set("token", tokenData)

	c.Next()
}
