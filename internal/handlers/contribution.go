package handlers

import (
	"context"
	"net/http"
	"os"
	"fmt"
	"geekcamp-vol10-backend/internal/services"
	"geekcamp-vol10-backend/internal/repositories"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


// githubのコントリビューション数を取得するハンドラー
// GET /contributions/:id
func GetContribution(c *gin.Context) {
	id := c.Param("id")
	// 一旦アクセストークンとユーザー名はenvから
	err := godotenv.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load environment variables"})
		return
	}

	// UserNameは:idを用いてDBから抽出する
	// アクセストークンはHeaderの中のAuthorizationから取得する
	ctx := context.Background()
	githubUserName, err := repositories.GetGitHubUserNameByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが見つかりません"})
		return
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	// tokenはc.Contextの中に入っているので、そこから取得する
	// Middlewareで設定した値を取得
    /*
	accessToken, exists := c.Get("githubAccessToken")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "アクセストークンがContextにありません"})
        return
    }
	*/

	// GitHubのAPIを叩くための準備
	if githubUserName == "" || githubToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHubのユーザー名とトークンは必須です"})
		return
	}

	// サービス層を呼び出してコントリビューション数を取得する
	githubData, err := services.GetContributions(githubUserName, githubToken)
	if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "コントリビューションの取得に失敗しました", // エラーメッセージを少し具体的に
        })
        return
    }
	
	currentMonster, err := repositories.SaveContribution(id, githubData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "内部エラーが発生しました",
		})
		return
	}
	// githubDataの中身をlogに出力
	fmt.Printf("GitHub Data: %+v\n", githubData)
	c.JSON(http.StatusOK, currentMonster)
}