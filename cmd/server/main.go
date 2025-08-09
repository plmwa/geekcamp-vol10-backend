package main

import (
	"geekcamp-vol10-backend/internal/end"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ginルーターを初期化
	r := gin.Default()

	// /ping エンドポイントを定義
	r.GET("/ping", end.Ping)
	r.GET("/hello", end.Hello)
	r.GET("/bye", end.Bye)

	// サーバーをポート8080で起動
	r.Run(":8080")
}
