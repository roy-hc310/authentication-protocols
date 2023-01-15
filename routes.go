package main

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	h := &Handler{DB: db}

	e.GET("/testing", h.Testing)
	e.POST("/signup", h.SignUp)
	e.POST("/signin", h.SignIn)
	e.GET("/google", h.HandleGoogleLogin)
	e.GET("/google-callback", h.GoogleCallBack)
	e.GET("/facebook", h.HandleFacebookLogin)
	e.GET("/facebook-callback", h.FacebookCallBack)

	e.GET("/verify/:code", h.VerifyEmail)
	e.POST("/forgot", h.ForgotPassword)
	e.PATCH("/reset/:resetToken", h.ResetPassword)

	auth := e.Group("/auth")
	auth.Use(h.DeserializeUser)

	auth.GET("/logout", h.Logout)
	auth.GET("/profile", h.Profile)

	e.POST("/product", h.AddProduct)
	e.PATCH("/product/:product_name", h.UpdateProduct)
	e.DELETE("/product/:product_name", h.DeleteProduct)
	e.GET("/product/:product_name", h.GetProduct)
	e.GET("/product", h.GetProducts)

	e.POST("/import", h.AddImport)
	e.PATCH("/import/:import_id", h.UpdateImport)
}
