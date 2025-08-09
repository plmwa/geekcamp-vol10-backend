package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"geekcamp-vol10-backend/internal/config"
)

var (
	// グローバルなFirestoreクライアント
	FirestoreClient *firestore.Client
	// グローバルなFirebaseアプリ
	FirebaseApp *firebase.App
)

// InitializeFirestore はFirestoreクライアントを初期化します
func InitializeFirestore(ctx context.Context, cfg *config.Config) error {
	var err error

	// エミュレータモードかどうかを判定
	if cfg.FirestoreEmulatorHost != "" {
		log.Println("Firestore エミュレータモードで初期化します")
		
		// エミュレータ用の設定
		if cfg.GCloudProject == "" {
			log.Fatal("エミュレータ利用時は環境変数 'GCLOUD_PROJECT' が必要です")
		}

		conf := &firebase.Config{
			ProjectID: cfg.GCloudProject,
		}
		FirebaseApp, err = firebase.NewApp(ctx, conf)
		if err != nil {
			log.Printf("Firebase App の初期化に失敗しました: %v", err)
			return err
		}

	} else {
		log.Println("本番環境のFirestoreで初期化します")
		
		// 本番環境用の設定
		if cfg.FirebaseCredentials == "" {
			log.Fatal("本番環境利用時は環境変数 'CREDENTIALS' が必要です")
		}

		opt := option.WithCredentialsFile(cfg.FirebaseCredentials)
		FirebaseApp, err = firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Printf("Firebase App の初期化に失敗しました: %v", err)
			return err
		}
	}

	// Firestoreクライアントを取得
	FirestoreClient, err = FirebaseApp.Firestore(ctx)
	if err != nil {
		log.Printf("Firestore クライアントの取得に失敗しました: %v", err)
		return err
	}

	log.Println("Firestore の初期化が完了しました")
	return nil
}

// CloseFirestore はFirestoreクライアントを閉じます
func CloseFirestore() error {
	if FirestoreClient != nil {
		return FirestoreClient.Close()
	}
	return nil
}

// GetFirestoreClient はFirestoreクライアントを返します
func GetFirestoreClient() *firestore.Client {
	return FirestoreClient
}

// GetFirebaseApp はFirebaseアプリを返します
func GetFirebaseApp() *firebase.App {
	return FirebaseApp
}
