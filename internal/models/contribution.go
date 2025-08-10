package models

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
				CommitContributionsByRepository []struct {
					Repository struct {
						Name  string `json:"name"`
						Owner struct {
							Login string `json:"login"`
						} `json:"owner"`
					} `json:"repository"`
					Contributions struct {
						Nodes []struct {
							CommitCount int    `json:"commitCount"`
							OccurredAt  string `json:"occurredAt"`
							User        struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"contributions"`
				} `json:"commitContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct { // GraphQLレベルのエラーも考慮
		Message string `json:"message"`
	} `json:"errors"`
}

type CurrentMonster struct {
	MonsterID                   string `json:"monsterId"`
	ProgressContributions       int    `json:"progressContributions"`
	RequiredContributions       int    `json:"requiredContributions"`
	LastContributionReflectedAt string `json:"lastContributionReflectedAt"`
	AssignedAt                  string `json:"assignedAt"`
}
