package models

import (

)

// GraphQLリクエストの構造体
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// GitHubからのレスポンスを格納する構造体
type GithubResponse struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []struct {
						ContributionDays []struct {
							Color             string `json:"color"`
							ContributionCount int    `json:"contributionCount"`
							Date              string `json:"date"`
							Weekday           int    `json:"weekday"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct { // GraphQLレベルのエラーも考慮
		Message string `json:"message"`
	} `json:"errors"`
}