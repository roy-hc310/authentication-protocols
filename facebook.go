package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	facebookOAuth "golang.org/x/oauth2/facebook"
)

// heheh
var oauthStateStringFl = "thisshouldberandom"
var oauthConfF1 = &oauth2.Config{
	ClientID:     config.FacebookClientID,
	ClientSecret: config.FacebookClientSecret,
	RedirectURL:  config.FacebookRedirectURL,
	Endpoint:     facebookOAuth.Endpoint,
	Scopes:       []string{"public_profile", "email"},
}

func (h Handler) HandleFacebookLogin(c echo.Context) error {

	url := oauthConfF1.AuthCodeURL(oauthStateStringFl)
	return c.Redirect(http.StatusTemporaryRedirect, url)

}

func (h Handler) FacebookCallBack(c echo.Context) error {
	code := c.QueryParam("code")
	if code != "" {
		tok, err := oauthConfF1.Exchange(context.Background(), code)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=" + url.QueryEscape(tok.AccessToken))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var facebook FacebookResponse
		if err = json.Unmarshal(body, &facebook); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		var account Account
		result := h.DB.First(&account, "email = ?", strings.ToLower(facebook.Email))
		if result.Error != nil {
			profile := Profile{
				Name: facebook.Name,
				Role: "facebok user",
				// Photo:     google.Picture,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			account := Account{
				Profile:   profile,
				ProfileID: profile.ID,
				Email:     strings.ToLower(facebook.Email),
				Verified:  true,
			}
			result = h.DB.Create(&account)
			if result.Error != nil {
				return c.JSON(http.StatusBadGateway, result.Error)
			}
		}

		tokenString, err := GenerateTokenRS256(config.AccessTokenExpiresIn, account.ID, config.AccessTokenPrivateKey)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		tokenString = "Bearer " + tokenString

		var token Token
		result = h.DB.First(&token, "token = ?", tokenString)
		if result.Error != nil {
			token.Token = tokenString
			result = h.DB.Create(&token)
			if result.Error != nil {
				return c.JSON(http.StatusBadGateway, result.Error)
			}
		}
		facebook.Token = tokenString
		return c.JSON(http.StatusOK, facebook)
	}
	return c.JSON(http.StatusBadRequest, "empty!!")
}
