package repositories

import (
	"log"
	"geekcamp-vol10-backend/internal/models"
)

func SaveContribution(id string, githubData models.GithubResponse) error {
	// DBに保存する処理を実装
	log.Printf("Saving contribution for user %s: %+v", id, githubData)
	return nil
}
