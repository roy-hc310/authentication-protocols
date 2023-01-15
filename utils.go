package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func ExtractClaimsHS256(token string, privateKey string) (jwt.MapClaims, error) {
	hmacSecretString := privateKey
	hmacSecret := []byte(hmacSecretString)

	key, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
		}
		return hmacSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalidate token: %w", err)
	}
	claims, ok := key.Claims.(jwt.MapClaims)
	if !ok || !key.Valid {
		return nil, fmt.Errorf("invalid token claim")
	}
	return claims, nil
}

func ExtractClaimsRS256(token string, publicKey string) (jwt.MapClaims, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("create: parse key: %w", err)
	}
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("validate: invalid token")
	}
	return claims, nil
}

func GenerateTokenRS256(ttl time.Duration, user interface{}, privateKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	now := time.Now().UTC()

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user
	claims["exp"] = now.Add(ttl).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims.VerifyExpiresAt(time.Now().Unix(), false)

	decodePrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("could not decode key: %w", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodePrivateKey)
	if err != nil {
		return "", fmt.Errorf("create: parse key: %w", err)
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("generating JWT Token failed: %w", err)
	}
	return tokenString, nil
}

func GenerateTokenHS256(ttl time.Duration, user interface{}, privateKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	now := time.Now().UTC()

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user
	claims["exp"] = now.Add(ttl).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims.VerifyExpiresAt(time.Now().Unix(), false)

	tokenString, err := token.SignedString([]byte(privateKey))

	if err != nil {
		return "", fmt.Errorf("generating JWT Token failed: %w", err)
	}

	return tokenString, nil
}

func GeneratePagination(c echo.Context) Pagination {
	limit := 3
	page := 1
	// sort := "created_at asc"
	query := c.Request().URL.Query()

	for key, value := range query {
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			limit, _ = strconv.Atoi(queryValue)
			// break myloop
		case "page":
			page, _ = strconv.Atoi(queryValue)
		}
	}
	return Pagination{Limit: limit, Page: page, Sort: "product_name"}
}
