package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole defines the role of a user
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleUser    UserRole = "user"
)

// OAuthProvider defines the OAuth provider
type OAuthProvider string

const (
	ProviderLocal    OAuthProvider = "local"
	ProviderGoogle   OAuthProvider = "google"
	ProviderGitHub   OAuthProvider = "github"
	ProviderMicrosoft OAuthProvider = "microsoft"
	ProviderFacebook OAuthProvider = "facebook"
)

// User represents a user in the system
type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Email         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"type:varchar(255)" json:"-"` // Never expose password in JSON
	Role          UserRole       `gorm:"type:varchar(20);default:'user'" json:"role"`
	OAuthProvider OAuthProvider  `gorm:"type:varchar(20);default:'local'" json:"oauth_provider"`
	OAuthID       string         `gorm:"type:varchar(255)" json:"oauth_id,omitempty"`
	Avatar        string         `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	Verified      bool           `gorm:"default:false" json:"verified"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// RefreshToken represents a refresh token for JWT authentication
type RefreshToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
	Token     string         `gorm:"type:varchar(500);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserResponse represents the user data returned to clients
type UserResponse struct {
	ID            uint          `json:"id"`
	Name          string        `json:"name"`
	Email         string        `json:"email"`
	Role          UserRole      `json:"role"`
	OAuthProvider OAuthProvider `json:"oauth_provider"`
	Avatar        string        `json:"avatar,omitempty"`
	Verified      bool          `json:"verified"`
	CreatedAt     time.Time     `json:"created_at"`
}

// ToResponse converts a User to UserResponse (safe for API responses)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		OAuthProvider: u.OAuthProvider,
		Avatar:        u.Avatar,
		Verified:      u.Verified,
		CreatedAt:     u.CreatedAt,
	}
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
