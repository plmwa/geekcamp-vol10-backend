package repositories

import (
	"geekcamp-vol10-backend/internal/models"
)

func SaveContribution(id string, githubData models.GithubResponse) (map[string]interface{}) {
	// githubDataを用いながらDBに保存
	// DBのUsersコレクションの:idの人のcurrentMonsterを返す
	currentMonster := map[string]interface{}{
		"monsterId": "monster-002",
		"progressContributions": 25,
		"requiredContributions": 30,
		"lastContributionReflectedAt": "2025-08-09T22:50:00Z", // 更新日時
		"assignedAt": "2025-08-01T18:00:00Z",
	}

	return currentMonster
}

func GetGitHubUserNameByID(id string) string {
	// DBからユーザー名を取得する処理を実装
	return "plmwa"
}