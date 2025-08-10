package handlers

import (
	"context"
	"net/http"

	"geekcamp-vol10-backend/internal/repositories"
	"geekcamp-vol10-backend/internal/services"
	"geekcamp-vol10-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

// Users ハンドラー
func Users(c *gin.Context) {
	var req struct {
		FirebaseId string `json:"firebaseId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	client := database.NewFirestoreClient(ctx)

	userRepo := repositories.NewUserRepository(client)
	userService := services.NewUserService(userRepo)

	mainData, err := userService.CreateUserWithMonsters(ctx, req.FirebaseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, mainData)
}

func GETUser(c *gin.Context) {
	id := c.Param("id") // URL の :id 部分を取得

	user, err := services.GetUserByIDService(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
