package services

import (
	"context"
	"time"

	"geekcamp-vol10-backend/internal/models"
	"geekcamp-vol10-backend/internal/repositories"
	"geekcamp-vol10-backend/pkg/database"
)

type UserService struct {
	UserRepo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{UserRepo: repo}
}

func (s *UserService) CreateUserWithMonsters(ctx context.Context, firebaseId string) (map[string]interface{}, error) {
	user := models.User{
		FirebaseId:           firebaseId,
		GithubUserName:       "nkyouYaba",
		PhotoURL:             "https://avatars.githubusercontent.com/u/12345678?v=4",
		CreatedAt:            time.Now(),
		ContinuousSealRecord: 0,
		MaxSealRecord:        0,
	}
	if err := s.UserRepo.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	currentMonster := models.CurrentMonster{
		MonsterId:                   1,
		ProgressContributions:       0,
		LastContributionReflectedAt: time.Now(),
		AssignedAt:                  time.Now(),
		RequiredContributions:       30,
	}
	if err := s.UserRepo.SaveCurrentMonster(ctx, firebaseId, currentMonster); err != nil {
		return nil, err
	}

	sealedMonster := models.SealedMonster{
		MonsterId:   "001",
		MonsterName: "スライム",
		SealedAt:    0,
	}
	if err := s.UserRepo.SaveSealedMonster(ctx, firebaseId, sealedMonster); err != nil {
		return nil, err
	}

	return s.UserRepo.GetUser(ctx, firebaseId)
}

func GetUserByIDService(ctx context.Context, id string) (*models.User, error) {
	client := database.NewFirestoreClient(ctx)

	// ユーザー本体取得
	main, err := client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := main.DataTo(&user); err != nil {
		return nil, err
	}

	// currentMonsterの取得
	sub, err := client.Collection("users").Doc(id).Collection("currentMonster").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	if len(sub) > 0 {
		var cm models.CurrentMonster
		if err := sub[0].DataTo(&cm); err != nil {
			return nil, err
		}
		user.CurrentMonster = &cm
	} else {
		user.CurrentMonster = nil
	}

	return &user, nil
}
