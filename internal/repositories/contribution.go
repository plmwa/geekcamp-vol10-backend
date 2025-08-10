package repositories

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	"cloud.google.com/go/firestore"
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
	currentMonsterDoc := docs[0]
	currentMonsterData := currentMonsterDoc.Data()
	log.Printf("currentMonsterサブコレクションのデータ: %+v", currentMonsterData)
	log.Printf("currentMonsterドキュメントID（monsterId）: %s", currentMonsterDoc.Ref.ID)
	
	// 現在のcurrentMonster情報を取得
	// MonsterIdはドキュメントIDから取得
	currentMonster := models.CurrentMonster{
		MonsterId:                   currentMonsterDoc.Ref.ID, // ドキュメントIDがmonsterIdになる
		ProgressContributions:       getInt(currentMonsterData, "progressContributions"),
		RequiredContributions:       getInt(currentMonsterData, "requiredContributions"),
		LastContributionReflectedAt: getTimestampAsTime(currentMonsterData, "lastContributionReflectedAt"),
		AssignedAt:                  getTimestampAsTime(currentMonsterData, "assignedAt"),
	}
	
	// lastContributionReflectedAtよりも最新のコントリビューションをgithubDataから取り出す
	lastReflectedTime := currentMonster.LastContributionReflectedAt
	log.Printf("currentMonster.LastContributionReflectedAt: %v", currentMonster.LastContributionReflectedAt)
	log.Printf("lastReflectedTime.IsZero(): %t", lastReflectedTime.IsZero())
	
	if lastReflectedTime.IsZero() {
		lastReflectedTime = time.Now().AddDate(0, 0, -30) // 30日前をデフォルトとする
		log.Printf("初回実行のため、30日前を基準時刻に設定: %v", lastReflectedTime)
	} else {
		log.Printf("前回反映時刻を基準に使用: %v", lastReflectedTime)
	}
	
	// GitHubデータから新しいコントリビューションを計算
	newContributions := calculateNewContributions(githubData, lastReflectedTime)
	log.Printf("新しいコントリビューション数: %d", newContributions)
	
	// 合計した値をprogressContributionsに足す
	updatedProgressContributions := currentMonster.ProgressContributions + newContributions
	now := time.Now()
	
	// デバッグ情報を詳細に出力
	log.Printf("=== コントリビューション計算結果 ===")
	log.Printf("現在のProgress: %d", currentMonster.ProgressContributions)
	log.Printf("新しいコントリビューション: %d", newContributions)
	log.Printf("更新後のProgress: %d", updatedProgressContributions)
	log.Printf("必要なコントリビューション: %d", currentMonster.RequiredContributions)
	log.Printf("封印条件: %d >= %d = %t", updatedProgressContributions, currentMonster.RequiredContributions, updatedProgressContributions >= currentMonster.RequiredContributions)
	
	// 新しいコントリビューションが0で、既に封印条件を満たしている場合は更新のみ行う
	if newContributions == 0 {
		log.Printf("新しいコントリビューションが0のため、データ更新のみ行います")
		// 既存のcurrentMonsterを更新（lastContributionReflectedAtのみ更新）
		updatedCurrentMonster := currentMonster
		updatedCurrentMonster.LastContributionReflectedAt = now
		
		err = updateCurrentMonster(ctx, db, id, currentMonsterDoc.Ref.ID, updatedCurrentMonster)
		if err != nil {
			log.Printf("currentMonster更新に失敗しました: %v", err)
			return models.CurrentMonster{}, fmt.Errorf("currentMonster更新に失敗しました")
		}
		
		return updatedCurrentMonster, nil
	}
	
	// progressContributionsがrequiredContributionsを超えた場合の処理
	if updatedProgressContributions >= currentMonster.RequiredContributions {
		log.Printf("モンスター封印完了！次のモンスターに更新します")
		
		// 現在のモンスターを封印済みに移動
		err = sealCurrentMonster(ctx, db, id, currentMonster)
		if err != nil {
			log.Printf("モンスター封印処理に失敗しました: %v", err)
			return models.CurrentMonster{}, fmt.Errorf("モンスター封印処理に失敗しました")
		}
		
		// 次のモンスターを取得して設定
		nextMonster, err := getNextMonster(ctx, db, currentMonster.MonsterId)
		if err != nil {
			log.Printf("次のモンスター取得に失敗しました: %v", err)
			return models.CurrentMonster{}, fmt.Errorf("次のモンスター取得に失敗しました")
		}
		
		// 余ったコントリビューションを次のモンスターに引き継ぎ
		carryOverContributions := updatedProgressContributions - currentMonster.RequiredContributions
		
		log.Printf("次のモンスターに引き継ぐコントリビューション数: %d", carryOverContributions)
		log.Printf("次のモンスターの必要コントリビューション数: %d (monstersコレクションから取得)", nextMonster.RequiredContributions)
		
		newCurrentMonster := models.CurrentMonster{
			MonsterId:                   nextMonster.MonsterId,
			ProgressContributions:       carryOverContributions,
			RequiredContributions:       nextMonster.RequiredContributions, // monstersコレクションから取得した値を使用
			LastContributionReflectedAt: now,
			AssignedAt:                  now,
		}
		
		// 新しいcurrentMonsterをFirestoreに保存
		err = updateCurrentMonster(ctx, db, id, currentMonsterDoc.Ref.ID, newCurrentMonster)
		if err != nil {
			log.Printf("新しいcurrentMonster保存に失敗しました: %v", err)
			return models.CurrentMonster{}, fmt.Errorf("新しいcurrentMonster保存に失敗しました")
		}
		
		// ユーザーのcontinuousSealRecordとmaxSealRecordを更新（モンスター封印時は常に更新）
		err = updateUserSealRecords(ctx, db, id, userData, now, true, githubData)
		if err != nil {
			log.Printf("ユーザーのsealRecord更新に失敗しました: %v", err)
			// sealRecord更新の失敗は処理を止めない（ログのみ出力）
		}
		
		return newCurrentMonster, nil
	} else {
		// progressContributionsを更新するだけ
		updatedCurrentMonster := currentMonster
		updatedCurrentMonster.ProgressContributions = updatedProgressContributions
		updatedCurrentMonster.LastContributionReflectedAt = now
		
		// 既存のcurrentMonsterを更新
		err = updateCurrentMonster(ctx, db, id, currentMonsterDoc.Ref.ID, updatedCurrentMonster)
		if err != nil {
			log.Printf("currentMonster更新に失敗しました: %v", err)
			return models.CurrentMonster{}, fmt.Errorf("currentMonster更新に失敗しました")
		}
		
		// コントリビューションがある場合のみユーザーのsealRecordを更新
		if newContributions > 0 {
			err = updateUserSealRecords(ctx, db, id, userData, now, true, githubData)
			if err != nil {
				log.Printf("ユーザーのsealRecord更新に失敗しました: %v", err)
				// sealRecord更新の失敗は処理を止めない（ログのみ出力）
			}
		} else {
			log.Printf("新しいコントリビューションがないため、sealRecord更新をスキップ")
		}
		
		return updatedCurrentMonster, nil
	}
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
	value, exists := data[key]
	if !exists {
		log.Printf("警告: キー '%s' が存在しません。利用可能なキー: %v", key, getKeys(data))
		return 0
	}
	
	log.Printf("キー '%s' の値: %v (型: %T)", key, value, value)
	
	if intValue, ok := value.(int); ok {
		log.Printf("int型として取得成功: %d", intValue)
		return intValue
	}
	if floatValue, ok := value.(float64); ok {
		intValue := int(floatValue)
		log.Printf("float64型から変換成功: %f -> %d", floatValue, intValue)
		return intValue
	}
	if int64Value, ok := value.(int64); ok {
		intValue := int(int64Value)
		log.Printf("int64型から変換成功: %d -> %d", int64Value, intValue)
		return intValue
	}
	
	log.Printf("警告: キー '%s' の値 %v (型: %T) を整数に変換できません", key, value, value)
	return 0
}

