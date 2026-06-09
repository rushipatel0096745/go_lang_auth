package models

import "time"

type Provider string

const (
	ProviderEmail  Provider = "email"
	ProviderGithub Provider = "github"
	ProviderGoogle Provider = "google"
)

type User struct {
    ID             string    `json:"id" db:"id"`
    Email          string    `json:"email" db:"email"`
    Name           string    `json:"name" db:"name"`
    AvatarURL      string    `json:"avatar_url" db:"avatar_url"`
    Provider       Provider  `json:"provider" db:"provider"`
    PasswordHash   string    `json:"-" db:"password_hash"` // "-" means never sent to client
    GoogleID       string    `json:"-" db:"google_id"`
    EmailVerified  bool      `json:"email_verified" db:"email_verified"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
    ID        string    `db:"id"`
    UserID    string    `db:"user_id"`
    TokenHash string    `db:"token_hash"`
    ExpiresAt time.Time `db:"expires_at"`
    CreatedAt time.Time `db:"created_at"`
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Name     string `json:"name" binding:"required"`
}

type AuthResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    User         *User  `json:"user"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

