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
	"golang.org/x/oauth2/google"
)

var oauthStateStringGl = "thisshouldberandom"
var oauthConfG1 = &oauth2.Config{
	ClientID:     config.GoogleClientID,
	ClientSecret: config.GoogleClientSecret,
	RedirectURL:  config.GoogleRedirectURL,
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// hehjeh
func (h Handler) HandleGoogleLogin(c echo.Context) error {

	URL, err := url.Parse(oauthConfG1.Endpoint.AuthURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	parameters := url.Values{}
	parameters.Add("client_id", oauthConfG1.ClientID)
	parameters.Add("scope", strings.Join(oauthConfG1.Scopes, " "))
	parameters.Add("redirect_uri", oauthConfG1.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateStringGl)

	URL.RawQuery = parameters.Encode()
	url := URL.String()
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h Handler) GoogleCallBack(c echo.Context) error {
	code := c.QueryParam("code")
	if code != "" {
		tok, err := oauthConfG1.Exchange(context.Background(), code)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(tok.AccessToken))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var google GoogleResponse
		if err = json.Unmarshal(body, &google); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		var account Account
		result := h.DB.First(&account, "email = ?", strings.ToLower(google.Email))
		if result.Error != nil {
			profile := Profile{
				Name:      google.Name,
				Role:      "google user",
				Photo:     google.Picture,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			account := Account{
				Profile:   profile,
				ProfileID: profile.ID,
				Email:     strings.ToLower(google.Email),
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
		google.Token = tokenString
		return c.JSON(http.StatusOK, google)
	}
	return c.JSON(http.StatusBadRequest, "code is empty!!")
}