// ヘルパー関数: map[string]interface{}のキー一覧を取得
func getKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
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

// ヘルパー関数: map[string]interface{}からタイムスタンプをtime.Time型で取得
func getTimestampAsTime(data map[string]interface{}, key string) time.Time {
	// Firestoreのタイムスタンプは通常time.Time型で格納される
	if value, ok := data[key].(time.Time); ok {
		return value
	}
	// 文字列として格納されている場合はパースを試行
	if value, ok := data[key].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, value); err == nil {
			return parsedTime
		}
	}
	// タイムスタンプが見つからない場合は現在時刻を返す
	return time.Now()
}

// GitHubデータから指定した日時以降の新しいコントリビューションを計算
func calculateNewContributions(githubData models.GithubResponse, lastReflectedTime time.Time) int {
	totalNewContributions := 0
	processedDates := make(map[string]bool) // 処理済み日付を追跡
	
	// 詳細な日時情報のみを使用してコントリビューションを計算
	log.Printf("詳細な日時情報を使用してコントリビューションを計算します")
	log.Printf("基準時刻（lastReflectedTime）: %v", lastReflectedTime)
	
	// lastReflectedTimeの日付を取得
	lastReflectedDate := lastReflectedTime.Format("2006-01-02")
	log.Printf("基準日付: %s", lastReflectedDate)
	
	for _, repoContrib := range githubData.Data.User.ContributionsCollection.CommitContributionsByRepository {
		repoName := repoContrib.Repository.Owner.Login + "/" + repoContrib.Repository.Name
		log.Printf("リポジトリ '%s' のコントリビューションを確認中", repoName)
		
		for _, contribution := range repoContrib.Contributions.Nodes {
			occurredTime, err := time.Parse(time.RFC3339, contribution.OccurredAt)
			if err != nil {
				log.Printf("日時のパースに失敗しました: %s, エラー: %v", contribution.OccurredAt, err)
				continue
			}
			
			// コントリビューションの日付を取得
			contributionDate := occurredTime.Format("2006-01-02")
			
			// 時刻の詳細な比較ログ
			timeDiff := occurredTime.Sub(lastReflectedTime)
			isAfter := occurredTime.After(lastReflectedTime)
			
			log.Printf("=== コントリビューション時刻比較 ===")
			log.Printf("コントリビューション時刻: %v (日付: %s)", occurredTime, contributionDate)
			log.Printf("基準時刻（lastReflected）: %v (日付: %s)", lastReflectedTime, lastReflectedDate)
			log.Printf("時刻差（秒）: %v", timeDiff.Seconds())
			log.Printf("After判定: %t", isAfter)
			
			// 日付ベースの重複チェック
			dateKey := fmt.Sprintf("%s:%s", repoName, contributionDate)
			if processedDates[dateKey] {
				log.Printf("同日の重複コントリビューションをスキップ: %s", dateKey)
				continue
			}
			
			// 基準日付より後の日付のコントリビューションのみを計算
			if contributionDate > lastReflectedDate {
				totalNewContributions += contribution.CommitCount
				processedDates[dateKey] = true
				log.Printf("新しいコントリビューション追加: リポジトリ=%s, 日付=%s, コミット数=%d", 
					repoName, contributionDate, contribution.CommitCount)
			} else {
				log.Printf("既に反映済みの日付のコントリビューションをスキップ: リポジトリ=%s, 日付=%s, コミット数=%d", 
					repoName, contributionDate, contribution.CommitCount)
			}
		}
	}
	
	return totalNewContributions
}
// 現在のモンスターを封印済みに移動
func sealCurrentMonster(ctx context.Context, db *firestore.Client, userID string, monster models.CurrentMonster) error {
	// monstersコレクションからモンスター名を取得
	monsterDoc, err := db.Collection("monsters").Doc(monster.MonsterId).Get(ctx)
	if err != nil {
		return fmt.Errorf("モンスター情報の取得に失敗しました: %v", err)
	}
	
	monsterData := monsterDoc.Data()
	monsterName := getString(monsterData, "name")
	
		// モンスター名が取得できない場合はデフォルト名を使用
		if monsterName == "" {
			monsterName = "モンスター" + monster.MonsterId
			log.Printf("警告: モンスターID '%s' の名前が取得できませんでした。デフォルト名を使用します: %s", monster.MonsterId, monsterName)
		}	// sealedMonstersサブコレクションに追加
	sealedData := map[string]interface{}{
		"monsterId":   monster.MonsterId,
		"monsterName": monsterName,
		"sealedAt":    time.Now().Format(time.RFC3339),
	}
	
	log.Printf("封印済みモンスターデータ: %+v", sealedData)
	
	_, _, err = db.Collection("users").Doc(userID).Collection("sealedMonsters").Add(ctx, sealedData)
	if err != nil {
		return fmt.Errorf("封印済みモンスターの保存に失敗しました: %v", err)
	}
	
	log.Printf("モンスター '%s' (%s) をユーザー '%s' の封印済みモンスターに追加しました", monsterName, monster.MonsterId, userID)
	return nil
}

