package models

import "time"

type User struct {
	FirebaseId           string          `json:"firebaseId"`
	GithubUserName       string          `json:"githubUserName"`
	PhotoURL             string          `json:"photoURL"`
	CreatedAt            time.Time       `json:"createdAt"`
	ContinuousSealRecord int             `json:"continuousSealRecord"`
	MaxSealRecord        int             `json:"maxSealRecord"`
	CurrentMonster       *CurrentMonster `firestore:"currentMonster,omitempty"`
}

type CurrentMonster struct {
	MonsterId                   string    `json:"monsterId"`
	ProgressContributions       int       `json:"progressContributions"`
	LastContributionReflectedAt time.Time `json:"lastContributionReflectedAt"`
	AssignedAt                  time.Time `json:"assignedAt"`
	RequiredContributions       int       `json:"requiredContributions"`
}

type SealedMonster struct {
	MonsterId   string `json:"monsterId"`
	MonsterName string `json:"monsterName"`
	SealedAt    int    `json:"sealedAt"`
}
