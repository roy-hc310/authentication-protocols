package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func (h Handler) DeserializeUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var tokenStr string
		authorizationHeader := c.Request().Header.Get("Authorization")
		var token Token
		result := h.DB.First(&token, "token = ?", authorizationHeader)
		if result.Error != nil {
			return c.JSON(http.StatusBadRequest, "Token no longer exist")
		}
		fields := strings.Fields(authorizationHeader)
		if len(fields) != 0 && fields[0] == "Bearer" {
			tokenStr = fields[1]
		}

		claims, err := ExtractClaimsRS256(tokenStr, config.AccessTokenPublicKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}
		exp := claims["exp"].(float64)
		sub := claims["sub"].(string)
		tokenExpire := time.Unix(int64(exp), 0)
		if tokenExpire.Before(time.Now()) {
			return c.JSON(http.StatusBadRequest, "Your token had expired")
		}
		var account Account
		result = h.DB.First(&account, "id = ?", sub)
		if result.Error != nil {
			return c.JSON(http.StatusBadRequest, "Invalid email or token")
		}
		c.Set("account", account)
		return next(c)
	}
}
