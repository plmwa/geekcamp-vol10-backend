package main

import (
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
)

func main() {
	// Ginルーターを初期化
	r := gin.Default()

	// /ping エンドポイントを定義
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "peng",
		})
	})
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})
	r.GET("/bye", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Goodbye, World!",
		})
	})
	r.GET("/add-bye", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Goodbye, World!",
		})
	})

	// サーバーをポート8080で起動
	// r.Runが返すエラーをチェックする
	if err := r.Run("localhost:8080"); err != nil {
		// エラーが発生した場合、ログに詳細を出力してプログラムを終了する
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}