package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/thanhpk/randstr"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func (h Handler) Testing(c echo.Context) error {

	return c.JSON(http.StatusOK, "hehe")
}

func (h Handler) ForgotPassword(c echo.Context) error {
	var forgot_password *ForgotPasswordInput
	if err := c.Bind(&forgot_password); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var account Account
	result := h.DB.First(&account, "email = ?", strings.ToLower(forgot_password.Email))
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, "Invalid email or Password")
	}
	if !account.Verified {
		return c.JSON(http.StatusUnauthorized, "Account not verified")
	}
	code := randstr.String(20)
	passwordResetToken := Encode(code)
	account.PasswordResetToken = passwordResetToken
	account.PasswordResetAt = time.Now().Add(time.Minute * 15)
	h.DB.Save(&account)

	email := EmailData{
		URL:       config.BaseUrl + "/reset/" + passwordResetToken,
		FirstName: account.Profile.Name,
		Subject:   "Your password reset token (valid for 10min)",
	}
	SendEmail(&account, &email, "verificationCode.html")
	return c.JSON(http.StatusOK, "Email verified successfully")
}

func (h Handler) VerifyEmail(c echo.Context) error {
	code := c.Param("code")

	var account Account
	result := h.DB.First(&account, "verification_code = ?", code)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, "Invalid verification code or account doesn't exists")
	}
	if account.Verified {
		return c.JSON(http.StatusConflict, "Account already verified")
	}
	account.VerificationCode = ""
	account.Verified = true
	h.DB.Save(&account)
	return c.JSON(http.StatusOK, "Email verified successfully")
}

func (h Handler) ResetPassword(c echo.Context) error {
	var reset_password *ResetPasswordInput
	code := c.Param("resetToken")

	if err := c.Bind(&reset_password); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if reset_password.Password != reset_password.PasswordConfirm {
		return c.JSON(http.StatusBadRequest, "Passwords do not match")
	}
	hashed_password, _ := HashPassword(reset_password.Password)

	var account Account
	result := h.DB.First(&account, "password_reset_token = ?", code)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, "The reset token is invalid or has expired")
	}
	account.Password = hashed_password
	account.PasswordResetToken = ""
	h.DB.Save(&account)

	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = ""
	cookie.Expires = time.Now()
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, "Password data updated successfully")
}

func (h Handler) Profile(c echo.Context) error {
	account := c.Get("account").(Account)

	var accountResponse AccountResponse
	accountResponse.ID = account.ID
	accountResponse.Name = account.Profile.Name
	accountResponse.Email = account.Email
	accountResponse.Role = account.Profile.Role
	accountResponse.Photo = account.Profile.Photo
	accountResponse.CreatedAt = account.Profile.CreatedAt
	accountResponse.UpdatedAt = account.Profile.UpdatedAt
	return c.JSON(http.StatusOK, accountResponse)
}

func (h Handler) SignUp(c echo.Context) error {
	var signup *SignUpInput
	if err := c.Bind(&signup); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var account Account
	result := h.DB.First(&account, "email = ?", strings.ToLower(signup.Email))
	if result.Error == nil {
		return c.JSON(http.StatusBadRequest, "Email already existed")
	}

	if strings.TrimSpace(signup.Email) == "" {
		return c.JSON(http.StatusBadRequest, "email cannot be empty")
	}
	if strings.TrimSpace(signup.Password) == "" {
		return c.JSON(http.StatusBadRequest, "passwords cannot be empty")
	}
	if signup.Password != signup.PasswordConfirm {
		return c.JSON(http.StatusBadRequest, "passwords do not match")
	}
	hashedPassword, err := HashPassword(signup.Password)
	if err != nil {
		return c.JSON(http.StatusBadGateway, err.Error())
	}
	code := randstr.String(20)
	verification_code := Encode(code)

	profile := Profile{
		Name:      signup.Name,
		Role:      signup.Role,
		Photo:     signup.Photo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	account = Account{
		Profile:          profile,
		ProfileID:        profile.ID,
		Email:            strings.ToLower(signup.Email),
		Password:         hashedPassword,
		Verified:         false,
		VerificationCode: verification_code,
	}

	result = h.DB.Create(&account)
	if result.Error != nil {
		return c.JSON(http.StatusBadGateway, result.Error)
	}

	// Send Email
	email := EmailData{
		URL:       config.BaseUrl + "/verify/" + verification_code,
		FirstName: account.Profile.Name,
		Subject:   "Your account verification code",
	}
	SendEmail(&account, &email, "verificationCode.html")
	return c.JSON(http.StatusCreated, "success")
}

func (h Handler) SignIn(c echo.Context) error {
	var signin *SignInInput
	if err := c.Bind(&signin); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var account Account
	result := h.DB.First(&account, "email = ?", strings.ToLower(signin.Email))
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, "Invalid email")
	}
	if !account.Verified {
		return c.JSON(http.StatusForbidden, "Please verify your email")
	}
	if err := VerifyPassword(account.Password, signin.Password); err != nil {
		return c.JSON(http.StatusBadRequest, "Wrong Password")
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
	return c.JSON(http.StatusOK, token)
}