// 次のモンスター情報を取得
func getNextMonster(ctx context.Context, db *firestore.Client, currentMonsterID string) (models.CurrentMonster, error) {
	// currentMonsterIDから数値部分を抽出して+1
	// 例: "001" -> "002"
	currentIDNum, err := strconv.Atoi(currentMonsterID)
	if err != nil {
		return models.CurrentMonster{}, fmt.Errorf("モンスターIDの変換に失敗しました: %v", err)
	}
	
	nextIDNum := currentIDNum + 1
	nextMonsterID := fmt.Sprintf("%03d", nextIDNum) // 3桁0埋め
	
	// monstersコレクションから次のモンスター情報を取得
	monsterDoc, err := db.Collection("monsters").Doc(nextMonsterID).Get(ctx)
	if err != nil {
		// 次のモンスターが存在しない場合は、最初のモンスターに戻る
		log.Printf("モンスターID '%s' が見つかりません。最初のモンスターに戻ります", nextMonsterID)
		nextMonsterID = "001"
		monsterDoc, err = db.Collection("monsters").Doc(nextMonsterID).Get(ctx)
		if err != nil {
			return models.CurrentMonster{}, fmt.Errorf("デフォルトモンスターの取得に失敗しました: %v", err)
		}
	}
	
	monsterData := monsterDoc.Data()
	log.Printf("Monstersコレクションから取得したデータ (ID=%s): %+v", nextMonsterID, monsterData)
	
	requiredContributions := getInt(monsterData, "requiredContributions")
	
	// requiredContributionsが0の場合は警告を出す
	if requiredContributions == 0 {
		log.Printf("警告: モンスターID '%s' のrequiredContributionsが0です。データを確認してください", nextMonsterID)
	}
	
	log.Printf("次のモンスター情報: ID=%s, 必要コントリビューション数=%d", nextMonsterID, requiredContributions)
	
	// 新しいモンスターの lastContributionReflectedAt を今日の終了時刻に設定
	now := time.Now()
	endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	
	return models.CurrentMonster{
		MonsterId:             nextMonsterID,
		RequiredContributions: requiredContributions,
		ProgressContributions: 0, // 新しいモンスターは0からスタート
		AssignedAt:           now,
		LastContributionReflectedAt: endOfToday,
	}, nil
}

