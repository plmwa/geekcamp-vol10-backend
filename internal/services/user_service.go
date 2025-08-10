package services

import (
	"context"
	"fmt"
	"log"
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

func (s *UserService) CreateUser(ctx context.Context, firebaseId, githubUserName, photoURL string) (map[string]interface{}, error) {
	log.Printf("CreateUser: 新しいユーザーを作成中 - FirebaseId: %s", firebaseId)
	
	user := models.User{
		FirebaseId:           firebaseId,
		GithubUserName:       githubUserName,
		PhotoURL:             photoURL,
		CreatedAt:            time.Now().Unix(),
		ContinuousSealRecord: 0,
		MaxSealRecord:        0,
	}

	if err := s.UserRepo.SaveUser(ctx, user); err != nil {
		log.Printf("CreateUser: ユーザー保存に失敗: %v", err)
		return nil, err
	}

	log.Printf("CreateUser: ユーザー作成完了")
	return s.UserRepo.GetUser(ctx, firebaseId)
}


func GetUserByIDService(ctx context.Context, id string) (*models.User, error) {
	log.Printf("GetUserByIDService: ユーザーID '%s' の取得を開始", id)

	client := database.GetFirestoreClient()
	if client == nil {
		log.Printf("GetUserByIDService: Firestoreクライアントの初期化に失敗")
		return nil, fmt.Errorf("Firestoreクライアントが初期化されていません")
	}

	// ユーザー本体取得
	log.Printf("GetUserByIDService: ユーザー本体の取得を試行中...")
	main, err := client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		log.Printf("GetUserByIDService: ユーザー本体の取得に失敗: %v", err)
		return nil, err
	}

	// 実際のFirestoreデータを確認
	data := main.Data()
	log.Printf("GetUserByIDService: Firestoreの生データ: %+v", data)
	for key, value := range data {
		log.Printf("フィールド '%s': 値=%v, 型=%T", key, value, value)
	}

	var user models.User
	if err := main.DataTo(&user); err != nil {
		log.Printf("GetUserByIDService: ユーザーデータのマッピングに失敗: %v", err)
		return nil, err
	}
	
	log.Printf("GetUserByIDService: ユーザー本体を正常に取得: %+v", user)

	// currentMonsterの取得
	log.Printf("GetUserByIDService: currentMonsterの取得を試行中...")
	sub, err := client.Collection("users").Doc(id).Collection("currentMonster").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("GetUserByIDService: currentMonsterの取得に失敗: %v", err)
		return nil, err
	}

	if len(sub) > 0 {
		var cm models.CurrentMonster
		if err := sub[0].DataTo(&cm); err != nil {
			log.Printf("GetUserByIDService: currentMonsterデータのマッピングに失敗: %v", err)
			return nil, err
		}
		user.CurrentMonster = &cm
		log.Printf("GetUserByIDService: currentMonsterを正常に取得: %+v", cm)
	} else {
		user.CurrentMonster = nil
		log.Printf("GetUserByIDService: currentMonsterが見つかりませんでした")
	}

	// sealedMonstersの取得
	log.Printf("GetUserByIDService: sealedMonstersの取得を試行中...")
	
	// まず、コレクション参照を明示的に作成
	sealedMonstersRef := client.Collection("users").Doc(id).Collection("sealedMonsters")
	log.Printf("GetUserByIDService: sealedMonstersコレクション参照: %v", sealedMonstersRef.Path)
	
	// すべてのドキュメントを取得
	sealedDocs, err := sealedMonstersRef.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("GetUserByIDService: sealedMonstersの取得に失敗: %v", err)
		return nil, err
	}

	log.Printf("GetUserByIDService: Firestoreから取得したsealedMonstersドキュメント数: %d", len(sealedDocs))
	
	var sealedMonsters []models.SealedMonster
	for i, doc := range sealedDocs {
		log.Printf("GetUserByIDService: sealedMonster[%d] ドキュメントID: %s", i, doc.Ref.ID)
		data := doc.Data()
		log.Printf("GetUserByIDService: sealedMonster[%d] 生データ: %+v", i, data)
		
		// 手動でSealedMonsterを構築
		sm := models.SealedMonster{}
		
		// MonsterId
		if monsterId, ok := data["monsterId"].(string); ok {
			sm.MonsterId = monsterId
		}
		
		// MonsterName
		if monsterName, ok := data["monsterName"].(string); ok {
			sm.MonsterName = monsterName
		}
		
		// SealedAt - 複数の型に対応
		if sealedAtValue, exists := data["sealedAt"]; exists {
			switch v := sealedAtValue.(type) {
			case time.Time:
				sm.SealedAt = v
			case string:
				if parsedTime, err := time.Parse(time.RFC3339, v); err == nil {
					sm.SealedAt = parsedTime
				} else {
					log.Printf("GetUserByIDService: sealedMonster[%d] SealedAtの文字列パースに失敗: %v", i, err)
					sm.SealedAt = time.Time{} // ゼロ値を設定
				}
			default:
				log.Printf("GetUserByIDService: sealedMonster[%d] SealedAtの型が不明: %T", i, v)
				sm.SealedAt = time.Time{} // ゼロ値を設定
			}
		}
		
		log.Printf("GetUserByIDService: sealedMonster[%d] マッピング完了: %+v", i, sm)
		sealedMonsters = append(sealedMonsters, sm)
	}
	user.SealedMonsters = sealedMonsters
	log.Printf("GetUserByIDService: sealedMonstersを正常に取得: %d件", len(sealedMonsters))

	log.Printf("GetUserByIDService: 最終的なユーザー情報:")
	log.Printf("  - FirebaseId: %s", user.FirebaseId)
	log.Printf("  - GithubUserName: %s", user.GithubUserName)
	log.Printf("  - CurrentMonster: %+v", user.CurrentMonster)
	log.Printf("  - SealedMonsters数: %d", len(user.SealedMonsters))
	for i, sm := range user.SealedMonsters {
		log.Printf("    [%d] %+v", i, sm)
	}
	return &user, nil
}