func (h Handler) Logout(c echo.Context) error {
	authorizationHeader := c.Request().Header.Get("Authorization")

	var token Token
	result := h.DB.First(&token, "token = ?", authorizationHeader)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, "You are not logged in")
	}
	h.DB.Delete(&token)
	return c.JSON(http.StatusOK, "Log out successfully")
}

func (h Handler) AddProduct(c echo.Context) error {
	// account := c.Get("account").(Account)

	var product Product
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	result := h.DB.First(&product, "product_name = ?", product.ProductName)
	if result.Error == nil {
		return c.JSON(http.StatusBadRequest, "Product name already existed")
	}
	result = h.DB.Create(&product)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}
	return c.JSON(http.StatusOK, product)
}

func (h Handler) UpdateProduct(c echo.Context) error {
	// account := c.Get("account").(Account)

	product_name := c.Param("product_name")
	var product Product
	result := h.DB.First(&product, "product_name = ?", product_name)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	h.DB.Save(&product)
	return c.JSON(http.StatusOK, product)
}

func (h Handler) DeleteProduct(c echo.Context) error {
	// account := c.Get("account").(Account)

	product_name := c.Param("product_name")
	var product Product
	result := h.DB.First(&product, "product_name = ?", product_name)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}
	h.DB.Delete(&product)
	return c.JSON(http.StatusOK, product)
}

func (h Handler) GetProduct(c echo.Context) error {
	product_name := c.Param("product_name")
	var product Product
	result := h.DB.First(&product, "product_name = ?", product_name)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	return c.JSON(http.StatusOK, product)
}

func (h Handler) GetProducts(c echo.Context) error {
	pagination := GeneratePagination(c)
	var product []Product
	offset := (pagination.Page - 1) * pagination.Limit
	queryBuilder := h.DB.Limit(pagination.Limit).Offset(offset).Order(pagination.Sort)
	result := queryBuilder.Find(&product)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	return c.JSON(http.StatusOK, product)
}

func (h Handler) AddImport(c echo.Context) error {
	// account := c.Get("account").(Account)

	var imp Import
	if err := c.Bind(&imp); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	result := h.DB.First(&imp, "id = ?", imp.ID)
	if result.Error == nil {
		return c.JSON(http.StatusBadRequest, "Import ID already existed")
	}
	result = h.DB.Create(&imp)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}
	return c.JSON(http.StatusOK, imp)
	//////////////////////////////////////////////////////////////////
	// importDetail := []ImportDetail{
	// 	{ProductName: "hehe1"},
	// 	{ProductName: "hehe2"},
	// }

	// imp := Import{
	// 	Status:       "hoho",
	// 	Description:  "I1",
	// 	ImportDetail: importDetail,
	// }
	// h.DB.Create(&imp)
	// return c.JSON(http.StatusOK, imp)
}

func (h Handler) UpdateImport(c echo.Context) error {
	// account := c.Get("account").(Account)

	import_id := c.Param("import_id")
	var imp Import
	result := h.DB.First(&imp, "id = ?", import_id)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}
	if err := c.Bind(&imp); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	h.DB.Save(&imp)
	return c.JSON(http.StatusOK, imp)
}

func (h Handler) DeleteImport(c echo.Context) error {
	// account := c.Get("account").(Account)

	import_id := c.Param("import_id")
	var imp Import
	result := h.DB.First(&imp, "id = ?", import_id)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}
	h.DB.Delete(&imp)
	return c.JSON(http.StatusOK, imp)
}

func (h Handler) GetImport(c echo.Context) error {
	import_id := c.Param("import_id")
	var imp Import
	result := h.DB.First(&imp, "id = ?", import_id)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	return c.JSON(http.StatusOK, imp)
}

func (h Handler) GetImports(c echo.Context) error {
	pagination := GeneratePagination(c)
	var imp []Import
	offset := (pagination.Page - 1) * pagination.Limit
	queryBuilder := h.DB.Limit(pagination.Limit).Offset(offset).Order(pagination.Sort)
	result := queryBuilder.Find(&imp)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	return c.JSON(http.StatusOK, imp)
}
