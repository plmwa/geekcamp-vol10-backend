package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"github.com/joho/godotenv"
)

// firebaseAppをグローバル変数として保持
var firebaseApp *firebase.App

func init() {
	ctx := context.Background()
	var err error

	// --- エミュレータ利用判定 ---
	// FIREBASE_AUTH_EMULATOR_HOST 環境変数が設定されている場合、エミュレータに接続します。
	// 例: FIREBASE_AUTH_EMULATOR_HOST="localhost:9099"
	godotenv.Load()
	if os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" {
		// エミュレータモード
		log.Println("エミュレータモードで初期化します。")

		// エミュレータ利用時はサービスアカウントキーは不要ですが、プロジェクトIDが必要です。
		// GCLOUD_PROJECT 環境変数から取得します。
		// 例: GCLOUD_PROJECT="your-project-id"
		projectID := os.Getenv("GCLOUD_PROJECT")
		if projectID == "" {
			log.Fatalf("エミュレータ利用時は環境変数 'GCLOUD_PROJECT' が必要です。")
		}

		conf := &firebase.Config{
			ProjectID: projectID,
		}
		firebaseApp, err = firebase.NewApp(ctx, conf)

	} else {
		// 本番環境モード
		log.Println("本番環境のFirebaseで初期化します。")

		// 従来通り、サービスアカウントキーのJSONファイルへのパスを環境変数から取得します。
		// 例: CREDENTIALS="./path/to/your/serviceAccountKey.json"
		credentialsPath := os.Getenv("CREDENTIALS")
		if credentialsPath == "" {
			log.Fatalf("本番環境利用時は環境変数 'CREDENTIALS' が設定されていません。")
		}
		opt := option.WithCredentialsFile(credentialsPath)
		firebaseApp, err = firebase.NewApp(ctx, nil, opt)
	}

	if err != nil {
		log.Fatalf("Firebaseの初期化に失敗しました: %v", err)
	}
}

// AuthMiddleware はGinのミドルウェアとして機能する関数を返す
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Firebase Authクライアントを取得
		authClient, err := firebaseApp.Auth(context.Background())
		if err != nil {
			log.Printf("Authクライアントの取得に失敗しました: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部エラー"})
			return
		}

		// リクエストヘッダーから "Authorization" を取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorizationヘッダーが必要です"})
			return
		}

		// "Bearer "の部分を削除してIDトークンを抽出
		idToken := strings.Replace(authHeader, "Bearer ", "", 1)
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "トークンが空です"})
			return
		}

		// IDトークンを検証
		token, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			log.Printf("IDトークンの検証に失敗しました: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
			return
		}

		// (オプション) 検証したユーザーIDをContextに保存して、後続のハンドラで利用できるようにします
		c.Set("firebase_uid", token.UID)

		// 検証に成功した場合、次の処理（ハンドラ）へ進みます
		c.Next()
	}
}