// currentMonsterを更新
func updateCurrentMonster(ctx context.Context, db *firestore.Client, userID, docID string, monster models.CurrentMonster) error {
	updateData := map[string]interface{}{
		"progressContributions":       monster.ProgressContributions,
		"requiredContributions":       monster.RequiredContributions,
		"lastContributionReflectedAt": monster.LastContributionReflectedAt,
		"assignedAt":                  monster.AssignedAt,
	}
	
	// 新しいモンスターの場合、ドキュメントIDも変更する必要がある
	if docID != monster.MonsterId {
		log.Printf("新しいモンスターのため、ドキュメントを置き換えます: %s -> %s", docID, monster.MonsterId)
		
		// 古いドキュメントを削除
		_, err := db.Collection("users").Doc(userID).Collection("currentMonster").Doc(docID).Delete(ctx)
		if err != nil {
			log.Printf("古いcurrentMonsterドキュメントの削除に失敗: %v", err)
		}
		
		// 新しいドキュメントを作成
		_, err = db.Collection("users").Doc(userID).Collection("currentMonster").Doc(monster.MonsterId).Set(ctx, updateData)
		if err != nil {
			return fmt.Errorf("新しいcurrentMonsterの作成に失敗しました: %v", err)
		}
	} else {
		// 同じモンスターの場合は既存ドキュメントを更新
		_, err := db.Collection("users").Doc(userID).Collection("currentMonster").Doc(docID).Set(ctx, updateData)
		if err != nil {
			return fmt.Errorf("currentMonsterの更新に失敗しました: %v", err)
		}
	}
	
	return nil
}

