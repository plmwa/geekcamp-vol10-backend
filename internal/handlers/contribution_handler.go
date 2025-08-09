package handlers

import (
	"net/http"
	"os"
	"geekcamp-vol10-backend/internal/services"
	"geekcamp-vol10-backend/internal/repositories"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// リクエストボディの構造体を定義
type postContributionRequest struct {
	GitHubUserName string `json:"githubUserName"`
}

// githubのコントリビューション数を取得するハンドラー
// POST /contributions/:id
func PostContribution(c *gin.Context) {
	id := c.Param("id")
	var req postContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストの形式が正しくありません: " + err.Error()})
		return
	}
	if req.GitHubUserName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "githubUserNameは必須です"})
		return
	}

	// 一旦アクセストークンはenvから
	err := godotenv.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load environment variables"})
		return
	}
	githubToken := os.Getenv("GITHUB_TOKEN")

	// サービス層を呼び出してコントリビューション数を取得する
	githubData, err := services.GetContributions(req.GitHubUserName, githubToken) 
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "コントリビューションの取得に失敗しました", // エラーメッセージを少し具体的に
        })
        return
    }
	err = repositories.SaveContribution(id, githubData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "内部エラーが発生しました",
		})
		return
	}
	c.Status(http.StatusNoContent)

}