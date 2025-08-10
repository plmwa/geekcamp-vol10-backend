package handlers

import (
	"context"
	"log"
	"net/http"

	"geekcamp-vol10-backend/internal/services"
	"geekcamp-vol10-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

// Users ハンドラー
func Users(c *gin.Context) {
	log.Printf("POST /users エンドポイントが呼び出されました")
	
	var req struct {
		FirebaseId     string `json:"firebaseId"`
		GithubUserName string `json:"githubUserName"`
		PhotoURL       string `json:"photoURL"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("JSONバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("受信したリクエスト: FirebaseId='%s', GithubUserName='%s', PhotoURL='%s'", 
		req.FirebaseId, req.GithubUserName, req.PhotoURL)

	ctx := context.Background()
	client := database.GetFirestoreClient()
	if client == nil {
		log.Printf("Firestoreクライアントの取得に失敗")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}
	
	log.Printf("Firestoreクライアントの取得成功")

	// services.CreateUserを使用してユーザーを作成
	userData, err := services.CreateUser(ctx, client, req.FirebaseId, req.GithubUserName, req.PhotoURL)
	if err != nil {
		log.Printf("CreateUser: ユーザー作成に失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	log.Printf("ユーザーデータの保存に成功しました: FirebaseId='%s'", req.FirebaseId)
	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user":    userData,
	})
}

func GETUser(c *gin.Context) {
	id := c.Param("id") // URL の :id 部分を取得
	log.Printf("ユーザーID '%s' の情報を取得中...", id)

	user, err := services.GetUserByIDService(c, id)
	if err != nil {
		log.Printf("ユーザーID '%s' の取得に失敗しました: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
		return
	}

	log.Printf("ユーザーID '%s' の情報を正常に取得しました", id)
	c.JSON(http.StatusOK, user)
}
