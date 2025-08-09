package main

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"geekcamp-vol10-backend/internal/handlers"
	//"geekcamp-vol10-backend/internal/middleware"
)

func main() {
	// Ginルーターを初期化
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})

	// /contributions エンドポイントを定義
	//r.POST("/contributions/:id", middleware.AuthMiddleware(handlers.PostContribution))
	r.POST("/contributions/:id", handlers.PostContribution)

	// サーバーをポート8080で起動
	if err := r.Run("localhost:8080"); err != nil {
		// エラーが発生した場合、ログに詳細を出力してプログラムを終了する
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
