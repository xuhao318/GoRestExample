package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

var db *gorm.DB

func main() {
	// Connect to PostgreSQL
	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	db.AutoMigrate(&User{})

	r := gin.Default()
	r.GET("/users", getUsers)
	r.Run()
}

func getUsers(c *gin.Context) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error fetching users"})
		return
	}
	c.JSON(200, users)
}
