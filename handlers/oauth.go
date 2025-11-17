package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"GoRestExample/config"
	"GoRestExample/models"
	"GoRestExample/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// OAuth2Provider represents an OAuth2 provider
type OAuth2Provider struct {
	Config       *oauth2.Config
	UserInfoURL  string
	ProviderName models.OAuthProvider
}

// OAuthUserInfo represents user information from OAuth providers
type OAuthUserInfo struct {
	ID       string
	Email    string
	Name     string
	Avatar   string
	Verified bool
}

// GetOAuthConfig returns OAuth2 config for a specific provider
func GetOAuthConfig(provider string) (*OAuth2Provider, error) {
	baseURL := config.AppConfig.OAuth.RedirectBaseURL
	redirectURL := fmt.Sprintf("%s/auth/oauth/%s/callback", baseURL, provider)

	switch provider {
	case "google":
		if !config.AppConfig.OAuth.Google.Enabled {
			return nil, fmt.Errorf("Google OAuth is not enabled")
		}
		return &OAuth2Provider{
			Config: &oauth2.Config{
				ClientID:     config.AppConfig.OAuth.Google.ClientID,
				ClientSecret: config.AppConfig.OAuth.Google.ClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
				Endpoint:     google.Endpoint,
			},
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
			ProviderName: models.ProviderGoogle,
		}, nil

	case "github":
		if !config.AppConfig.OAuth.GitHub.Enabled {
			return nil, fmt.Errorf("GitHub OAuth is not enabled")
		}
		return &OAuth2Provider{
			Config: &oauth2.Config{
				ClientID:     config.AppConfig.OAuth.GitHub.ClientID,
				ClientSecret: config.AppConfig.OAuth.GitHub.ClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"user:email"},
				Endpoint:     github.Endpoint,
			},
			UserInfoURL:  "https://api.github.com/user",
			ProviderName: models.ProviderGitHub,
		}, nil

	case "microsoft":
		if !config.AppConfig.OAuth.Microsoft.Enabled {
			return nil, fmt.Errorf("Microsoft OAuth is not enabled")
		}
		return &OAuth2Provider{
			Config: &oauth2.Config{
				ClientID:     config.AppConfig.OAuth.Microsoft.ClientID,
				ClientSecret: config.AppConfig.OAuth.Microsoft.ClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"User.Read"},
				Endpoint:     microsoft.AzureADEndpoint("common"),
			},
			UserInfoURL:  "https://graph.microsoft.com/v1.0/me",
			ProviderName: models.ProviderMicrosoft,
		}, nil

	case "facebook":
		if !config.AppConfig.OAuth.Facebook.Enabled {
			return nil, fmt.Errorf("Facebook OAuth is not enabled")
		}
		return &OAuth2Provider{
			Config: &oauth2.Config{
				ClientID:     config.AppConfig.OAuth.Facebook.ClientID,
				ClientSecret: config.AppConfig.OAuth.Facebook.ClientSecret,
				RedirectURL:  redirectURL,
				Scopes:       []string{"email", "public_profile"},
				Endpoint:     facebook.Endpoint,
			},
			UserInfoURL:  "https://graph.facebook.com/me?fields=id,name,email,picture",
			ProviderName: models.ProviderFacebook,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// OAuthLogin initiates OAuth login flow
func OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")

	oauthProvider, err := GetOAuthConfig(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate state for CSRF protection
	state := fmt.Sprintf("%s-%d", provider, time.Now().Unix())

	// In production, store state in session/redis with expiration
	// For now, we'll use a simple state parameter
	authURL := oauthProvider.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// OAuthCallback handles OAuth callback
func OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	// In production, verify state parameter against stored value
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter missing"})
		return
	}

	oauthProvider, err := GetOAuthConfig(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := oauthProvider.Config.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info from provider
	userInfo, err := fetchOAuthUserInfo(ctx, oauthProvider, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	// Find or create user
	var user models.User
	result := models.DB.Where("email = ?", userInfo.Email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new one
		user = models.User{
			Name:          userInfo.Name,
			Email:         userInfo.Email,
			OAuthProvider: oauthProvider.ProviderName,
			OAuthID:       userInfo.ID,
			Avatar:        userInfo.Avatar,
			Role:          models.RoleUser,
			Verified:      userInfo.Verified,
		}

		if err := models.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		// User exists, update OAuth info if needed
		if user.OAuthProvider == models.ProviderLocal {
			// Link OAuth to existing local account
			user.OAuthProvider = oauthProvider.ProviderName
			user.OAuthID = userInfo.ID
			user.Avatar = userInfo.Avatar
			if err := models.DB.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
				return
			}
		}
	}

	// Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		string(user.Role),
		config.AppConfig.JWT.Secret,
		config.AppConfig.JWT.AccessExpiry,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		config.AppConfig.JWT.RefreshSecret,
		config.AppConfig.JWT.RefreshExpiry,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Store refresh token
	refreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(config.AppConfig.JWT.RefreshExpiry),
	}
	if err := models.DB.Create(&refreshTokenModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(config.AppConfig.JWT.AccessExpiry.Seconds()),
		User:         user.ToResponse(),
	})
}

