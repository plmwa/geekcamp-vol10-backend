package repositories

import (
	"context"
	"fmt"
	"log"
	"time"
	"geekcamp-vol10-backend/internal/models"
	"geekcamp-vol10-backend/pkg/database"
)


func SaveContribution(id string, githubData models.GithubResponse) (models.CurrentMonster, error) {
	// githubDataを用いながらDBに保存
	// DBのUsersコレクションの:idの人のcurrentMonsterを返す
	ctx := context.Background()
	db := database.GetFirestoreClient()
	if db == nil {
		log.Printf("Firestoreクライアントが初期化されていません")
		return models.CurrentMonster{}, fmt.Errorf("Firestoreクライアントが初期化されていません")
	}

	// dbからcurrentMonsterのprogressContributionsとrequiredContributionsとlastContributionReflectedAtを取得
	// lastContributionReflectedAtよりも最新のコントリビューションをgithubDataから取り出す
	// 取り出したgithubDataのContributionCountを合計する
	// 合計した値をprogressContributionsに足す
	// progressContributionsがrequiredContributionsを超えた場合、currentMonsterを更新する
	// 更新するのはcurrentMonsterのmonsterIdに1を足したmonsterIdを持つMonsterscollectionのドキュメントである


	// currentMonsterはサブコレクション
	// 最初にusersドキュメントを取得
	userDoc, err := db.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		log.Printf("ユーザーID '%s' のドキュメント取得に失敗しました: %v", id, err)
		return models.CurrentMonster{}, fmt.Errorf("ユーザーが見つかりません")
	}

	// デバッグ: ドキュメント全体の構造を確認
	userData := userDoc.Data()
	log.Printf("ユーザーID '%s' のドキュメント全体: %+v", id, userData)
	
	// currentMonsterはサブコレクション
	currentMonsterCollection := db.Collection("users").Doc(id).Collection("currentMonster")
	docs, err := currentMonsterCollection.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("currentMonsterサブコレクションの取得に失敗しました: %v", err)
		return models.CurrentMonster{}, fmt.Errorf("currentMonsterの取得に失敗しました")
	}
	
	if len(docs) == 0 {
		log.Printf("ユーザーID '%s' のcurrentMonsterが見つかりません", id)
		return models.CurrentMonster{}, fmt.Errorf("currentMonsterが見つかりません")
	}
	
	// 最初のドキュメントを使用（通常は1つのみ存在）
	currentMonsterData := docs[0].Data()
	log.Printf("currentMonsterサブコレクションのデータ: %+v", currentMonsterData)
	
	currentMonster := models.CurrentMonster{
		MonsterID:                   getString(currentMonsterData, "monsterId"),
		ProgressContributions:       getInt(currentMonsterData, "progressContributions"),
		RequiredContributions:       getInt(currentMonsterData, "requiredContributions"),
		LastContributionReflectedAt: getTimestamp(currentMonsterData, "lastContributionReflectedAt"),
		AssignedAt:                  getTimestamp(currentMonsterData, "assignedAt"),
	}
	
	return currentMonster, nil
}

func GetGitHubUserNameByID(ctx context.Context, id string) (string, error) {
	// データベースクライアントを取得
	db := database.GetFirestoreClient()
	if db == nil {
		return "", fmt.Errorf("Firestoreクライアントが初期化されていません")
	}
	
	// DBからユーザー名を取得する処理を実装
	doc, err := db.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		log.Printf("ユーザーID '%s' のドキュメント取得に失敗しました: %v", id, err)
		return "", fmt.Errorf("ユーザーが見つかりません")
	}

	// ドキュメントから "githubUserName" フィールドの値を取得します。
	// 型がstringであることを期待しています。
	githubUserName, ok := doc.Data()["githubUserName"].(string)
	if !ok {
		log.Printf("ユーザーID '%s' の 'githubUserName' フィールドがstring型ではありません。", id)
		return "", fmt.Errorf("データ形式が正しくありません")
	}

	return githubUserName, nil
}

// ヘルパー関数: map[string]interface{}から文字列を安全に取得
func getString(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

// ヘルパー関数: map[string]interface{}から整数を安全に取得
func getInt(data map[string]interface{}, key string) int {
	if value, ok := data[key].(int); ok {
		return value
	}
	if value, ok := data[key].(float64); ok {
		return int(value)
	}
	return 0
}

// ヘルパー関数: map[string]interface{}からタイムスタンプを安全に取得
func getTimestamp(data map[string]interface{}, key string) string {
	// Firestoreのタイムスタンプは通常time.Time型で格納される
	if value, ok := data[key].(time.Time); ok {
		return value.Format(time.RFC3339)
	}
	// 文字列として格納されている場合
	if value, ok := data[key].(string); ok {
		return value
	}
	// タイムスタンプが見つからない場合は現在時刻を返す
	return time.Now().Format(time.RFC3339)
}