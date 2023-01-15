package main

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Token string `gorm:"primary_key"`
}

type Pagination struct {
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Sort  string `json:"sort"`
}

type Import struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	// ImportDetailID uuid.UUID `gorm:"type:uuid"`

	Status      string `json:"status"`
	Description string `json:"description"`

	ImportDetail []ImportDetail `gorm:"foreignkey:ImportID;association_foreignkey:id" json:"import_detail"`
}

type ImportDetail struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	ImportID uuid.UUID //`gorm:"import_id"`

	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float32 `json:"price"`
	TotalPrice  float32 `json:"total_price"`
}

type Product struct {
	ProductID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`

	ProductName     string `json:"product_name" gorm:"unique;not null"`
	CurrentQuantity int    `json:"current_quantity"`
	// ImportPrice     float32 `json:"import_price"`
	// ExportPrice     float32 `json:"export_price"`

	Unit        string `json:"unit"`
	Category    string `json:"category"`
	Description string `json:"description"`
	SKU         string `json:"sku"`
	Photo       string `json:"photo"`

	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
	DeletedBy string `json:"delete_by"`
	// CreatedAt time.Time `json:"created_at" gorm:"not null"`
	// UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
	// DeleteAt  time.Time `json:"delete_at" gorm:"not null"`
}

type Account struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	ProfileID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Email              string    `gorm:"uniqueIndex;not null"`
	Password           string    `gorm:"not null"`
	AccountType        string    `gorm:"not null"`
	VerificationCode   string
	Verified           bool
	PasswordResetToken string
	PasswordResetAt    time.Time

	Profile Profile `gorm:"foreignKey:ID;references:ProfileID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Profile struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Role      string    `gorm:"type:varchar(255);not null"`
	Photo     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
	Role            string `json:"role" binding:"required"`
	Photo           string `json:"photo" binding:"required"`
}

type SignInInput struct {
	Email    string `json:"email"  binding:"required"`
	Password string `json:"password"  binding:"required"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordInput struct {
	Password        string `json:"password" binding:"required"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
}

type GoogleResponse struct {
	ID            string `json:"id" binding:"required"`
	Email         string `json:"email" binding:"required"`
	Name          string `json:"name" binding:"required"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	Token         string `json:"token"`
}

type FacebookResponse struct {
	ID    string `json:"id" binding:"required"`
	Email string `json:"email" binding:"required"`
	Name  string `json:"name" binding:"required"`
	// Picture string `json:"picture"`
	Token string `json:"token"`
}

type AccountResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Photo     string    `json:"photo,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
