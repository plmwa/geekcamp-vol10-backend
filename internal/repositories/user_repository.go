package repositories

import (
	"context"
	"fmt"
	"log"

	"geekcamp-vol10-backend/internal/models"

	"cloud.google.com/go/firestore"
)

// UserRepository 構造体
type UserRepository struct {
	Client *firestore.Client
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(client *firestore.Client) *UserRepository {
	return &UserRepository{
		Client: client,
	}
}


// userコレクション保存
func SaveUser(ctx context.Context, db *firestore.Client, user models.User) error {
	log.Printf("SaveUser: ユーザー '%s' を保存中...", user.FirebaseId)
	log.Printf("SaveUser: Firestoreクライアント状態: %v", db != nil)

	if db == nil {
		log.Printf("SaveUser: Firestoreクライアントがnilです")
		return fmt.Errorf("Firestoreクライアントが初期化されていません")
	}

	// ユーザー情報をFirestoreに保存
	userData := map[string]interface{}{
		"firebaseId":           user.FirebaseId,
		"githubUserName":       user.GithubUserName,
		"photoURL":             user.PhotoURL,
		"createdAt":            user.CreatedAt,
		"continuousSealRecord": user.ContinuousSealRecord,
		"maxSealRecord":        user.MaxSealRecord,
	}

	log.Printf("SaveUser: Firestoreに保存するデータ: %+v", userData)
	_, err := db.Collection("users").Doc(user.FirebaseId).Set(ctx, userData)
	if err != nil {
		log.Printf("SaveUser: Firestore保存エラー: %v", err)
		return err
	}

	log.Printf("SaveUser: ユーザー '%s' の保存に成功しました", user.FirebaseId)

	// currentMonsterサブコレクションの初期値を保存
	log.Printf("SaveUser: currentMonsterサブコレクションの初期値を作成中...")
	currentMonsterData := map[string]interface{}{
		"monsterId":                   "001", // 初期モンスターID（スライム）
		"progressContributions":       0,
		"requiredContributions":       30, // 初期モンスター（スライム）の必要コントリビューション数
		"lastContributionReflectedAt": user.CreatedAt,
		"assignedAt":                  user.CreatedAt,
	}
	
	// ドキュメントIDは自動生成（README.mdの仕様通り）
	_, _, err = db.Collection("users").Doc(user.FirebaseId).Collection("currentMonster").Add(ctx, currentMonsterData)
	if err != nil {
		log.Printf("SaveUser: currentMonster初期値の保存に失敗: %v", err)
		return fmt.Errorf("currentMonster初期値の保存に失敗: %v", err)
	}
	log.Printf("SaveUser: currentMonster初期値の保存に成功（monsterId: 001, requiredContributions: 30）")

	// sealedMonstersサブコレクションは初期状態では空なので、プレースホルダーは作成しない
	log.Printf("SaveUser: sealedMonstersサブコレクションは初期状態では空のため、プレースホルダーは作成しません")

	log.Printf("SaveUser: ユーザー '%s' とサブコレクションの初期化が完了しました", user.FirebaseId)
	return nil
}

// サブコレクションでcurrentMonster
func (r *UserRepository) SaveCurrentMonster(ctx context.Context, firebaseId string, monster models.CurrentMonster) error {
	log.Printf("SaveCurrentMonster: ユーザー '%s' のcurrentMonsterを保存中...", firebaseId)
	_, err := r.Client.Collection("users").Doc(firebaseId).Collection("currentMonster").Doc("monster").Set(ctx, map[string]interface{}{
		"monsterId":                   monster.MonsterId,
		"progressContributions":       monster.ProgressContributions,
		"lastContributionReflectedAt": monster.LastContributionReflectedAt, // contributionからもらう
		"assignedAt":                  monster.AssignedAt,
		"requiredContributions":       monster.RequiredContributions, // contributionからもらう
	})
	if err != nil {
		log.Printf("SaveCurrentMonster: currentMonster保存エラー: %v", err)
	} else {
		log.Printf("SaveCurrentMonster: currentMonster保存成功")
	}
	return err
}

// SealedMonsters
func (r *UserRepository) SaveSealedMonster(ctx context.Context, firebaseId string, sealed models.SealedMonster) error {
	log.Printf("SaveSealedMonster: ユーザー '%s' のsealedMonster '%s' を保存中...", firebaseId, sealed.MonsterId)
	_, err := r.Client.Collection("users").Doc(firebaseId).Collection("sealedMonsters").Doc(sealed.MonsterId).Set(ctx, map[string]interface{}{
		"monsterId":   sealed.MonsterId,
		"monsterName": sealed.MonsterName,
		"sealedAt":    sealed.SealedAt,
	})
	if err != nil {
		log.Printf("SaveSealedMonster: sealedMonster保存エラー: %v", err)
	} else {
		log.Printf("SaveSealedMonster: sealedMonster保存成功")
	}
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
