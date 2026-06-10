package main

import (
	"log"
	"net/http"

	"crudapi/config"
	"crudapi/database"
	"crudapi/handlers"
	"crudapi/middleware"
	"crudapi/repositories"
	"crudapi/services"

	"github.com/gin-gonic/gin"
)

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

type updateAlbumRequest struct {
	Title  *string  `json:"title"`
	Artist *string  `json:"artist"`
	Price  *float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func testPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func postAlbums(c *gin.Context) {
	var newAlbum album
	if err := c.BindJSON(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid album data"})
		return
	}
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		} // 17 dec 2025  17 may 2026
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Album not found"})
}

func updateAlbum(c *gin.Context) {
	id := c.Param("id")
	var updatedAlbum updateAlbumRequest

	if err := c.BindJSON(&updatedAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid album data"})
		return
	}

	for i := range albums {
		if albums[i].ID == id {
			if updatedAlbum.Title != nil {
				albums[i].Title = *updatedAlbum.Title
			}

			if updatedAlbum.Artist != nil {
				albums[i].Artist = *updatedAlbum.Artist
			}

			if updatedAlbum.Price != nil {
				albums[i].Price = *updatedAlbum.Price
			}
			c.IndentedJSON(http.StatusOK, albums[i])
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Album not found"})
}

func deleteAlbum(c *gin.Context) {
	id := c.Param("id")
	for i, a := range albums {
		if a.ID == id {
			albums = append(albums[:i], albums[i+1:]...)
			c.IndentedJSON(http.StatusOK, gin.H{"message": "Album deleted successfully"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Album not found"})
}

func main() {
	cfg := config.Load()

	// connect to database
	pool := database.Connect()
	defer pool.Close()

	// Wire up layers
	userRepo := repositories.NewUserRepository(pool)
	authService := services.NewAuthService(userRepo, cfg)
	authHandler := handlers.NewAuthHandler(authService, cfg, userRepo)

	router := gin.Default()

	router.GET("/ping", testPing)
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)
	router.PUT("/albums/:id", updateAlbum)
	router.DELETE("/albums/:id", deleteAlbum)

	// Public auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	// Protected routes — add JWT middleware
	api := router.Group("/api", middleware.AuthMiddleware(authService))
	{
		api.GET("/me", authHandler.Me)
	}

	if err := router.Run(cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

}
