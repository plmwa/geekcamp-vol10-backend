package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"geekcamp-vol10-backend/internal/end"
	//"geekcamp-vol10-backend/internal/config"
	//"geekcamp-vol10-backend/internal/handlers33"
	//"geekcamp-vol10-backend/internal/middleware"
	//"geekcamp-vol10-backend/internal/repositories"
	//"geekcamp-vol10-backend/internal/services"
)

func main() {
	// Ginルーターを初期化
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})

	end.RegisterUserRoutes(r)
	end.GETUserRoutes(r)

	// サーバーをポート8080で起動
	if err := r.Run("localhost:8081"); err != nil {
		// エラーが発生した場合、ログに詳細を出力してプログラムを終了する
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}

	/*
		userRoutes := r.Group("/users")
		{
			userRoutes.GET("", userHandler.GetAll) // GET /users/
			userRoutes.POST("", userHandler.Register) // POST /users/
		}

		monsterRoutes := r.Group("/monsters")
		monsterRoutes.Use(authMiddleware.Authenticate)
		{
			monsterRoutes.GET("", monsterHandler.GetByUID)
		}

		if err := r.Run(":8080"); err != nil {
			log.Fatalf("サーバーの起動に失敗しました: %v", err)
		}
	*/
}
