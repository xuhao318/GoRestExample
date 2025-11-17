package handlers

import (
	"net/http"

	"GoRestExample/models"

	"github.com/gin-gonic/gin"
)

// GetUsers returns all users (admin only)
func GetUsers(c *gin.Context) {
	var users []models.User
	if err := models.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}

	// Convert to response format (remove sensitive data)
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"count": len(userResponses),
	})
}

// GetUser returns a specific user by ID
func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateUser updates user information (user can update own profile, admin can update any)
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get("user_role")
	currentUserID, _ := c.Get("user_id")

	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is updating their own profile or is admin
	if userRole != string(models.RoleAdmin) && user.ID != currentUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
		return
	}

	// Bind request data
	var updateData struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update allowed fields
	if updateData.Name != "" {
		user.Name = updateData.Name
	}
	if updateData.Avatar != "" {
		user.Avatar = updateData.Avatar
	}

	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteUser deletes a user (admin only)
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Soft delete
	if err := models.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// UpdateUserRole updates a user's role (admin only)
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var roleData struct {
		Role models.UserRole `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&roleData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if roleData.Role != models.RoleAdmin && roleData.Role != models.RoleUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'admin' or 'user'"})
		return
	}

	user.Role = roleData.Role

	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user":    user.ToResponse(),
	})
}