// fetchOAuthUserInfo fetches user information from OAuth provider
func fetchOAuthUserInfo(ctx context.Context, provider *OAuth2Provider, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := provider.Config.Client(ctx, token)
	resp, err := client.Get(provider.UserInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	userInfo := &OAuthUserInfo{}

	switch provider.ProviderName {
	case models.ProviderGoogle:
		userInfo.ID = data["id"].(string)
		userInfo.Email = data["email"].(string)
		userInfo.Name = data["name"].(string)
		if picture, ok := data["picture"].(string); ok {
			userInfo.Avatar = picture
		}
		if verified, ok := data["verified_email"].(bool); ok {
			userInfo.Verified = verified
		}

	case models.ProviderGitHub:
		userInfo.ID = fmt.Sprintf("%v", data["id"])
		if email, ok := data["email"].(string); ok && email != "" {
			userInfo.Email = email
		} else {
			// GitHub might not return email in main endpoint
			userInfo.Email = fetchGitHubEmail(ctx, client)
		}
		userInfo.Name = data["name"].(string)
		if avatar, ok := data["avatar_url"].(string); ok {
			userInfo.Avatar = avatar
		}
		userInfo.Verified = true // GitHub emails are verified

	case models.ProviderMicrosoft:
		userInfo.ID = data["id"].(string)
		if email, ok := data["mail"].(string); ok && email != "" {
			userInfo.Email = email
		} else if userPrincipalName, ok := data["userPrincipalName"].(string); ok {
			userInfo.Email = userPrincipalName
		}
		userInfo.Name = data["displayName"].(string)
		userInfo.Verified = true // Microsoft emails are verified

	case models.ProviderFacebook:
		userInfo.ID = data["id"].(string)
		userInfo.Email = data["email"].(string)
		userInfo.Name = data["name"].(string)
		if picture, ok := data["picture"].(map[string]interface{}); ok {
			if pictureData, ok := picture["data"].(map[string]interface{}); ok {
				if url, ok := pictureData["url"].(string); ok {
					userInfo.Avatar = url
				}
			}
		}
		userInfo.Verified = true
	}

	return userInfo, nil
}

// fetchGitHubEmail fetches primary verified email from GitHub
func fetchGitHubEmail(ctx context.Context, client *http.Client) string {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var emails []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}

	// Find primary verified email
	for _, email := range emails {
		if primary, ok := email["primary"].(bool); ok && primary {
			if verified, ok := email["verified"].(bool); ok && verified {
				if emailStr, ok := email["email"].(string); ok {
					return emailStr
				}
			}
		}
	}

	return ""
}