// ユーザーのcontinuousSealRecordとmaxSealRecordを更新
func updateUserSealRecords(ctx context.Context, db *firestore.Client, userID string, userData map[string]interface{}, now time.Time, hasNewContributions bool, githubData models.GithubResponse) error {
	log.Printf("ユーザー '%s' のsealRecord更新を開始 (新しいコントリビューション: %t)", userID, hasNewContributions)
	
	// 新しいコントリビューションがない場合は更新しない
	if !hasNewContributions {
		log.Printf("新しいコントリビューションがないため、sealRecord更新をスキップ")
		return nil
	}
	
	// 現在のcontinuousSealRecordとmaxSealRecordを取得
	currentContinuous := getInt(userData, "continuousSealRecord")
	currentMax := getInt(userData, "maxSealRecord")
	
	// lastContributionReflectedAtを取得（ユーザーコレクションから）
	lastContributionTime := getTimestampAsTime(userData, "lastContributionReflectedAt")
	
	// 初回の場合はlastContributionTimeが0値になるので、現在時刻を設定
	if lastContributionTime.IsZero() {
		lastContributionTime = now.AddDate(0, 0, -2) // 2日前に設定して確実にリセット
		log.Printf("初回実行のため、lastContributionTimeを2日前に設定: %v", lastContributionTime)
	}
	
	// GitHubデータから最新のコントリビューション時刻を取得
	latestContributionTime := getLatestContributionTime(githubData, lastContributionTime)
	log.Printf("最新のコントリビューション時刻: %v", latestContributionTime)
	log.Printf("前回のコントリビューション反映時刻: %v", lastContributionTime)
	
	// 最新のコントリビューション時刻と前回反映時刻の差を計算
	var timeDiff time.Duration
	var isWithinOneDay bool
	
	if !latestContributionTime.IsZero() {
		timeDiff = latestContributionTime.Sub(lastContributionTime)
		isWithinOneDay = timeDiff <= 24*time.Hour && timeDiff > 0
		log.Printf("コントリビューション時刻差: %v", timeDiff)
	} else {
		// GitHubデータから時刻が取得できない場合は、現在時刻を基準にする
		timeDiff = now.Sub(lastContributionTime)
		isWithinOneDay = timeDiff <= 24*time.Hour
		log.Printf("GitHubデータから時刻取得不可、現在時刻で判定: %v", timeDiff)
	}
	
	log.Printf("1日以内: %t", isWithinOneDay)
	log.Printf("現在のcontinuousSealRecord: %d", currentContinuous)
	log.Printf("現在のmaxSealRecord: %d", currentMax)
	
	var newContinuous int
	var newMax int
	
	if isWithinOneDay {
		// 1日以内の場合、continuousSealRecordを+1
		newContinuous = currentContinuous + 1
		log.Printf("1日以内のため、continuousSealRecordを+1: %d -> %d", currentContinuous, newContinuous)
	} else {
		// 1日以上経過している場合、continuousSealRecordを1にリセット
		newContinuous = 1
		log.Printf("1日以上経過のため、continuousSealRecordを1にリセット: %d -> %d", currentContinuous, newContinuous)
	}
	
	// maxSealRecordを更新（新しいcontinuousが最大値を超えた場合）
	if newContinuous > currentMax {
		newMax = newContinuous
		log.Printf("新記録！maxSealRecordを更新: %d -> %d", currentMax, newMax)
	} else {
		newMax = currentMax
		log.Printf("maxSealRecordは変更なし: %d", newMax)
	}
	
	// lastContributionReflectedAtを今日の終了時刻（23:59:59）に更新
	// これにより、同じ日のコントリビューションの重複処理を防ぐ
	endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	
	// Firestoreを更新
	_, err := db.Collection("users").Doc(userID).Update(ctx, []firestore.Update{
		{Path: "continuousSealRecord", Value: newContinuous},
		{Path: "maxSealRecord", Value: newMax},
		{Path: "lastContributionReflectedAt", Value: endOfToday},
	})
	
	if err != nil {
		log.Printf("ユーザーsealRecord更新エラー: %v", err)
		return fmt.Errorf("ユーザーsealRecord更新に失敗: %v", err)
	}
	
	log.Printf("ユーザー '%s' のsealRecord更新完了: continuous=%d, max=%d", userID, newContinuous, newMax)
	return nil
}

// GitHubデータから最新のコントリビューション時刻を取得
func getLatestContributionTime(githubData models.GithubResponse, lastReflectedTime time.Time) time.Time {
	var latestTime time.Time
	
	for _, repoContrib := range githubData.Data.User.ContributionsCollection.CommitContributionsByRepository {
		for _, contribution := range repoContrib.Contributions.Nodes {
			occurredTime, err := time.Parse(time.RFC3339, contribution.OccurredAt)
			if err != nil {
				log.Printf("コントリビューション時刻のパースに失敗: %s, エラー: %v", contribution.OccurredAt, err)
				continue
			}
			
			// lastReflectedTimeより後のコントリビューションのみを対象とする
			if occurredTime.After(lastReflectedTime) {
				if latestTime.IsZero() || occurredTime.After(latestTime) {
					latestTime = occurredTime
					log.Printf("最新コントリビューション時刻を更新: %v", latestTime)
				}
			}
		}
	}
	
	return latestTime
}