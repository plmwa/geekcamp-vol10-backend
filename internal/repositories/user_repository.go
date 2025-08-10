package repositories

import (
	"context"

	"geekcamp-vol10-backend/internal/models"

	"cloud.google.com/go/firestore"
)

type UserRepository struct {
	Client *firestore.Client
}

func NewUserRepository(client *firestore.Client) *UserRepository {
	return &UserRepository{Client: client}
}

// userコレクション保存
func (r *UserRepository) SaveUser(ctx context.Context, user models.User) error {
	_, err := r.Client.Collection("users").Doc(user.FirebaseId).Set(ctx, map[string]interface{}{
		"githubUserName":       user.GithubUserName,
		"photoURL":             user.PhotoURL,
		"createdAt":            user.CreatedAt,
		"continuousSealRecord": user.ContinuousSealRecord, // contributionからもらう
		"maxSealRecord":        user.MaxSealRecord,
	})
	return err
}

// サブコレクションでcurrentMonster
func (r *UserRepository) SaveCurrentMonster(ctx context.Context, firebaseId string, monster models.CurrentMonster) error {
	_, err := r.Client.Collection("users").Doc(firebaseId).Collection("currentMonster").Doc("monster").Set(ctx, map[string]interface{}{
		"monsterId":                   monster.MonsterId,
		"progressContributions":       monster.ProgressContributions,
		"lastContributionReflectedAt": monster.LastContributionReflectedAt, // contributionからもらう
		"assignedAt":                  monster.AssignedAt,
		"requiredContributions":       monster.RequiredContributions, // contributionからもらう
	})
	return err
}

// SealedMonsters
func (r *UserRepository) SaveSealedMonster(ctx context.Context, firebaseId string, sealed models.SealedMonster) error {
	_, err := r.Client.Collection("users").Doc(firebaseId).Collection("sealedMonsters").Doc(sealed.MonsterId).Set(ctx, map[string]interface{}{
		"monsterId":   sealed.MonsterId,
		"monsterName": sealed.MonsterName,
		"SealedAt":    sealed.SealedAt,
	})
	return err
}

// ユーザーとサブコレクションを取得
func (r *UserRepository) GetUser(ctx context.Context, firebaseId string) (map[string]interface{}, error) {
	mainDoc, err := r.Client.Collection("users").Doc(firebaseId).Get(ctx)
	if err != nil {
		return nil, err
	}
	mainData := mainDoc.Data()

	subDocs, err := mainDoc.Ref.Collection("currentMonster").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	if len(subDocs) > 0 {
		mainData["currentMonster"] = subDocs[0].Data()
	}
	return mainData, nil
}

func GetUserByIDRepo(ctx context.Context, client *firestore.Client, id string) (*models.User, error) {
	doc, err := client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
