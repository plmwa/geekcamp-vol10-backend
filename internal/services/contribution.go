package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"geekcamp-vol10-backend/internal/models"
)

const githubAPIURL = "https://api.github.com/graphql"

func GetContributions(githubUserName, githubToken string) (models.GithubResponse, error) {

	// 本番環境では、headerからアクセストークンを取得,bodyからユーザー名を取得
	/*
	if githubToken == "" {
		githubToken = c.GetHeader("Authorization")
	}*/

	// GitHub APIに送信するGraphQLクエリを定義 (詳細な日時情報のみ)
	query := `
        query($githubUserName: String!) {
            user(login: $githubUserName) {
                contributionsCollection {
                    commitContributionsByRepository {
                        repository {
                            name
                            owner {
                                login
                            }
                        }
                        contributions(first: 100) {
                            nodes {
                                commitCount
                                occurredAt
                                user {
                                    login
                                }
                            }
                        }
                    }
                }
            }
        }`

	// クエリと変数をリクエストボディにまとめる
	graphQLReq := models.GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"githubUserName": githubUserName,
		},
	}
	requestBody, err := json.Marshal(graphQLReq)
	if err != nil {
		log.Printf("ERROR: Failed to marshal GraphQL request: %v", err)
		return models.GithubResponse{}, fmt.Errorf("Failed to build request body: %w", err)
	}

	// GitHub APIへのHTTPリクエストを作成
	request, err := http.NewRequest("POST", githubAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("ERROR: Failed to create HTTP request: %v", err)
		return models.GithubResponse{}, fmt.Errorf("Failed to create HTTP request: %w", err)
	}
	request.Header.Set("Authorization", "bearer "+githubToken)
	request.Header.Set("Content-Type", "application/json") // Content-Typeの指定は必須

	// 4. リクエストを実行
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("ERROR: Failed to send request to GitHub: %v", err)
		return models.GithubResponse{}, fmt.Errorf("Failed to send request to GitHub: %w", err)
	}
	defer response.Body.Close()

	// GitHubからのレスポンスステータスコードをチェック
	if response.StatusCode != http.StatusOK {
		log.Printf("ERROR: GitHub API returned status code: %d", response.StatusCode)
		return models.GithubResponse{}, fmt.Errorf("GitHub API returned status code: %d", response.StatusCode)
	}

	// 5. レスポンスをデコード
	var githubResponse models.GithubResponse
	if err := json.NewDecoder(response.Body).Decode(&githubResponse); err != nil {
		log.Printf("ERROR: Failed to decode GitHub response: %v", err)
		return models.GithubResponse{}, fmt.Errorf("Failed to decode GitHub response: %w", err)
	}

	// GraphQLレベルのエラーもチェック
	if len(githubResponse.Errors) > 0 {
		log.Printf("ERROR: GraphQL error: %s", githubResponse.Errors[0].Message)
		return models.GithubResponse{}, fmt.Errorf("GraphQL error: %s", githubResponse.Errors[0].Message)
	}

	// Responseを出力 json
	return githubResponse, nil
}
