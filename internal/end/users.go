package end

import (
	"geekcamp-vol10-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine) {
	r.POST("/users", handlers.Users)
}

func GETUserRoutes(r *gin.Engine) {
	r.GET("/users/:id", handlers.GETUser)
}
