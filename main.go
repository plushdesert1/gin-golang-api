package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
)

type User struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Post struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	AuthorID  uint      `json:"author_id" gorm:"not null"`
	Author    User      `json:"author" gorm:"foreignkey:AuthorID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

var users []User
var posts []Post
var userCounter uint = 1
var postCounter uint = 1

func main() {
	r := gin.New()

	// Middleware
	r.Use(logger.SetLogger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "gin-golang-api",
			"timestamp": time.Now().UTC(),
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Gin Golang API Starter",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health": "/health",
				"users": gin.H{
					"GET":    "/users",
					"POST":   "/users",
					"GET":    "/users/:id",
					"PUT":    "/users/:id",
					"DELETE": "/users/:id",
				},
				"posts": gin.H{
					"GET":    "/posts",
					"POST":   "/posts",
					"GET":    "/posts/:id",
					"PUT":    "/posts/:id",
					"DELETE": "/posts/:id",
				},
			},
		})
	})

	// User routes
	usersGroup := r.Group("/users")
	{
		usersGroup.GET("", getUsers)
		usersGroup.POST("", createUser)
		usersGroup.GET("/:id", getUser)
		usersGroup.PUT("/:id", updateUser)
		usersGroup.DELETE("/:id", deleteUser)
	}

	// Post routes
	postsGroup := r.Group("/posts")
	{
		postsGroup.GET("", getPosts)
		postsGroup.POST("", createPost)
		postsGroup.GET("/:id", getPost)
		postsGroup.PUT("/:id", updatePost)
		postsGroup.DELETE("/:id", deletePost)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

func createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	for _, user := range users {
		if user.Username == req.Username || user.Email == req.Email {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
	}

	user := User{
		ID:       userCounter,
		Username: req.Username,
		Email:    req.Email,
	}

	users = append(users, user)
	userCounter++

	c.JSON(http.StatusCreated, user)
}

func getUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	for _, user := range users {
		if user.ID == uint(id) {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func updateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, user := range users {
		if user.ID == uint(id) {
			// Check if new username/email conflicts with existing users
			for _, otherUser := range users {
				if otherUser.ID != user.ID && (otherUser.Username == req.Username || otherUser.Email == req.Email) {
					c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
					return
				}
			}

			users[i].Username = req.Username
			users[i].Email = req.Email
			users[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, users[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func deleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	for i, user := range users {
		if user.ID == uint(id) {
			users = append(users[:i], users[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func getPosts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"count": len(posts),
	})
}

func createPost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For demo purposes, assign to first user or create one
	var authorID uint = 1
	if len(users) > 0 {
		authorID = users[0].ID
	}

	post := Post{
		ID:       postCounter,
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
	}

	posts = append(posts, post)
	postCounter++

	c.JSON(http.StatusCreated, post)
}

func getPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	for _, post := range posts {
		if post.ID == uint(id) {
			c.JSON(http.StatusOK, post)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
}

func updatePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, post := range posts {
		if post.ID == uint(id) {
			posts[i].Title = req.Title
			posts[i].Content = req.Content
			posts[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, posts[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
}

func deletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	for i, post := range posts {
		if post.ID == uint(id) {
			posts = append(posts[:i], posts[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
}